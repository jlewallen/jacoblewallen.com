package main

import (
	"flag"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"encoding/json"
	"encoding/xml"

	"image/jpeg"

	texttemplate "text/template"

	"github.com/muesli/smartcrop"
	"github.com/muesli/smartcrop/nfnt"
)

var (
	verbose = false
)

type AlbumFile struct {
	Name         string
	OriginalPath string
	CreatedAt    time.Time
	PhotoPath    string
	Xmp          *XmpFile
	Original     *ImageMeta
	Large        *ImageMeta
}

var (
	ThumbnailSizes = []uint{200}
)

type Album struct {
	Config *AlbumConfig
	Files  []*AlbumFile
	Date   time.Time
}

type CachedImage struct {
	Path  string
	Image image.Image
}

type Cache struct {
	XmpsByBaseName map[string]string
	AllAlbums      []*Album
	AlbumsByTag    map[string]*Album
	Images         CachedImage
}

func (c *Cache) Load(path string) (image.Image, error) {
	if c.Images.Path == path {
		return c.Images.Image, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	i, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	c.Images.Path = path
	c.Images.Image = i

	return c.Images.Image, nil
}

func (c *Cache) Fill(o *Configuration) error {
	c.AllAlbums = make([]*Album, 0)
	c.AlbumsByTag = make(map[string]*Album)

	for _, albumCfg := range o.Albums {
		album := &Album{
			Config: albumCfg,
			Files:  make([]*AlbumFile, 0),
			Date:   time.Now(),
		}

		c.AllAlbums = append(c.AllAlbums, album)
		c.AlbumsByTag[albumCfg.Tag] = album
	}

	c.XmpsByBaseName = make(map[string]string)

	if err := c.AddExtensions(o, ".arw.xmp"); err != nil {
		return err
	}
	if err := c.AddExtensions(o, ".jpg.xmp"); err != nil {
		return err
	}
	return nil
}

func (c *Cache) AddExtensions(o *Configuration, extension string) error {
	return filepath.Walk(o.Library.Path, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}

		if info.Mode().IsRegular() {
			base := removeAllExtensions(info.Name())
			if strings.HasSuffix(strings.ToLower(info.Name()), strings.ToLower(extension)) {
				if existing, ok := c.XmpsByBaseName[base]; ok {
					if verbose {
						log.Printf("already have xmp for base: %s (%s)", info.Name(), existing)
					}
				} else {
					c.XmpsByBaseName[base] = path
				}
			}
		}
		return nil
	})
}

func (c *Cache) FindXmp(name string) (string, error) {
	base := removeAllExtensions(name)
	if c.XmpsByBaseName[base] != "" {
		return c.XmpsByBaseName[base], nil
	}
	return "", nil
}

type Subjects struct {
	Subjects []string `xml:"Bag>li"`
}

type HierarchicalSubjects struct {
	Subjects []string `xml:"Bag>li"`
}

type RdfRoot struct {
	Description RdfDescription `xml:"Description"`
}

type RdfDescription struct {
	Rating           int64  `xml:"Rating,attr"`
	DateTimeOriginal string `xml:"DateTimeOriginal,attr"`
	DerivedFrom      string `xml:"DerivedFrom,attr"`

	Subjects             Subjects             `xml:"subject"`
	HierarchicalSubjects HierarchicalSubjects `xml:"hierarchicalSubject"`
	History              []DarkTableHistory   `xml:"history>Seq>li"`
}

type DarkTableHistory struct {
	Number         int64  `xml:"num,attr"`
	Operation      string `xml:"operation,attr"`
	Enabled        int64  `xml:"enabled,attr"`
	ModuleVersion  int64  `xml:"modversion,attr"`
	Params         string `xml:"params,attr"`
	MultiName      string `xml:"multi_name,attr"`
	MultiPriority  int64  `xml:"multi_priority,attr"`
	IopOrder       string `xml:"iop_order,attr"`
	BlendOpVersion int64  `xml:"blendop_version"`
	BlendOpParams  string `xml:"blendop_params"`
}

type XmpFile struct {
	Rdf RdfRoot `xml:"RDF"`
}

func openXmp(path string) (*XmpFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	data, _ := ioutil.ReadAll(file)

	xmpFile := &XmpFile{}
	err = xml.Unmarshal(data, xmpFile)
	if err != nil {
		return nil, err
	}

	return xmpFile, nil
}

type ImageMeta struct {
	Path string
	Dx   uint
	Dy   uint
}

func getImageMeta(path string) (im *ImageMeta, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	image, _, err := image.DecodeConfig(file)
	if err != nil {
		return nil, err
	}

	im = &ImageMeta{
		Path: path,
		Dx:   uint(image.Width),
		Dy:   uint(image.Height),
	}

	return
}

type Generator struct {
	Cache      *Cache
	AlbumsRoot string
}

func NewGenerator(configPath, albumsRoot string) (g *Generator, err error) {
	g = &Generator{
		Cache:      &Cache{},
		AlbumsRoot: albumsRoot,
	}

	cfg, err := g.OpenConfiguration(configPath)
	if err != nil {
		return nil, err
	}

	// This scans the library and looks for side car files, then opens
	// those side car files and tries to find photos that belong in
	// one of our albums.
	err = g.Cache.Fill(cfg)
	if err != nil {
		return nil, err
	}

	// This looks at all the sources, which are generally exported
	// images and tries to see if one of them has a corresponding XMP
	// that we found earlier.
	for _, source := range cfg.Sources {
		err = g.IncludeDirectory(source)
		if err != nil {
			return nil, err
		}
	}

	return
}

func (g *Generator) IncludeImage(path string) error {
	originalMeta, err := getImageMeta(path)
	if err != nil {
		return err
	}

	// We rely on images having globally unique file names, which is
	// risky but would be annoying to deal with otherwise. This takes
	// the base name of the exported image and tries to find it's XMP.
	name := filepath.Base(path)
	xmpPath, err := g.Cache.FindXmp(name)
	if err != nil {
		return err
	}
	if len(xmpPath) == 0 {
		if verbose {
			log.Printf("missing xmp: %v (%v)", path, name)
		}
		return nil
	}

	xmp, err := openXmp(xmpPath)
	if err != nil {
		return err
	}

	if false {
		log.Printf("include: %v %v %v %v", path, originalMeta, xmpPath, xmp.Rdf.Description.HierarchicalSubjects.Subjects)
	}

	createdAt := time.Time{}
	if xmp.Rdf.Description.DateTimeOriginal != "" {
		dto, err := time.Parse("2006:01:02 15:04:05", xmp.Rdf.Description.DateTimeOriginal)
		if err != nil {
			return err
		}

		createdAt = dto
	}

	for _, hs := range xmp.Rdf.Description.HierarchicalSubjects.Subjects {
		album := g.Cache.AlbumsByTag[hs]
		if album != nil {
			af := &AlbumFile{
				OriginalPath: path,
				PhotoPath:    name,
				CreatedAt:    createdAt,
				Name:         name,
				Original:     originalMeta,
				Large:        CalculateNewSizes(g.AlbumsRoot, originalMeta, 1600, 1200, "large"),
				Xmp:          xmp,
			}

			if verbose {
				log.Printf("adding to album '%s' (%s) : %v", album.Config.Title, hs, af.PhotoPath)
			}

			album.Files = append(album.Files, af)
			album.Date = af.CreatedAt
			break
		}
	}

	return nil
}

func (g *Generator) IncludeDirectory(path string) error {
	err := filepath.Walk(path, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}

		if info.Mode().IsRegular() {
			fileExt := strings.ToLower(filepath.Ext(info.Name()))

			if !strings.HasSuffix(info.Name(), ".haar.jpg") && !strings.HasSuffix(info.Name(), ".cnn.jpg") {
				if fileExt == ".jpg" {
					err := g.IncludeImage(path)
					if err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

type Configuration struct {
	Sources []string       `json:"sources"`
	Library *LibraryConfig `json:"library"`
	Albums  []*AlbumConfig `json:"albums"`
}

type LibraryConfig struct {
	Path string `json:"path"`
}

type AlbumConfig struct {
	Title    string `json:"title"`
	PathName string `json:"path"`
	Tag      string `json:"tag"`
}

func (g *Generator) OpenConfiguration(path string) (*Configuration, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	data, _ := ioutil.ReadAll(file)

	cfg := &Configuration{}
	err = json.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (g *Generator) SaveJpeg(image image.Image, path string) error {
	err := os.MkdirAll(filepath.Dir(path), 755)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	options := jpeg.Options{
		Quality: 80,
	}

	err = jpeg.Encode(file, image, &options)
	if err != nil {
		return err
	}

	return nil
}

func ResizedPath(albumRoot, original string, size string) string {
	name := filepath.Base(original)
	return filepath.Join(albumRoot, size, name)
}

func ThumbnailPath(albumRoot, original string, size uint) string {
	return ResizedPath(albumRoot, original, fmt.Sprintf("%d", size))
}

func (g *Generator) HasAllThumbnails(albumRoot, original string, sizes []uint) bool {
	for _, size := range sizes {
		tp := ThumbnailPath(albumRoot, original, size)
		if _, err := os.Stat(tp); os.IsNotExist(err) {
			return false
		}
	}

	return true
}

func (g *Generator) Thumbnails(albumRoot, original string, sizes []uint) error {
	if g.HasAllThumbnails(albumRoot, original, sizes) {
		return nil
	}

	originalImage, err := g.Cache.Load(original)
	if err != nil {
		return err
	}

	for _, size := range sizes {
		tp := ThumbnailPath(albumRoot, original, size)
		if _, err := os.Stat(tp); os.IsNotExist(err) {
			err := g.Thumbnail(originalImage, size, tp)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *Generator) Thumbnail(original image.Image, size uint, path string) error {
	resizer := nfnt.NewDefaultResizer()
	analyzer := smartcrop.NewAnalyzer(resizer)
	topCrop, _ := analyzer.FindBestCrop(original, int(size), int(size))

	log.Printf("generating thumbnail %s %dpx", path, size)

	type SubImager interface {
		SubImage(r image.Rectangle) image.Image
	}
	cropped := original.(SubImager).SubImage(topCrop)
	thumb := resizer.Resize(cropped, size, size)

	err := g.SaveJpeg(thumb, path)
	if err != nil {
		return err
	}

	return nil
}

func calculateScalingFactors(width, height uint, oldWidth, oldHeight float64) (scaleX, scaleY float64) {
	if width == 0 {
		if height == 0 {
			scaleX = 1.0
			scaleY = 1.0
		} else {
			scaleY = oldHeight / float64(height)
			scaleX = scaleY
		}
	} else {
		scaleX = oldWidth / float64(width)
		if height == 0 {
			scaleY = scaleX
		} else {
			scaleY = oldHeight / float64(height)
		}
	}
	return
}

func CalculateNewSizes(albumsRoot string, original *ImageMeta, maxX, maxY uint, name string) *ImageMeta {
	newX := uint(original.Dx)
	newY := uint(original.Dy)

	if newX > newY {
		scaleX, scaleY := calculateScalingFactors(maxX, 0, float64(original.Dx), float64(original.Dy))

		newX = maxX
		newY = uint(float64(original.Dy) / scaleY)

		if false {
			log.Printf("resize-y (%d x %d) -> (%d x %d) (%f x %f)", original.Dx, original.Dy, newX, newY, scaleX, scaleY)
		}
	} else {
		scaleX, scaleY := calculateScalingFactors(0, maxY, float64(original.Dx), float64(original.Dy))

		newX = uint(float64(original.Dx) / scaleX)
		newY = maxY

		if false {
			log.Printf("resize-x (%d x %d) -> (%d x %d) (%f x %f)", original.Dx, original.Dy, newX, newY, scaleX, scaleY)
		}
	}

	return &ImageMeta{
		Path: ResizedPath(albumsRoot, original.Path, name),
		Dx:   newX,
		Dy:   newY,
	}
}

func (g *Generator) ResizePhoto(original string, newSize *ImageMeta) error {
	if _, err := os.Stat(newSize.Path); !os.IsNotExist(err) {
		return nil
	}

	log.Printf("resizing '%s' (%d x %d)", original, newSize.Dx, newSize.Dy)

	originalImage, err := g.Cache.Load(original)
	if err != nil {
		return err
	}

	resizer := nfnt.NewDefaultResizer()
	resizedImage := resizer.Resize(originalImage, newSize.Dx, newSize.Dy)
	if err != nil {
		return err
	}

	err = g.SaveJpeg(resizedImage, newSize.Path)
	if err != nil {
		return err
	}

	return nil
}

func (g *Generator) Resize(albumRoot, original string) error {
	originalMeta, err := getImageMeta(original)
	if err != nil {
		return err
	}

	largeSize := CalculateNewSizes(albumRoot, originalMeta, 1600, 1200, "large")

	err = g.ResizePhoto(original, largeSize)
	if err != nil {
		return err
	}

	smallSize := CalculateNewSizes(albumRoot, originalMeta, 320, 240, "small")

	err = g.ResizePhoto(original, smallSize)
	if err != nil {
		return err
	}

	return nil
}

func (g *Generator) Json(album *Album, path string) error {
	data, err := json.Marshal(album)
	if err != nil {
		return err
	}

	generatedFile, err := os.Create(path)
	if err != nil {
		return err
	}

	defer generatedFile.Close()

	_, err = generatedFile.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (g *Generator) MarkDown(album *Album, path string, templateName string, overwrite bool) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		if !overwrite {
			return nil
		}
	}

	templateData, err := ioutil.ReadFile(templateName)
	if err != nil {
		return err
	}

	template := texttemplate.New(templateName)

	template.Delims("[[", "]]")

	parsed, err := template.Parse(string(templateData))
	if err != nil {
		return err
	}

	generatedFile, err := os.Create(path)
	if err != nil {
		return err
	}

	defer generatedFile.Close()

	err = parsed.Execute(generatedFile, album)
	if err != nil {
		return err
	}

	return nil
}

func (g *Generator) GenerateAlbum(album *Album) error {
	log.Printf("generating '%s' (%d files)", album.Config.Title, len(album.Files))

	mdPath := filepath.Join(g.AlbumsRoot, fmt.Sprintf("%s.md", album.Config.PathName))
	err := g.MarkDown(album, mdPath, "album.md.template", false)
	if err != nil {
		return err
	}

	if false {
		mdGalleryPath := filepath.Join(g.AlbumsRoot, fmt.Sprintf("%s.gallery.md", album.Config.PathName))
		err = g.MarkDown(album, mdGalleryPath, "album.gallery.md.template", true)
		if err != nil {
			return err
		}
	}

	jsonPath := filepath.Join(g.AlbumsRoot, fmt.Sprintf("%s.gallery.json", album.Config.PathName))
	err = g.Json(album, jsonPath)
	if err != nil {
		return err
	}

	if true {
		for _, af := range album.Files {
			err = g.Thumbnails(g.AlbumsRoot, af.OriginalPath, ThumbnailSizes)
			if err != nil {
				return err
			}

			err = g.Resize(g.AlbumsRoot, af.OriginalPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type Options struct {
	AlbumsRoot string
}

func main() {
	o := &Options{}

	flag.StringVar(&o.AlbumsRoot, "albums", "", "albums root directory")

	flag.Parse()

	if o.AlbumsRoot == "" {
		flag.Usage()
		os.Exit(2)
	}

	g, err := NewGenerator("config.json", o.AlbumsRoot)
	if err != nil {
		panic(err)
	}

	for _, album := range g.Cache.AllAlbums {
		err = g.GenerateAlbum(album)
		if err != nil {
			panic(err)
		}
	}
}

func removeAllExtensions(name string) string {
	removed := name
	for {
		maybeExt := filepath.Ext(removed)
		if maybeExt == "" {
			suffixed := strings.Split(removed, "-")
			return suffixed[0]
		}
		removed = strings.TrimSuffix(removed, maybeExt)
	}
}

func copyFile(s, d string) (int64, error) {
	sfs, err := os.Stat(s)
	if err != nil {
		return 0, err
	}

	if !sfs.Mode().IsRegular() {
		return 0, fmt.Errorf("%s should be regular file", s)
	}

	source, err := os.Open(s)
	if err != nil {
		return 0, err
	}

	defer source.Close()

	destination, err := os.Create(d)
	if err != nil {
		return 0, err
	}

	defer destination.Close()

	bytes, err := io.Copy(destination, source)

	return bytes, err
}

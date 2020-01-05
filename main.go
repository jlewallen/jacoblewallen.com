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

type AlbumFile struct {
	Name          string
	OriginalPath  string
	CreatedAt     time.Time
	PhotoPath     string
	ThumbnailPath string
	Xmp           *XmpFile
	Original      *ImageMeta
	Large         *ImageMeta
}

const (
	MainThumbnailSize = 200
)

var (
	ThumbnailSizes = []uint{140, 200, 220, 240, 260}
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

	return filepath.Walk(o.Library.Path, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}

		if info.Mode().IsRegular() {
			base := removeAllExtensions(info.Name())
			if strings.HasSuffix(strings.ToLower(info.Name()), ".xmp") {
				if c.XmpsByBaseName[base] != "" {
					panic(fmt.Sprintf("already have xmp for base: %s", base))
				}
				c.XmpsByBaseName[base] = path
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
	return "", fmt.Errorf("unable to find xmp for: %s", name)
}

func (c *Cache) AddAlbumFile(af *AlbumFile) error {
	for _, hs := range af.Xmp.Rdf.Description.HierarchicalSubjects.Subjects {
		album := c.AlbumsByTag[hs]
		if album != nil {
			log.Printf("adding to album '%s' (%s) : %v", album.Config.Title, hs, af.PhotoPath)
			album.Files = append(album.Files, af)
			album.Date = af.CreatedAt
			break
		}
	}

	return nil
}

type Subjects struct {
	Subjects []string `xml:"Seq>li"`
}

type HierarchicalSubjects struct {
	Subjects []string `xml:"Seq>li"`
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
	Cache *Cache
}

func NewGenerator(configPath string) (g *Generator, err error) {
	g = &Generator{
		Cache: &Cache{},
	}

	cfg, err := g.OpenConfiguration(configPath)
	if err != nil {
		return nil, err
	}

	log.Printf("library: %+v", cfg.Library)

	err = g.Cache.Fill(cfg)
	if err != nil {
		return nil, err
	}

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

	name := filepath.Base(path)
	xmpPath, err := g.Cache.FindXmp(name)
	if err != nil {
		return err
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
		// 2019:12:29 10:46:29
		dto, err := time.Parse("2006:01:02 15:04:05", xmp.Rdf.Description.DateTimeOriginal)
		if err != nil {
			return err
		}

		createdAt = dto
	}

	albumFile := &AlbumFile{
		OriginalPath:  path,
		PhotoPath:     name,
		CreatedAt:     createdAt,
		ThumbnailPath: filepath.Join(fmt.Sprintf("%d", MainThumbnailSize), name),
		Name:          name,
		Original:      originalMeta,
		Large:         CalculateNewSizes(originalMeta, 1600, 1200, "large"),
		Xmp:           xmp,
	}

	g.Cache.AddAlbumFile(albumFile)

	return nil
}

func (g *Generator) IncludeDirectory(path string) error {
	err := filepath.Walk(path, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}

		if info.Mode().IsRegular() {
			fileExt := strings.ToLower(filepath.Ext(info.Name()))

			if fileExt == ".jpg" {
				err := g.IncludeImage(path)
				if err != nil {
					return err
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

func ResizedPath(path string, size string) string {
	dir := filepath.Dir(path)
	name := filepath.Base(path)
	return filepath.Join(dir, size, name)
}

func ThumbnailPath(path string, size uint) string {
	return ResizedPath(path, fmt.Sprintf("%d", size))
}

func (g *Generator) HasAllThumbnails(path string, sizes []uint) bool {
	for _, size := range sizes {
		tp := ThumbnailPath(path, size)
		if _, err := os.Stat(tp); os.IsNotExist(err) {
			return false
		}
	}

	return true
}

func (g *Generator) Thumbnails(path string, sizes []uint) error {
	if g.HasAllThumbnails(path, sizes) {
		return nil
	}

	original, err := g.Cache.Load(path)
	if err != nil {
		return err
	}

	for _, size := range sizes {
		tp := ThumbnailPath(path, size)
		if _, err := os.Stat(tp); os.IsNotExist(err) {
			err := g.Thumbnail(original, size, tp)
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

func CalculateNewSizes(original *ImageMeta, maxX, maxY uint, name string) *ImageMeta {
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
		Path: ResizedPath(original.Path, name),
		Dx:   newX,
		Dy:   newY,
	}
}

func (g *Generator) ResizePhoto(path string, newSize *ImageMeta) error {
	if _, err := os.Stat(newSize.Path); !os.IsNotExist(err) {
		return nil
	}

	log.Printf("resizing '%s' (%d x %d)", path, newSize.Dx, newSize.Dy)

	original, err := g.Cache.Load(path)
	if err != nil {
		return err
	}

	resizer := nfnt.NewDefaultResizer()
	resized := resizer.Resize(original, newSize.Dx, newSize.Dy)
	if err != nil {
		return err
	}

	err = g.SaveJpeg(resized, newSize.Path)
	if err != nil {
		return err
	}

	return nil
}

func (g *Generator) Resize(path string) error {
	originalMeta, err := getImageMeta(path)
	if err != nil {
		return err
	}

	newSize := CalculateNewSizes(originalMeta, 1600, 1200, "large")

	err = g.ResizePhoto(path, newSize)
	if err != nil {
		return err
	}

	return nil
}

func (g *Generator) MarkDown(album *Album, path string) error {
	templateData, err := ioutil.ReadFile("album.md.template")
	if err != nil {
		return err
	}

	template := texttemplate.New("album.md")

	template.Delims("[[", "]]")

	template, err = template.Parse(string(templateData))
	if err != nil {
		return err
	}

	generatedFile, err := os.Create(path)
	if err != nil {
		return err
	}

	defer generatedFile.Close()

	err = template.Execute(generatedFile, album)
	if err != nil {
		return err
	}

	return nil
}

func (g *Generator) GenerateAlbum(album *Album, albumsRoot string) error {
	albumImagesPath := filepath.Join(albumsRoot, album.Config.PathName)
	mdPath := filepath.Join(albumsRoot, fmt.Sprintf("%s.md", album.Config.PathName))

	log.Printf("generating '%s' (%d files) %s", album.Config.Title, len(album.Files), albumImagesPath)

	err := os.MkdirAll(albumImagesPath, 0755)
	if err != nil {
		return err
	}

	err = g.MarkDown(album, mdPath)
	if err != nil {
		return err
	}

	for _, af := range album.Files {
		destination := filepath.Join(albumsRoot, album.Config.PathName, af.PhotoPath)

		if _, err := os.Stat(destination); os.IsNotExist(err) {
			log.Printf("copying photo %s %s", af.OriginalPath, af.PhotoPath)

			_, err := copyFile(af.OriginalPath, destination)
			if err != nil {
				return err
			}
		}

		err = g.Thumbnails(destination, ThumbnailSizes)
		if err != nil {
			return err
		}

		err = g.Resize(destination)
		if err != nil {
			return err
		}
	}

	return nil
}

type Options struct {
	AlbumsPath string
}

func main() {
	o := &Options{}

	flag.StringVar(&o.AlbumsPath, "albums", "", "albums root directory")

	flag.Parse()

	if o.AlbumsPath == "" {
		flag.Usage()
		os.Exit(2)
	}

	g, err := NewGenerator("config.json")
	if err != nil {
		panic(err)
	}

	for _, album := range g.Cache.AllAlbums {
		err = g.GenerateAlbum(album, o.AlbumsPath)
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
			return removed
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

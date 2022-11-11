package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"encoding/base64"
	"encoding/hex"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"

	texttemplate "text/template"
)

var (
	// ErrInvalidBlockSize indicates hash blocksize <= 0.
	ErrInvalidBlockSize = errors.New("invalid blocksize")
	// ErrInvalidPKCS7Data indicates bad input to PKCS7 pad or unpad.
	ErrInvalidPKCS7Data = errors.New("invalid PKCS7 data (empty or not padded)")
	// ErrInvalidPKCS7Padding indicates PKCS7 unpad fails to bad input.
	ErrInvalidPKCS7Padding = errors.New("invalid padding on input")
)

type Options struct {
	Plaintext  string
	Ciphertext string
	Passphrase string
	Title      string
	Inline     bool
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}

// pkcs7Pad right-pads the given byte slice with 1 to n bytes, where
// n is the block size. The size of the result is x times n, where x
// is at least 1.
func pkcs7Pad(b []byte, blocksize int) ([]byte, error) {
	if blocksize <= 0 {
		return nil, ErrInvalidBlockSize
	}
	if b == nil || len(b) == 0 {
		return nil, ErrInvalidPKCS7Data
	}
	n := blocksize - (len(b) % blocksize)
	pb := make([]byte, len(b)+n)
	copy(pb, b)
	copy(pb[len(b):], bytes.Repeat([]byte{byte(n)}, n))
	return pb, nil
}

func signAndEncrypt(plainKey string, plaintext []byte) ([]byte, error) {
	// Generate salt and IV, these will be prepended to the cipher text.
	salt, err := generateRandomBytes(16)
	if err != nil {
		return nil, err
	}

	iv, err := generateRandomBytes(16)
	if err != nil {
		return nil, err
	}

	// Generate the key we'll use via pbkdf2
	dk := pbkdf2.Key([]byte(plainKey), salt, 4096, 32, sha1.New)

	block, err := aes.NewCipher(dk)
	if err != nil {
		return nil, err
	}

	// Read file and then pad the plaintext.
	plaintextPadded, err := pkcs7Pad(plaintext, block.BlockSize())
	if err != nil {
		return nil, err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, len(plaintextPadded))
	bm := cipher.NewCBCEncrypter(block, iv)
	bm.CryptBlocks(ciphertext[:], plaintextPadded)

	// This is the payload we sign, hex encoded random data and then the base64 encoded ciphertext.
	payloadEncoded := hex.EncodeToString(salt) + hex.EncodeToString(iv) + base64.StdEncoding.EncodeToString(ciphertext)
	payload := []byte(payloadEncoded)

	// Sign the payload, prepending with the hex encoded HMAC signature.
	keyHash := sha256.Sum256([]byte(plainKey))
	hmac := hmac.New(sha256.New, []byte(hex.EncodeToString(keyHash[:])))
	hmac.Write(payload)
	hmacHash := hmac.Sum(nil)

	signed := hex.EncodeToString(hmacHash) + payloadEncoded

	return []byte(signed), nil
}

func signAndEncryptFile(passphrase, path string) ([]byte, error) {
	plaintext, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	signedCiphertext, err := signAndEncrypt(passphrase, plaintext)
	if err != nil {
		return nil, err
	}

	return signedCiphertext, nil
}

type EncryptedData struct {
	Title      string
	Ciphertext string
}

func generateDecryptor(payload []byte, title, path string) error {
	templateName := "secure.html.template"

	templateData, err := ioutil.ReadFile(templateName)
	if err != nil {
		return err
	}

	template, err := texttemplate.New(templateName).Parse(string(templateData))
	if err != nil {
		return err
	}

	data := &EncryptedData{
		Title:      title,
		Ciphertext: string(payload),
	}

	generatedFile, err := os.Create(path)
	if err != nil {
		return err
	}

	defer generatedFile.Close()

	err = template.Execute(generatedFile, data)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	o := &Options{}

	flag.StringVar(&o.Plaintext, "plaintext", "", "plaintext")
	flag.StringVar(&o.Ciphertext, "ciphertext", "", "ciphertext")
	flag.StringVar(&o.Passphrase, "passphrase", "", "passphrase")
	flag.StringVar(&o.Title, "title", "", "title")
	flag.BoolVar(&o.Inline, "inline", false, "inline")

	flag.Parse()

	if o.Plaintext == "" || o.Ciphertext == "" || o.Passphrase == "" {
		flag.Usage()
		os.Exit(2)
	}

	if o.Inline {
		d, err := ioutil.ReadFile(o.Plaintext)
		if err != nil {
			panic(err)
		}

		err = SecureInline(bytes.NewReader(d), o.Passphrase, o.Ciphertext)
		if err != nil {
			panic(err)
		}
	} else {
		if o.Title == "" {
			d, err := ioutil.ReadFile(o.Plaintext)
			if err != nil {
				panic(err)
			}

			title, err := FindHtmlTitle(bytes.NewReader(d))
			if err != nil {
				panic(err)
			}

			o.Title = title
		}

		ciphertext, err := signAndEncryptFile(o.Passphrase, o.Plaintext)
		if err != nil {
			panic(err)
		}

		err = generateDecryptor(ciphertext, o.Title, o.Ciphertext)
		if err != nil {
			panic(err)
		}
	}
}

func isTitleElement(n *html.Node) bool {
	return n.Type == html.ElementNode && n.Data == "title"
}

func findTitleTraverse(n *html.Node) (string, bool) {
	if isTitleElement(n) {
		return n.FirstChild.Data, true
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result, ok := findTitleTraverse(c)
		if ok {
			return result, ok
		}
	}

	return "", false
}

func FindHtmlTitle(r io.Reader) (string, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", err
	}

	title, ok := findTitleTraverse(doc)
	if !ok {
		return "", fmt.Errorf("unable to find title")
	}

	return title, nil
}

func ApplyInlineDecryptorTemplate(w io.Writer, ciphertext string) error {
	templateName := "secure-inline.html.template"
	templateData, err := ioutil.ReadFile(templateName)
	if err != nil {
		return err
	}
	template, err := texttemplate.New(templateName).Parse(string(templateData))
	if err != nil {
		return err
	}
	data := &EncryptedData{
		Ciphertext: ciphertext,
	}

	err = template.Execute(w, data)
	if err != nil {
		return err
	}

	return nil
}

func SecureInline(r io.Reader, passphrase, path string) error {
	doc, err := html.Parse(r)
	if err != nil {
		return err
	}

	bodyNode, err := FindNodeWithClass(doc, "jlewallen-private-body")
	if err != nil {
		log.Printf("unable to find jlewallen-private-body")
		return nil
	}

	renderedBody := renderNode(bodyNode)

	signedCiphertext, err := signAndEncrypt(passphrase, []byte(renderedBody))
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = ApplyInlineDecryptorTemplate(&buf, string(signedCiphertext))
	if err != nil {
		return err
	}

	// Now we parse the template output into a fragment and replace
	// the old body with this new part, writing the file out.

	parseCtx := &html.Node{
		Type:     html.ElementNode,
		Data:     "body",
		DataAtom: atom.Body,
	}

	pf, err := html.ParseFragment(&buf, parseCtx)
	if err != nil {
		return err
	}

	bodyNode.FirstChild = pf[0]
	bodyNode.LastChild = pf[0]

	generatedFile, err := os.Create(path)
	if err != nil {
		return err
	}

	defer generatedFile.Close()

	html.Render(generatedFile, doc)

	return nil
}

func hasClass(n *html.Node, klass string) bool {
	for _, attr := range n.Attr {
		if attr.Key == "class" {
			return strings.Contains(attr.Val, klass)
		}
	}
	return false
}

func FindNodeWithClass(doc *html.Node, class string) (*html.Node, error) {
	var found *html.Node
	var crawler func(*html.Node)

	crawler = func(node *html.Node) {
		if node.Type == html.ElementNode && (node.Data == "div" || node.Data == "article") {
			if hasClass(node, class) {
				found = node
				return
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			crawler(child)
		}
	}

	crawler(doc)

	if found != nil {
		return found, nil
	}

	return nil, errors.New("missing body node in the tree")
}

func renderNode(n *html.Node) string {
	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, n)
	return buf.String()
}

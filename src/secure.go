package main

import (
	"bytes"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"

	"io"

	"encoding/base64"
	"encoding/hex"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/net/html"

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

	flag.Parse()

	if o.Plaintext == "" || o.Ciphertext == "" || o.Passphrase == "" {
		flag.Usage()
		os.Exit(2)
	}

	if o.Title == "" {
		d, err := ioutil.ReadFile(o.Plaintext)
		if err != nil {
			panic(err)
		}

		title, found := FindHtmlTitle(bytes.NewReader(d))
		if found {
			log.Printf("found title '%s'", title)
		} else {
			log.Printf("unable to find title")
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

func FindHtmlTitle(r io.Reader) (string, bool) {
	doc, err := html.Parse(r)
	if err != nil {
		panic("Fail to parse html")
	}

	return findTitleTraverse(doc)
}

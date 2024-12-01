package davinci

import (
	"image"
	"io"
	"io/fs"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

type Generator interface {
	Generate(avatar io.Reader, name, message string) (*EncodedImage, error)
}

type EncodedImage struct {
	Buf         []byte
	Extension   string
	ContentType string
}

func DecodeImage(r io.Reader) (image.Image, error) {
	im, _, err := image.Decode(r)
	return im, err
}

func LoadFont(fs fs.FS, name string) (*truetype.Font, error) {
	fontFile, err := fs.Open(name)
	if err != nil {
		return nil, err
	}
	defer fontFile.Close()

	fontBuf, err := io.ReadAll(fontFile)
	if err != nil {
		return nil, err
	}

	return freetype.ParseFont(fontBuf)
}

func LoadTemplate(fs fs.FS, name string) (image.Image, error) {
	imgBuf, err := fs.Open(name)
	if err != nil {
		return nil, err
	}
	defer imgBuf.Close()

	return DecodeImage(imgBuf)
}

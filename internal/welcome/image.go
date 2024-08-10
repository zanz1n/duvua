package welcome

import (
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/fs"

	"github.com/chai2010/webp"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
)

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

func NewImageGenerator(
	template image.Image,
	ttf *truetype.Font,
	quality float32,
) *ImageGenerator {
	return &ImageGenerator{
		template: template,
		ttf:      ttf,
		quality:  quality,
	}
}

type ImageGenerator struct {
	template image.Image
	ttf      *truetype.Font
	quality  float32
}

func (g *ImageGenerator) Generate(avatar io.Reader, name, message string) ([]byte, error) {
	fgColor := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}

	avatarImg, _, err := image.Decode(avatar)
	if err != nil {
		return nil, err
	}

	fg := image.NewUniform(fgColor)

	rgba := image.NewRGBA(image.Rect(0, 0, 1024, 450))

	draw.BiLinear.Scale(
		rgba,
		image.Rect(32, 84, 320, 366),
		avatarImg,
		avatarImg.Bounds(),
		draw.Src,
		&draw.Options{},
	)

	draw.Draw(rgba, rgba.Bounds(), g.template, image.Pt(0, 0), draw.Over)

	nameCtx := freetype.NewContext()
	nameCtx.SetDPI(72)
	nameCtx.SetFont(g.ttf)
	nameCtx.SetFontSize(49)
	nameCtx.SetClip(image.Rect(375, 167, 961, 231))
	nameCtx.SetDst(rgba)
	nameCtx.SetSrc(fg)
	nameCtx.SetHinting(font.HintingNone)

	messageCtx := freetype.NewContext()
	messageCtx.SetDPI(72)
	messageCtx.SetFont(g.ttf)
	messageCtx.SetFontSize(36)
	messageCtx.SetClip(image.Rect(320, 349, 971, 396))
	messageCtx.SetDst(rgba)
	messageCtx.SetSrc(fg)
	messageCtx.SetHinting(font.HintingNone)

	_, err = nameCtx.DrawString(name, freetype.Pt(375, 217))
	if err != nil {
		return nil, err
	}
	_, err = messageCtx.DrawString(message, freetype.Pt(320, 386))
	if err != nil {
		return nil, err
	}

	b, err := webp.EncodeRGB(rgba, g.quality)
	if err != nil {
		return nil, err
	}

	return b, nil
}

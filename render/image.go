package render

import (
	"image"
	"image/color"
	_ "image/gif" // register gif, jpeg, png format handlers
	_ "image/jpeg"
	"image/png"
	"math"
	"os"

	"golang.org/x/image/draw"
)

func ReadImage(path string) (*image.NRGBA, error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}
	return imgToNRGBA(img), nil
}

func WriteImage(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return err
	}
	return nil
}

func imgToNRGBA(img image.Image) *image.NRGBA {
	b := img.Bounds()
	r := image.NewNRGBA(b)
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			r.Set(x, y, img.At(x, y))
		}
	}
	return r
}

// ImgColorFilter replaces `from` with `to` in `img`, and sets all other pixels
// to color.Transparent
func ImgColorFilter(img *image.NRGBA, from, to color.NRGBA) *image.NRGBA {
	b := img.Bounds()
	r := image.NewNRGBA(b)
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			if img.At(x, y) == from {
				r.SetNRGBA(x, y, to)
			} else {
				r.Set(x, y, color.Transparent)
			}
		}
	}
	return r
}

func ScaleImage(img image.Image, factorX, factorY float64, highQuality bool) (scaled *image.NRGBA) {
	b := img.Bounds()
	newX := int(math.Ceil(factorX * float64(b.Max.X)))
	newY := int(math.Ceil(factorY * float64(b.Max.Y)))
	scaledBounds := image.Rect(0, 0, newX, newY)
	scaledImg := image.NewNRGBA(scaledBounds)
	scaler := draw.NearestNeighbor
	if highQuality {
		scaler = draw.CatmullRom
	}
	scaler.Scale(scaledImg, scaledBounds, img, b, draw.Src, nil)
	return scaledImg
}

func RotateImage90(img *image.NRGBA) (rotated *image.NRGBA) {
	b := img.Bounds()
	rotated = image.NewNRGBA(image.Rect(0, 0, b.Max.Y, b.Max.X))
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			col := img.NRGBAAt(x, y)
			rotated.SetNRGBA(b.Max.Y-y, x, col)
		}
	}
	return
}

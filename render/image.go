package render

import (
	"image"
	"image/color"
	_ "image/gif" // register gif, jpeg, png format handlers
	_ "image/jpeg"
	"image/png"
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

// ImgRGBSplit returns three images containing the RGB components, optionally shifted by some offset
func ImgRGBSplit(img *image.NRGBA, shift int) (*image.NRGBA, *image.NRGBA, *image.NRGBA) {
	bounds := img.Bounds().Inset(-shift)
	r := image.NewNRGBA(bounds)
	g := image.NewNRGBA(bounds)
	b := image.NewNRGBA(bounds)
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			c := img.At(x, y).(color.NRGBA)
			if c.A != 0 {
				r.Set(x-shift, y-shift, color.NRGBA{R: c.R, A: c.A / 3})
				g.Set(x+shift, y, color.NRGBA{G: c.G, A: c.A / 3})
				b.Set(x-shift, y+shift, color.NRGBA{B: c.B, A: c.A / 3})
			}
		}
	}
	return r, g, b
}

func ScaleImage(img image.Image, factor int) (scaled *image.NRGBA) {
	b := img.Bounds()
	scaledBounds := image.Rect(0, 0, b.Max.X*factor, b.Max.Y*factor)
	scaledImg := image.NewNRGBA(scaledBounds)
	draw.NearestNeighbor.Scale(scaledImg, scaledBounds, img, b, draw.Src, nil)
	return scaledImg
}

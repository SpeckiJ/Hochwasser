package render

import (
	"image"
	"image/color"
	"math"

	"golang.org/x/image/draw"
)

// pattern provides infinite algorithmically generated patterns and implements
// the image.Image interface.
type pattern struct{}

// Image interface
func (c *pattern) ColorModel() color.Model { return color.NRGBAModel }
func (c *pattern) At(x, y int) color.Color { return color.Transparent }
func (c *pattern) Bounds() image.Rectangle {
	return image.Rectangle{image.Point{-1e9, -1e9}, image.Point{1e9, 1e9}}
}

func (p *pattern) ToFixedImage(bounds image.Rectangle) *image.NRGBA {
	img := image.NewNRGBA(bounds)
	draw.Draw(img, bounds, p, image.Point{}, draw.Src)
	return img
}

type StripePattern struct {
	pattern
	Palette color.Palette
	Size    int
}

func (c *StripePattern) At(x, y int) color.Color {
	n := len(c.Palette)
	if n == 0 {
		return color.Transparent
	}

	pxPerStripe := c.Size / n
	if pxPerStripe <= 0 {
		pxPerStripe = 1
	}
	stripeIndex := y / pxPerStripe % n // FIXME: y - c.Size? -> negative modulo fix
	return c.Palette[stripeIndex]
}

func NewPrideImage(p color.Palette, bounds image.Rectangle) image.Image {
	pattern := StripePattern{Palette: p, Size: bounds.Dy()}
	return pattern.ToFixedImage(bounds)
}

var PrideFlags = map[string]color.Palette{
	"lgbti": color.Palette{
		color.NRGBA{0xe4, 0x03, 0x03, 0xff},
		color.NRGBA{0xff, 0x8c, 0x00, 0xff},
		color.NRGBA{0xff, 0xed, 0x00, 0xff},
		color.NRGBA{0x00, 0x80, 0x26, 0xff},
		color.NRGBA{0x00, 0x4d, 0xff, 0xff},
		color.NRGBA{0x75, 0x07, 0x87, 0xff},
	},
	"nonbinary": color.Palette{
		color.NRGBA{0x9c, 0x5c, 0xd4, 0xff},
		color.NRGBA{0xfc, 0xfc, 0xfc, 0xff},
		color.NRGBA{0xfc, 0xf4, 0x34, 0xff},
		color.NRGBA{0x2c, 0x2c, 0x2c, 0xff},
	},
	"trans": color.Palette{
		color.NRGBA{0x5b, 0xde, 0xfa, 0xff},
		color.NRGBA{0xf5, 0xa9, 0xb8, 0xff},
		color.NRGBA{0xff, 0xff, 0xff, 0xff},
		color.NRGBA{0xf5, 0xa9, 0xb8, 0xff},
		color.NRGBA{0x5b, 0xde, 0xfa, 0xff},
	},
}

type SineColorPattern struct {
	pattern
	Luma  uint8
	Freq  float64
	Phase float64
}

func (c *SineColorPattern) At(x, y int) color.Color {
	TAU := 6.28318
	offset := 0.5
	scale := 0.5
	v := math.Atan2(float64(x), float64(y))

	col := color.NRGBA{}
	col.R = uint8((offset+scale*math.Sin(v/c.Freq*TAU))*128 + 128)
	col.G = uint8((offset+scale*math.Sin(v/c.Freq*TAU+TAU/3))*128 + 128)
	col.B = uint8((offset+scale*math.Sin(v/c.Freq*TAU+TAU/1.5))*128 + 128)
	col.A = 0xff

	// col.R = c.Luma
	// col.G = uint8(math.Sin(float64(x)/c.Freq+c.Phase) * 128 + 128)
	// col.B = uint8(math.Cos(float64(y)/c.Freq+c.Phase) * 128 + 128)
	return col
}

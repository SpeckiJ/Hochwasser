package render

import (
	"image"
	"image/color"
	"math"

	"golang.org/x/image/draw"
)

// Pattern provides infinite algorithmically generated Patterns and implements
// the image.Image interface.
type Pattern struct {
	Size image.Point
}

// Image interface
func (c *Pattern) ColorModel() color.Model { return color.NRGBAModel }
func (c *Pattern) At(x, y int) color.Color { return color.Transparent }
func (c *Pattern) Bounds() image.Rectangle {
	return image.Rectangle{image.Point{-1e9, -1e9}, image.Point{1e9, 1e9}}
}

func (p *Pattern) ToFixedImage(bounds image.Rectangle) *image.NRGBA {
	img := image.NewNRGBA(bounds)
	// TODO: allow setting a mask?
	// override size for pattern?
	p.Size = bounds.Size() // override size for this Pattern
	draw.Draw(img, bounds, p, image.Point{}, draw.Src)
	return img
}

type StripePattern struct {
	Pattern
	Palette color.Palette
}

func (c *StripePattern) At(x, y int) color.Color {
	n := len(c.Palette)
	if n == 0 {
		return color.Transparent
	}

	pxPerStripe := c.Size.Y / n
	if pxPerStripe <= 0 {
		pxPerStripe = 1
	}
	stripeIndex := y / pxPerStripe % n // FIXME: y - c.Size? -> negative modulo fix
	return c.Palette[stripeIndex]
}

func NewPrideImage(p color.Palette, bounds image.Rectangle) image.Image {
	Pattern := StripePattern{Palette: p}
	return Pattern.ToFixedImage(bounds)
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

type DynamicPattern struct {
	Pattern

	Color func(x, y int, p *DynamicPattern) (float64, float64, float64) // should return rgb in [0,1] range

	Luma  float64 // color parameter
	Freq  float64 // frequency parameter
	Phase float64 // phase parameter

	Brightness float64 // additive base brightness
	ColFactor  float64 // Pattern intensity
}

func (c *DynamicPattern) At(x, y int) color.Color {
	col := color.NRGBA{}

	if c.ColFactor == 0.0 && c.Brightness == 0.0 {
		c.ColFactor = 1.0
	}
	if c.Color != nil {
		r, g, b := c.Color(x, y, c)
		col.R = uint8(r * c.ColFactor * 255)
		col.G = uint8(g * c.ColFactor * 255)
		col.B = uint8(b * c.ColFactor * 255)
	}
	if c.Freq == 0.0 {
		c.Freq = 1.0
	}

	col.R = col.R + uint8(255*c.Brightness)
	col.G = col.G + uint8(255*c.Brightness)
	col.B = col.B + uint8(255*c.Brightness)
	col.A = 0xff

	return col
}

var DynPatterns = map[string]image.Image{
	"rbow": &DynamicPattern{
		Color: func(x, y int, p *DynamicPattern) (float64, float64, float64) {
			xx := float64(x) / 26.0 // TODO: dynamic sizing
			yy := float64(y) / 13.0
			r := 0.5 + 0.5*math.Cos(xx+p.Phase)
			g := 0.5 + 0.5*math.Sin(yy+p.Phase)
			b := 0.5 + 0.5*math.Sin(xx+p.Phase)*math.Cos(yy+p.Phase)
			return r, g, b
		},
	},

	"pastel": &DynamicPattern{
		Brightness: 0.5, // TODO: expose these params?
		ColFactor:  0.5,
		Freq:       5.0,

		Color: func(x, y int, p *DynamicPattern) (float64, float64, float64) {
			offset := 0.0
			v := math.Atan2(float64(x)-offset, float64(y)-offset) + p.Phase
			r := 0.5 + 0.5*math.Sin(v*p.Freq)
			g := 0.5 + 0.5*math.Cos(v*p.Freq)
			b := 0.5 + 0.5*math.Sin(v*p.Freq+6.28318/p.Luma)
			return r, g, b
		},
	},
}

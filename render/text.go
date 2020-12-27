package render

import (
	"image"

	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func pt(p fixed.Point26_6) image.Point {
	return image.Point{
		X: int(p.X+32) >> 6,
		Y: int(p.Y+32) >> 6,
	}
}

func RenderText(text string, scale int, texture, bgTex image.Image) *image.NRGBA {
	// @incomplete: draw with texture via Drawer.Src
	face := basicfont.Face7x13
	stringBounds, _ := font.BoundString(face, text)

	b := image.Rectangle{pt(stringBounds.Min), pt(stringBounds.Max)}
	img := image.NewNRGBA(b)

	// fill with black bg
	if bgTex != nil {
		draw.Draw(img, b, bgTex, image.Point{}, draw.Src)
	}

	d := font.Drawer{
		Dst:  img,
		Src:  texture,
		Face: face,
	}
	d.DrawString(text)

	// normalize bounds to start at 0,0
	img.Rect = img.Bounds().Sub(img.Bounds().Min)

	// scale up, as this font is quite small
	return ScaleImage(img, scale)
}

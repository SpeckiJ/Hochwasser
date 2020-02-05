package pixelflut

import (
	"fmt"
	"image"
	"image/color"
)

// CommandsFromImage converts an image to the respective pixelflut commands
func CommandsFromImage(img image.Image, offset image.Point) (commands Commands) {
	b := img.Bounds()
	commands = make([][]byte, b.Size().X*b.Size().Y)
	numCmds := 0

	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			// ensure we're working with RGBA colors (non-alpha-pre-multiplied)
			c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)

			// ignore transparent pixels
			if c.A == 0 {
				continue
			}
			// @incomplete: also send alpha? -> bandwidth tradeoff
			// @speed: this sprintf call is quite slow..
			cmd := fmt.Sprintf("PX %d %d %.2x%.2x%.2x\n",
				x + offset.X, y + offset.Y, c.R, c.G, c.B)
			commands[numCmds] = []byte(cmd)
			numCmds++
		}
	}

	return commands[:numCmds]
}

package pixelflut

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"
)

// Commands represent a list of messages to be sent to a pixelflut server.
type commands [][]byte

// Chunk splits commands into equally sized chunks, while flattening each chunk
// so that all commands are concatenated as a single `[]byte`.
func (c commands) Chunk(numChunks int) [][]byte {
	chunks := make([][]byte, numChunks)
	chunkLength := len(c) / numChunks
	for i := 0; i < numChunks; i++ {
		cmdOffset := i * chunkLength
		for j := 0; j < chunkLength; j++ {
			chunks[i] = append(chunks[i], c[cmdOffset+j]...)
		}
	}
	return chunks
}

// Shuffle reorders commands randomly, in place.
func (c commands) Shuffle() {
	for i := range c {
		j := rand.Intn(i + 1)
		c[i], c[j] = c[j], c[i]
	}
}

// OffsetCmd applies offset to all following requests. Not supported by all servers. example: https://github.com/TobleMiner/shoreline.
func OffsetCmd(x, y int) []byte {
	return []byte(fmt.Sprintf("OFFSET %d %d\n", x, y))
}

// CommandsFromImage converts an image to the respective pixelflut commands
func commandsFromImage(img *image.NRGBA, order RenderOrder, offset image.Point) (cmds commands) {
	b := img.Bounds()
	cmds = make([][]byte, b.Size().X*b.Size().Y)
	numCmds := 0

	max1 := b.Max.X
	max2 := b.Max.Y
	min1 := b.Min.X
	min2 := b.Min.Y
	dir := 1
	if order.IsVertical() {
		max1, max2 = max2, max1
		min1, min2 = min2, min1
	}
	if order.IsReverse() {
		min1, max1 = max1, min1
		min2, max2 = max2, min2
		dir *= -1
	}

	for i1 := min1; i1 != max1; i1 += dir {
		for i2 := min2; i2 != max2; i2 += dir {
			x := i1
			y := i2
			if order.IsVertical() {
				x, y = y, x
			}

			c := img.At(x, y).(color.NRGBA)
			if c.A == 0 {
				continue
			}
			var cmd string
			if c.A != 255 {
				cmd = fmt.Sprintf("PX %d %d %.2x%.2x%.2x%.2x\n",
					x+offset.X, y+offset.Y, c.R, c.G, c.B, c.A)
			} else {
				// @speed: this sprintf call is quite slow..
				cmd = fmt.Sprintf("PX %d %d %.2x%.2x%.2x\n",
					x+offset.X, y+offset.Y, c.R, c.G, c.B)
			}
			cmds[numCmds] = []byte(cmd)
			numCmds++
		}
	}

	cmds = cmds[:numCmds]

	if order == Shuffle {
		cmds.Shuffle()
	}

	return
}

func cmdsFetchImage(bounds image.Rectangle) (cmds commands) {
	cmds = make([][]byte, bounds.Size().X*bounds.Size().Y)
	numCmds := 0
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			cmd := fmt.Sprintf("PX %d %d\n", x, y)
			cmds[numCmds] = []byte(cmd)
			numCmds++
		}
	}
	return cmds
}

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

// CommandsFromImage converts an image to the respective pixelflut commands
func commandsFromImage(img *image.NRGBA, offset image.Point) (cmds commands) {
	b := img.Bounds()
	cmds = make([][]byte, b.Size().X*b.Size().Y)
	numCmds := 0

	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			c := img.At(x, y).(color.NRGBA)
			// ignore transparent pixels
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

	return cmds[:numCmds]
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

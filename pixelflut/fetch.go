package pixelflut

import (
	"fmt"
	"image"
)

func CmdsFetchImage(bounds image.Rectangle) (cmds Commands) {
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

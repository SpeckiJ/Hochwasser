package pixelflut

import (
	"fmt"
	"image"
	"image/color"
	"sync"
	"time"

	"github.com/SpeckiJ/Hochwasser/render"
)

// FlutTask contains all data that is needed to flut
type FlutTask struct {
	FlutTaskOpts
	Img FlutTaskData
}

// FlutTaskOpts specifies parameters of the flut
type FlutTaskOpts struct {
	Address    string
	MaxConns   int
	Offset     image.Point
	Paused     bool
	Shuffle    bool
	RGBSplit   bool
	RandOffset bool
}

// FlutTaskData contains the actual pixeldata to flut, separated because of size
type FlutTaskData = *image.NRGBA

func (t FlutTask) String() string {
	img := "nil"
	if t.Img != nil {
		img = t.Img.Bounds().Size().String()
	}
	return fmt.Sprintf(
		"	%d conns @ %s\n	img %v	offset %v\n	shuffle %v	rgbsplit %v	randoffset %v	paused %v",
		t.MaxConns, t.Address, img, t.Offset, t.Shuffle, t.RGBSplit, t.RandOffset, t.Paused,
	)
}

// IsFlutable indicates if a task is properly initialized & not paused
func (t FlutTask) IsFlutable() bool {
	return t.Img != nil && t.MaxConns > 0 && t.Address != "" && !t.Paused
}

// Flut asynchronously sends the given image to pixelflut server at `address`
//   using `conns` connections. Pixels are sent column wise, unless `shuffle`
//   is set. Stops when stop is closed.
// @cleanup: use FlutTask{} as arg
func Flut(t FlutTask, stop chan bool, wg *sync.WaitGroup) {
	if !t.IsFlutable() {
		return // @robustness: actually return an error here?
	}

	var cmds commands
	if t.RGBSplit {
		// do a RGB split of white
		imgmod := render.ImgColorFilter(t.Img, color.NRGBA{0xff, 0xff, 0xff, 0xff}, color.NRGBA{0xff, 0, 0, 0xff})
		cmds = append(cmds, commandsFromImage(imgmod, image.Pt(t.Offset.X-10, t.Offset.Y-10))...)
		imgmod = render.ImgColorFilter(t.Img, color.NRGBA{0xff, 0xff, 0xff, 0xff}, color.NRGBA{0, 0xff, 0, 0xff})
		cmds = append(cmds, commandsFromImage(imgmod, image.Pt(t.Offset.X+10, t.Offset.Y))...)
		imgmod = render.ImgColorFilter(t.Img, color.NRGBA{0xff, 0xff, 0xff, 0xff}, color.NRGBA{0, 0, 0xff, 0xff})
		cmds = append(cmds, commandsFromImage(imgmod, image.Pt(t.Offset.X-10, t.Offset.Y+10))...)
		cmds = append(cmds, commandsFromImage(t.Img, t.Offset)...)
	} else {
		cmds = commandsFromImage(t.Img, t.Offset)
	}

	if t.Shuffle {
		cmds.Shuffle()
	}

	var messages [][]byte
	var maxOffsetX, maxOffsetY int
	if t.RandOffset {
		maxX, maxY := CanvasSize(t.Address)
		maxOffsetX = maxX - t.Img.Bounds().Canon().Dx()
		maxOffsetY = maxY - t.Img.Bounds().Canon().Dy()
		messages = cmds.Chunk(1) // each connection should send the full img
	} else {
		messages = cmds.Chunk(t.MaxConns)
	}

	bombWg := sync.WaitGroup{}
	for i := 0; i < t.MaxConns; i++ {
		msg := messages[0]
		if len(messages) > i {
			msg = messages[i]
		}

		time.Sleep(66 * time.Millisecond) // avoid crashing the server

		go bombAddress(msg, t.Address, maxOffsetX, maxOffsetY, stop, &bombWg)
	}
	bombWg.Wait()
	if wg != nil {
		wg.Done()
	}
}

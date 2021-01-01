package pixelflut

import (
	"fmt"
	"image"
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
	Address     string
	MaxConns    int
	Offset      image.Point
	Paused      bool
	RGBSplit    bool
	RandOffset  bool
	RenderOrder RenderOrder
}

// FlutTaskData contains the actual pixeldata to flut, separated because of size
type FlutTaskData = *image.NRGBA

func (t FlutTask) String() string {
	img := "nil"
	if t.Img != nil {
		img = t.Img.Bounds().Size().String()
	}
	return fmt.Sprintf(
		"	%d conns @ %s\n	img %v	offset %v\n	order %s	rgbsplit %v	randoffset %v	paused %v",
		t.MaxConns, t.Address, img, t.Offset, t.RenderOrder, t.RGBSplit, t.RandOffset, t.Paused,
	)
}

// IsFlutable indicates if a task is properly initialized & not paused
func (t FlutTask) IsFlutable() bool {
	return t.Img != nil && t.MaxConns > 0 && t.Address != "" && !t.Paused
}

type RenderOrder uint8

func (t RenderOrder) String() string   { return []string{"→", "↓", "←", "↑", "random"}[t] }
func (t RenderOrder) IsVertical() bool { return t&0b01 != 0 }
func (t RenderOrder) IsReverse() bool  { return t&0b10 != 0 }
func NewOrder(v string) RenderOrder {
	switch v {
	case "ltr", "l", "→":
		return LeftToRight
	case "rtl", "r", "←":
		return RightToLeft
	case "ttb", "t", "↓":
		return TopToBottom
	case "btt", "b", "↑":
		return BottomToTop
	default:
		return Shuffle
	}
}

const (
	LeftToRight = 0b000
	TopToBottom = 0b001
	RightToLeft = 0b010
	BottomToTop = 0b011
	Shuffle     = 0b100
)

// Flut asynchronously sends the given image to pixelflut server at `address`
//   using `conns` connections. Pixels are sent column wise, unless `shuffle`
//   is set. Stops when stop is closed.
// @cleanup: use FlutTask{} as arg
func Flut(t FlutTask, stop chan bool, wg *sync.WaitGroup) {
	if !t.IsFlutable() {
		return // @robustness: actually return an error here?
	}

	cmds := generateCommands(t)

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

		time.Sleep(50 * time.Millisecond) // avoid crashing the server

		go bombAddress(msg, t.Address, maxOffsetX, maxOffsetY, stop, &bombWg)
	}
	bombWg.Wait()
	if wg != nil {
		wg.Done()
	}
}

func generateCommands(t FlutTask) (cmds commands) {
	if t.RGBSplit {
		r, g, b := render.ImgRGBSplit(t.Img, 10)
		cmds = append(cmds, commandsFromImage(r, t.RenderOrder, t.Offset)...)
		cmds = append(cmds, commandsFromImage(g, t.RenderOrder, t.Offset)...)
		cmds = append(cmds, commandsFromImage(b, t.RenderOrder, t.Offset)...)
		if t.RenderOrder == Shuffle {
			cmds.Shuffle()
		}
	} else {
		cmds = append(cmds, commandsFromImage(t.Img, t.RenderOrder, t.Offset)...)
	}
	return
}

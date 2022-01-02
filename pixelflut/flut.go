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
	Address     string
	MaxConns    int
	Paused      bool
	RGBSplit    bool // @cleanup: replace with `FX: []Effect`
	RenderOrder RenderOrder
	Offset      RandOffsetter
}

// FlutTaskData contains the actual pixeldata to flut, separated because of size
type FlutTaskData = *image.NRGBA

func (t FlutTask) String() string {
	img := "nil"
	if t.Img != nil {
		img = t.Img.Bounds().Size().String()
	}
	return fmt.Sprintf(
		"	%d conns @ %s\n	img %v	offset %v\n	order %s	rgbsplit %v	paused %v",
		t.MaxConns, t.Address, img, t.Offset, t.RenderOrder, t.RGBSplit, t.Paused,
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

// Flut asynchronously executes the given FlutTask, until `stop` is closed.
func Flut(t FlutTask, stop chan bool, wg *sync.WaitGroup) {
	if !t.IsFlutable() {
		return // @robustness: actually return an error here?
	}

	cmds := generateCommands(t)

	var messages [][]byte
	var maxOffsetX, maxOffsetY int
	if t.Offset.Random {
		maxX, maxY := CanvasSize(t.Address)
		maxOffsetX = maxX - t.Img.Bounds().Canon().Dx()
		maxOffsetY = maxY - t.Img.Bounds().Canon().Dy()
		t.Offset.SetMaximumOffset(image.Pt(maxOffsetX, maxOffsetY))
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

		go bombAddress(msg, t.Address, &t.Offset, stop, &bombWg)
	}
	bombWg.Wait()
	if wg != nil {
		wg.Done()
	}
}

func generateCommands(t FlutTask) (cmds commands) {
	if t.RGBSplit {
		white := color.NRGBA{0xff, 0xff, 0xff, 0xff}
		imgmod := render.ImgColorFilter(t.Img, white, color.NRGBA{0xff, 0, 0, 0xff})
		// FIXME: this offset is ignored with the latest changes, restore this behaviour.
		cmds = append(cmds, commandsFromImage(imgmod, t.RenderOrder, t.Offset.Add(image.Pt(-10, -10)))...)
		imgmod = render.ImgColorFilter(t.Img, white, color.NRGBA{0, 0xff, 0, 0xff})
		cmds = append(cmds, commandsFromImage(imgmod, t.RenderOrder, t.Offset.Add(image.Pt(10, 0)))...)
		imgmod = render.ImgColorFilter(t.Img, white, color.NRGBA{0, 0, 0xff, 0xff})
		cmds = append(cmds, commandsFromImage(imgmod, t.RenderOrder, t.Offset.Add(image.Pt(-10, 10)))...)
	}
	cmds = append(cmds, commandsFromImage(t.Img, t.RenderOrder, t.Offset)...)
	return
}

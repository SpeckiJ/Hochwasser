package pixelflut

import (
	"bufio"
	"encoding/hex"
	"image"
	"image/color"
	"log"
	"net"
	"sync"

	"github.com/SpeckiJ/Hochwasser/render"
)

var funmode = true

// Flut asynchronously sends the given image to pixelflut server at `address`
//   using `conns` connections. Pixels are sent column wise, unless `shuffle`
//   is set. Stops when stop is closed.
// @cleanup: use FlutTask{} as arg
func Flut(img *image.NRGBA, position image.Point, shuffle bool, address string, conns int, stop chan bool, wg *sync.WaitGroup) {
	var cmds commands
	if funmode {
		// do a RGB split of white
		imgmod := render.ImgReplaceColors(img, color.NRGBA{0xff, 0xff, 0xff, 0xff}, color.NRGBA{0xff, 0, 0, 0xff})
		cmds = append(cmds, commandsFromImage(imgmod, image.Pt(position.X-10, position.Y-10))...)
		imgmod = render.ImgReplaceColors(img, color.NRGBA{0xff, 0xff, 0xff, 0xff}, color.NRGBA{0, 0xff, 0, 0xff})
		cmds = append(cmds, commandsFromImage(imgmod, image.Pt(position.X+10, position.Y))...)
		imgmod = render.ImgReplaceColors(img, color.NRGBA{0xff, 0xff, 0xff, 0xff}, color.NRGBA{0, 0, 0xff, 0xff})
		cmds = append(cmds, commandsFromImage(imgmod, image.Pt(position.X-10, position.Y+10))...)
		cmds = append(cmds, commandsFromImage(img, position)...)
	} else {
		cmds = commandsFromImage(img, position)
	}

	if shuffle {
		cmds.Shuffle()
	}
	messages := cmds.Chunk(conns)

	bombWg := sync.WaitGroup{}
	for _, msg := range messages {
		bombWg.Add(1)
		go bombAddress(msg, address, stop, &bombWg)
	}
	bombWg.Wait()
	if wg != nil {
		wg.Done()
	}
}

// FetchImage asynchronously uses `conns` to fetch pixels within `bounds` from
//   a pixelflut server at `address`, and writes them into the returned Image.
func FetchImage(bounds image.Rectangle, address string, conns int, stop chan bool) (img *image.NRGBA) {
	img = image.NewNRGBA(bounds)
	cmds := cmdsFetchImage(bounds).Chunk(conns)

	for i := 0; i < conns; i++ {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			log.Fatal(err)
		}

		go readPixels(img, conn, stop)
		go bombConn(cmds[i], conn, stop)
	}

	return img
}

func readPixels(target *image.NRGBA, conn net.Conn, stop chan bool) {
	reader := bufio.NewReader(conn)
	col := make([]byte, 3)
	for {
		select {
		case <-stop:
			return

		default:
			res, err := reader.ReadSlice('\n')
			if err != nil {
				log.Fatal(err)
			}

			// parse response ("PX <x> <y> <col>\n")
			colorStart := len(res) - 7
			xy := res[3 : colorStart-1]
			yStart := 0
			for yStart = len(xy) - 2; yStart >= 0; yStart-- {
				if xy[yStart] == ' ' {
					break
				}
			}
			x := asciiToInt(xy[:yStart])
			y := asciiToInt(xy[yStart+1:])
			hex.Decode(col, res[colorStart:len(res)-1])

			target.SetNRGBA(x, y, color.NRGBA{col[0], col[1], col[2], 255})
		}
	}
}

func asciiToInt(buf []byte) (v int) {
	for _, c := range buf {
		v = v*10 + int(c-'0')
	}
	return v
}

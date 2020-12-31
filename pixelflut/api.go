package pixelflut

import (
	"bufio"
	"encoding/hex"
	"image"
	"image/color"
	"log"
	"net"
	"sync"
	"time"

	"github.com/SpeckiJ/Hochwasser/render"
)

// Flut asynchronously sends the given image to pixelflut server at `address`
//   using `conns` connections. Pixels are sent column wise, unless `shuffle`
//   is set. Stops when stop is closed.
// @cleanup: use FlutTask{} as arg
func Flut(img *image.NRGBA, position image.Point, shuffle, rgbsplit, randoffset bool, address string, conns int, stop chan bool, wg *sync.WaitGroup) {
	var cmds commands
	if rgbsplit {
		// do a RGB split of white
		imgmod := render.ImgColorFilter(img, color.NRGBA{0xff, 0xff, 0xff, 0xff}, color.NRGBA{0xff, 0, 0, 0xff})
		cmds = append(cmds, commandsFromImage(imgmod, image.Pt(position.X-10, position.Y-10))...)
		imgmod = render.ImgColorFilter(img, color.NRGBA{0xff, 0xff, 0xff, 0xff}, color.NRGBA{0, 0xff, 0, 0xff})
		cmds = append(cmds, commandsFromImage(imgmod, image.Pt(position.X+10, position.Y))...)
		imgmod = render.ImgColorFilter(img, color.NRGBA{0xff, 0xff, 0xff, 0xff}, color.NRGBA{0, 0, 0xff, 0xff})
		cmds = append(cmds, commandsFromImage(imgmod, image.Pt(position.X-10, position.Y+10))...)
		cmds = append(cmds, commandsFromImage(img, position)...)
	} else {
		cmds = commandsFromImage(img, position)
	}

	if shuffle {
		cmds.Shuffle()
	}

	var messages [][]byte
	var maxOffsetX, maxOffsetY int
	if randoffset {
		maxX, maxY := CanvasSize(address)
		maxOffsetX = maxX - img.Bounds().Canon().Dx()
		maxOffsetY = maxY - img.Bounds().Canon().Dy()
		messages = cmds.Chunk(1) // each connection should send the full img
	} else {
		messages = cmds.Chunk(conns)
	}

	bombWg := sync.WaitGroup{}
	for i := 0; i < conns; i++ {
		msg := messages[0]
		if len(messages) > i {
			msg = messages[i]
		}

		time.Sleep(66 * time.Millisecond) // avoid crashing the server

		go bombAddress(msg, address, maxOffsetX, maxOffsetY, stop, &bombWg)
	}
	bombWg.Wait()
	if wg != nil {
		wg.Done()
	}
}

// CanvasSize returns the size of the canvas as returned by the server
func CanvasSize(address string) (int, int) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	conn.Write([]byte("SIZE\n"))
	reader := bufio.NewReader(conn)
	res, err := reader.ReadSlice('\n')
	if err != nil {
		log.Fatal(err)
	}
	return parseXY(res[5:])
}

// FetchImage asynchronously uses `conns` to fetch pixels within `bounds` from
// a pixelflut server at `address`, and writes them into the returned Image.
// If bounds is nil, the server's entire canvas is fetched.
func FetchImage(bounds *image.Rectangle, address string, conns int, stop chan bool) (img *image.NRGBA) {
	if bounds == nil {
		x, y := CanvasSize(address)
		bounds = &image.Rectangle{Max: image.Pt(x, y)}
	}

	img = image.NewNRGBA(*bounds)
	cmds := cmdsFetchImage(*bounds).Chunk(conns)

	for i := 0; i < conns; i++ {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			log.Fatal(err)
		}

		go readPixels(img, conn, stop)
		go bombConn(cmds[i], 0, 0, conn, stop)
	}

	return img
}

func readPixels(target *image.NRGBA, conn net.Conn, stop chan bool) {
	reader := bufio.NewReader(conn)
	col := make([]byte, 4)
	for {
		select {
		case <-stop:
			return

		default:
			res, err := reader.ReadSlice('\n')
			if err != nil {
				log.Fatal(err)
			}

			// parse response ("PX <x> <y> <rrggbbaa>\n")
			// NOTE: shoreline sends alpha, pixelnuke does not!
			colorStart := len(res) - 9
			x, y := parseXY(res[3:colorStart])
			hex.Decode(col, res[colorStart:len(res)-1])

			target.SetNRGBA(x, y, color.NRGBA{col[0], col[1], col[2], col[3]})
		}
	}
}

func parseXY(xy []byte) (int, int) {
	last := len(xy) - 1
	yStart := last - 1 // y is at least one char long
	for ; yStart >= 0; yStart-- {
		if xy[yStart] == ' ' {
			break
		}
	}
	x := asciiToInt(xy[:yStart])
	y := asciiToInt(xy[yStart+1 : last])
	return x, y
}

func asciiToInt(buf []byte) (v int) {
	for _, c := range buf {
		v = v*10 + int(c-'0')
	}
	return v
}

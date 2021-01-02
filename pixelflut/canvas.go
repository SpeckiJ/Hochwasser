package pixelflut

import (
	"bufio"
	"encoding/hex"
	"image"
	"image/color"
	"log"
	"net"
)

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

			// parse response ("PX <x> <y> <rrggbb>")
			colorStart := len(res) - 7
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

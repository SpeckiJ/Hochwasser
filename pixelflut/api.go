package pixelflut

import (
	"bufio"
	"encoding/hex"
	"image"
	"image/color"
	"log"
	"net"
)

// Flut asynchronously sends the given image to pixelflut server at `address`
//   using `conns` connections. Pixels are sent row wise, unless `shuffle` is set.
func Flut(img image.Image, position image.Point, shuffle bool, address string, conns int) {
	cmds := commandsFromImage(img, position)
	if shuffle {
		cmds.Shuffle()
	}

	messages := cmds.Chunk(conns)
	for _, msg := range messages {
		go bombAddress(msg, address)
	}
}

// FetchImage asynchronously uses `conns` to fetch pixels within `bounds` from
//   a pixelflut server at `address`, and writes them into the returned Image.
func FetchImage(bounds image.Rectangle, address string, conns int) (img *image.NRGBA) {
	img = image.NewNRGBA(bounds)
	cmds := cmdsFetchImage(bounds).Chunk(conns)

	for i := 0; i < conns; i++ {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			log.Fatal(err)
		}

		// @cleanup: parsePixels calls conn.Close(), as deferring it here would
		//   instantly close it
		go readPixels(img, conn)
		go bombConn(cmds[i], conn)
	}

	return img
}

func readPixels(target *image.NRGBA, conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	col := make([]byte, 3)
	for {
		res, err := reader.ReadSlice('\n')
		if err != nil {
			log.Fatal(err)
		}

		// parse response ("PX <x> <y> <col>\n")
		colorStart := len(res) - 7
		xy := res[3:colorStart - 1]
		yStart := 0
		for yStart = len(xy) - 2; yStart >= 0; yStart-- {
			if xy[yStart] == ' ' {
				break
			}
		}
		x := asciiToInt(xy[:yStart])
		y := asciiToInt(xy[yStart + 1:])
		hex.Decode(col, res[colorStart:len(res) - 1])

		target.SetNRGBA(x, y, color.NRGBA{ col[0], col[1], col[2], 255 })
	}
}

func asciiToInt(buf []byte) (v int) {
	for _, c := range buf {
		v = v * 10 + int(c - '0')
	}
	return v
}

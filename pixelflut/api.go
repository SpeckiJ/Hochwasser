package pixelflut

import (
	"bufio"
	"encoding/hex"
	"image"
	"image/color"
	"log"
	"net"
	"net/textproto"
	"strconv"
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
		go parsePixels(img, conn)
		go bombConn(cmds[i], conn)
	}

	return img
}

func parsePixels(target *image.NRGBA, conn net.Conn) {
	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)
	defer conn.Close()

	for {
		// @speed: textproto seems not the fastest, buffer text manually & split at \n ?
		res, err := tp.ReadLine()
		if err != nil {
			log.Fatal(err)
		}

		// parse response ("PX <x> <y> <col>")
		// @incomplete errorhandling
		colorStart := len(res) - 6
		xy := res[3:colorStart - 1]
		yStart := 0
		for yStart = len(xy) - 1; yStart >= 0; yStart-- {
			if xy[yStart] == ' ' {
				break
			}
		}
		x, _ := strconv.Atoi(xy[:yStart])
		y, _ := strconv.Atoi(xy[yStart+1:])
		col, _ := hex.DecodeString(res[colorStart:])

		target.Set(x, y, color.NRGBA{
			uint8(col[0]),
			uint8(col[1]),
			uint8(col[2]),
			255,
		})
	}
}

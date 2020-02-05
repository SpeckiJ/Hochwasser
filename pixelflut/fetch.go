package pixelflut

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"log"
	"net"
	"net/textproto"
	"strings"
)

func FetchImage(img *image.NRGBA, address string) {
	// FIXME @speed: this is unusably s l o w w w
	// bottleneck seems to be our pixel reading/parsing code. cpuprofile!
	// -> should buffer it just as in bomb()

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)

	b := img.Bounds()
	for {
		for x := b.Min.X; x < b.Max.X; x++ {
			for y := b.Min.Y; y < b.Max.Y; y++ {
				// request pixel
				fmt.Fprintf(conn, "PX %d %d\n", x, y)
				if err != nil {
					log.Fatal(err)
				}

				// read pixel
				// @speed try to run this in a separate goroutine?
				// we probably need to buffer the responses then
				res, err := tp.ReadLine()
				if err != nil {
					log.Fatal(err)
				}
				res2 := strings.Split(res, " ")
				col, _ := hex.DecodeString(res2[3])
				img.Set(x, y, color.NRGBA{
					uint8(col[0]),
					uint8(col[1]),
					uint8(col[2]),
					255,
				})
			}
		}
	}
}

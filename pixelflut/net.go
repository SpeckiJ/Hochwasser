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
	"strings"
)

// @speed: add some performance reporting mechanism on these functions when
//   called as goroutines

// Bomb writes the given message via plain TCP to the given address,
// forever, as fast as possible.
func Bomb(message []byte, address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	Bomb2(message, conn)
}

// @cleanup: find common interface instead of Bomb2
func Bomb2(message []byte, conn net.Conn) {
	for {
		_, err := conn.Write(message)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func FetchPixels(target *image.NRGBA, conn net.Conn) {
	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)

	for {
		// @speed: textproto seems not the fastest, buffer text manually & split at \n ?
		res, err := tp.ReadLine()
		if err != nil {
			log.Fatal(err)
		}

		// @speed: Split is ridiculously slow due to mallocs!
		//   chunk last 6 chars off -> color, remove first 3 chars, find space in
		//   remainder, then Atoi() xy?
		res2 := strings.Split(res, " ")
		x, _ := strconv.Atoi(res2[1])
		y, _ := strconv.Atoi(res2[2])
		col, _ := hex.DecodeString(res2[3])

		target.Set(x, y, color.NRGBA{
			uint8(col[0]),
			uint8(col[1]),
			uint8(col[2]),
			255,
		})
	}
}

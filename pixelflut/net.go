package pixelflut

import (
	"net"
	"log"
)

// Bomb writes the given message via plain TCP to the given address,
// forever, as fast as possible.
func Bomb(message []byte, address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	for {
		_, err := conn.Write(message)
		if err != nil {
			log.Fatal(err)
		}
	}
}

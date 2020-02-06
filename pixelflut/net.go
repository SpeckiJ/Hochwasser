package pixelflut

import (
	"log"
	"net"
)

// @speed: add some performance reporting mechanism on these functions when
//   called as goroutines

// bombAddress writes the given message via plain TCP to the given address,
// forever, as fast as possible.
func bombAddress(message []byte, address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	bombConn(message, conn)
}

func bombConn(message []byte, conn net.Conn) {
	for {
		_, err := conn.Write(message)
		if err != nil {
			log.Fatal(err)
		}
	}
}

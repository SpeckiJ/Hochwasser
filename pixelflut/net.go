package pixelflut

import (
	"log"
	"net"
	"sync"
)

// @speed: add some performance reporting mechanism on these functions when
//   called as goroutines

// bombAddress writes the given message via plain TCP to the given address,
// forever, as fast as possible.
func bombAddress(message []byte, address string, stop chan bool, wg *sync.WaitGroup) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	bombConn(message, conn, stop)
	wg.Done()
}

func bombConn(message []byte, conn net.Conn, stop chan bool) {
	for {
		select {
		case <-stop:
			log.Println("stopChan bombConn")
			return
		default:
			_, err := conn.Write(message)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

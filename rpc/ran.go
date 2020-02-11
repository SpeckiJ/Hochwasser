package rpc

import (
	"fmt"
	"image"
	"log"
	"net"
	"net/rpc"
	"time"
)

func SummonRán(address string) {
	rán := new(Rán)

	l, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("[rán] rpc server listening on %s\n", l.Addr())

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("[rán] client connected (%v)\n", conn.RemoteAddr())

			client := rpc.NewClient(conn)
			rán.clientConns = append(rán.clientConns, client)

			ack := FlutAck{}
			err = client.Call("Hevring.Flut", RánJob{}, &ack)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("[rán] client accepted job: %v\n", ack.Ok)
		}
	}()

	// poll clients
	go func() {
		for {
			time.Sleep(500 * time.Millisecond)

			var clients []*rpc.Client
			for _, c := range rán.clientConns {
				status := FlutAck{}
				c.Call("Hevring.Status", 0, &status)
				if status.Ok {
					clients = append(clients, c)
				}
			}
			rán.clientConns = clients
			fmt.Printf("[rán] current clients: %v\n", clients)

			// @incomplete: if clients changed, assign tasks anew
		}
	}()
}


type Rán struct {
	clientConns []*rpc.Client
}

type RánJob struct {
	Address  string
	MaxConns int
	Img      image.Image
	Bounds   image.Rectangle
	Shuffle  bool
}

package rpc

import (
	"fmt"
	"image"
	"log"
	"net"
	"net/rpc"
)

func StartRán(address string) {
	rán := new(Rán)
	rpc.Register(rán)

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
			rán.clientAddresses = append(rán.clientAddresses, conn.RemoteAddr())
			fmt.Printf("[rán] client connected (%v)\n", rán.clientAddresses)

			// @incomplete: detect client disconnect, update clients.
			go rpc.ServeConn(conn)

			// @bug: second connection does not send Hello..?!
		}
	}()
}

type Rán struct {
	clientAddresses []net.Addr
}

type RánHelloReq struct{}

type RánJob struct {
	Address  string
	MaxConns int
	Img      image.Image
	Bounds   image.Rectangle
	Shuffle  bool
}

func (z *Rán) Hello(args RánHelloReq, reply *RánJob) error {
	fmt.Printf("[rán] a client said hello!\n")
	reply = nil
	return nil
}

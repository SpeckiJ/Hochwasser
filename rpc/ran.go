package rpc

import (
	"fmt"
	"image"
	"log"
	"net"
	"net/rpc"
	"time"
)

type Rán struct {
	clients []*rpc.Client
	task    FlutTask
}

func SummonRán(address string) *Rán {
	r := new(Rán)

	l, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("[rán] rpc server listening on %s\n", l.Addr())

	// serve tcp port, handshake
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Fatal(err)
			}
			client := rpc.NewClient(conn)
			r.clients = append(r.clients, client)
			fmt.Printf("[rán] client connected (%v). current clients: %v\n",
				conn.RemoteAddr(), len(r.clients))

			if (r.task != FlutTask{}) {
				ack := FlutAck{}
				err = client.Call("Hevring.Flut", r.task, &ack)
				if err != nil || !ack.Ok {
					log.Printf("[rán] client didn't accept task")
				}
			}
		}
	}()

	// poll clients
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)

			var clients []*rpc.Client
			for _, c := range r.clients {
				status := FlutAck{}
				err3 := c.Call("Hevring.Status", 0, &status)
				if err3 == nil || status.Ok {
					clients = append(clients, c)
				}
			}
			if len(r.clients) != len(clients) {
				fmt.Printf("[rán] client disconnected. current clients: %v\n", len(clients))
			}
			r.clients = clients
		}
	}()

	// @incomplete: REPL to change tasks without loosing clients

	return r
}

func (r *Rán) SetTask(img image.Image, offset image.Point, address string, maxConns int) {
	// @incomplete: smart task creation:
	//   fetch server state & sample foreign activity in image regions. assign
	//   subregions to clients (per connection), considering their bandwidth.

	// @bug :imageType
	r.task = FlutTask{address, maxConns, img.(*image.NRGBA), offset, true}
	for _, c := range r.clients {
		ack := FlutAck{}
		// @speed: should send tasks async
		err := c.Call("Hevring.Flut", r.task, &ack)
		if err != nil || !ack.Ok {
			log.Printf("[rán] client didn't accept task")
		}
	}
}

package rpc

import (
	// "io"
	"bufio"
	"fmt"
	"image"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"strings"
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

	// REPL to change tasks without loosing clients
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			input := strings.Split(scanner.Text(), " ")
			cmd := strings.ToLower(input[0])
			args := input[1:]
			if cmd == "stop" {
				for _, c := range r.clients {
					ack := FlutAck{}
					c.Call("Hevring.Stop", 0, &ack) // @speed: async
				}

			} else if cmd == "img" && len(args) > 0 {
				// // @incomplete
				// path := args[0]
				// img := readImage(path)
				// offset := image.Pt(0, 0)
				// if len(args) == 3 {
				// 	x := strconv.Atoi(args[1])
				// 	y := strconv.Atoi(args[2])
				// 	offset = image.Pt(x, y)
				// }
				// task := FlutTask{}
			}
		}
	}()

	// kill clients on exit
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)
		for {
			<-sigChan
			for _, c := range r.clients {
				ack := FlutAck{}
				c.Call("Hevring.Die", 0, &ack) // @speed: async
			}
			os.Exit(0) // @bug :cleanExit
		}
	}()

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

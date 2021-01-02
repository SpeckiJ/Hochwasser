package rpc

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"

	"github.com/SpeckiJ/Hochwasser/pixelflut"
)

// Rán represents the RPC hub, used to coordinate `Hevring` clients.
// Implements `Fluter`
type Rán struct {
	clients []*rpc.Client
	task    pixelflut.FlutTask
	metrics pixelflut.Performance
}

// SummonRán sets up the RPC master, accepting connections at addres (":1234")
// Connects calls methods on each client's rpc provider, killing all clients
// when stopChan is closed.
func SummonRán(address string, stopChan chan bool, wg *sync.WaitGroup) *Rán {
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

			if r.task.IsFlutable() {
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

			r.metrics.Conns = 0
			r.metrics.BytesPerSec = 0
			r.metrics.BytesTotal = 0

			for _, c := range r.clients {
				status := FlutStatus{}
				err := c.Call("Hevring.Status", r.metrics.Enabled, &status)
				if err == nil && status.Ok {
					clients = append(clients, c)
					r.metrics.Conns += status.Conns
					r.metrics.BytesPerSec += status.BytesPerSec
					r.metrics.BytesTotal += status.BytesTotal
				}
			}
			if len(r.clients) != len(clients) {
				fmt.Printf("[rán] client disconnected. current clients: %v\n", len(clients))
			}
			r.clients = clients
		}
	}()

	// print performance
	go func() {
		for {
			time.Sleep(5 * time.Second)
			if r.metrics.Enabled {
				fmt.Println(r.metrics)
			}
		}
	}()

	go RunREPL(r)
	go r.handleExit(stopChan, wg)

	return r
}

func (r *Rán) getTask() pixelflut.FlutTask { return r.task }

func (r *Rán) toggleMetrics() {
	r.metrics.Enabled = !r.metrics.Enabled
}

func (r *Rán) applyTask(t pixelflut.FlutTask) {
	r.task = t
	if !t.IsFlutable() {
		return
	}
	for i, c := range r.clients {
		ack := FlutAck{}
		err := c.Call("Hevring.Flut", r.task, &ack)
		if err != nil || !ack.Ok {
			log.Printf("[rán] client %d didn't accept task", i)
		}
	}
}

func (r *Rán) stopTask() {
	// @robustness: errorchecking
	for _, c := range r.clients {
		ack := FlutAck{}
		c.Call("Hevring.Stop", 0, &ack) // @speed: async
	}
}

func (r *Rán) handleExit(stopChan <-chan bool, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	<-stopChan
	for _, c := range r.clients {
		ack := FlutAck{}
		c.Call("Hevring.Die", 0, &ack) // @speed: async
	}
	// FIXME: why the fuck are we quitting before this loop is complete?
}

// SetTask assigns a pixelflut.FlutTask to Rán, distributing it to all clients
func (r *Rán) SetTask(t pixelflut.FlutTask) {
	// @incomplete: smart task creation:
	//   fetch server state & sample foreign activity in image regions. assign
	//   subregions to clients (per connection), considering their bandwidth.
	r.applyTask(t)
}

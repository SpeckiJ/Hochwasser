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

func ConnectHevring(ránAddress string, stop chan bool, wg *sync.WaitGroup) {
	h := new(Hevring)
	rpc.Register(h)

	fmt.Printf("[hevring] greeting Rán at %s\n", ránAddress)
	conn, err := net.Dial("tcp", ránAddress)
	if err != nil {
		log.Fatal(err)
	}
	go rpc.ServeConn(conn)
	fmt.Printf("[hevring] awaiting task from Rán\n")

	h.quit = stop
	h.wg = wg
	h.wg.Add(1)
	go func() {
		select {
		case <-h.quit:
		}
		if h.taskQuit != nil {
			close(h.taskQuit)
			h.taskQuit = nil
		}
		h.wg.Done()
	}()
}

type Hevring struct {
	task     pixelflut.FlutTask
	taskQuit chan bool
	quit     chan bool
	wg       *sync.WaitGroup
}

type FlutAck struct{ Ok bool }

type FlutStatus struct {
	*pixelflut.Performance
	Ok      bool
	Fluting bool
}

func (h *Hevring) Flut(task pixelflut.FlutTask, reply *FlutAck) error {
	// stop old task if new task is received
	if h.taskQuit != nil {
		close(h.taskQuit)
	}

	fmt.Printf("[hevring] Rán gave us work!\n%v\n", task)
	h.task = task
	h.taskQuit = make(chan bool)

	go pixelflut.Flut(task, h.taskQuit, nil)
	reply.Ok = true
	return nil
}

func (h *Hevring) Status(metrics bool, reply *FlutStatus) error {
	pixelflut.PerformanceReporter.Enabled = metrics
	reply.Performance = pixelflut.PerformanceReporter
	reply.Ok = true
	reply.Fluting = h.taskQuit != nil
	return nil
}

func (h *Hevring) Stop(x int, reply *FlutAck) error {
	if h.taskQuit != nil {
		fmt.Println("[hevring] stopping task")
		h.task = pixelflut.FlutTask{}
		close(h.taskQuit)
		h.taskQuit = nil
		reply.Ok = true
	}
	return nil
}

func (h *Hevring) Die(x int, reply *FlutAck) error {
	// @robustness: waiting for reply to be sent via timeout
	// @incomplete: should try to reconnect for a bit first
	go func() {
		fmt.Println("[hevring] Rán disconnected, stopping")
		time.Sleep(100 * time.Millisecond)
		close(h.quit)
	}()
	reply.Ok = true
	return nil
}

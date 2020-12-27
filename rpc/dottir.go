package rpc

import (
	"fmt"
	"image"
	"log"
	"net"
	"net/rpc"
	"os"
	"time"

	"github.com/SpeckiJ/Hochwasser/pixelflut"
)

func ConnectHevring(ránAddress string) {
	rpc.Register(new(Hevring))

	fmt.Printf("[hevring] greeting Rán at %s\n", ránAddress)
	conn, err := net.Dial("tcp", ránAddress)
	if err != nil {
		log.Fatal(err)
	}
	go rpc.ServeConn(conn)
	fmt.Printf("[hevring] awaiting task from Rán\n")
}

type Hevring struct {
	task     FlutTask
	taskQuit chan bool
}

type FlutTask struct {
	Address  string
	MaxConns int
	Img      *image.NRGBA
	Offset   image.Point
	Shuffle  bool
	// Effects  []string // @idea shuffle, rgbsplit, randoffset, ...
}

type FlutAck struct{ Ok bool }

type FlutStatus struct {
	*pixelflut.Performance
	Ok      bool
	Fluting bool
}

func (h *Hevring) Flut(task FlutTask, reply *FlutAck) error {
	// stop old task if new task is received
	if h.taskQuit != nil {
		close(h.taskQuit)
	}

	fmt.Printf("[hevring] Rán gave us /w o r k/! %v\n", task)
	h.task = task
	h.taskQuit = make(chan bool)

	go pixelflut.Flut(task.Img, task.Offset, task.Shuffle, task.Address, task.MaxConns, h.taskQuit, nil)
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
		h.task = FlutTask{}
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
		time.Sleep(100 * time.Millisecond)
		fmt.Println("[hevring] Rán disconnected, stopping")
		os.Exit(0)
	}()
	reply.Ok = true
	return nil
}

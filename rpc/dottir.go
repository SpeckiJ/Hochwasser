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
}

type FlutAck struct{ Ok bool }

type FlutStatus struct {
	*pixelflut.Performance
	Ok      bool
	Fluting bool
}

func (h *Hevring) Flut(task FlutTask, reply *FlutAck) error {
	if (h.task != FlutTask{}) {
		// @incomplete: stop old task if new task is received
		fmt.Println("[hevring] already have a task")
		reply.Ok = false
		return nil
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
	if (h.task != FlutTask{}) {
		fmt.Println("[hevring] stopping task")
		h.task = FlutTask{}
		close(h.taskQuit)
		h.taskQuit = nil
		reply.Ok = true
	}
	return nil
}

func (h *Hevring) Die(x int, reply *FlutAck) error {
	go func() { // @cleanup: hacky
		time.Sleep(100 * time.Millisecond)
		fmt.Println("[hevring] Rán disconnected, stopping")
		os.Exit(0)
	}()
	reply.Ok = true
	return nil
}

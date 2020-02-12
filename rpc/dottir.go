package rpc

import (
	"fmt"
	"image"
	"log"
	"net"
	"net/rpc"
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
	task FlutTask
}

type FlutTask struct {
	Address  string
	MaxConns int
	Img      *image.NRGBA // bug :imageType: should be image.Image, but can't be serialized. do conversion in task creation?
	Offset   image.Point
	Shuffle  bool
}

type FlutAck struct{ Ok bool }

func (h *Hevring) Flut(task FlutTask, reply *FlutAck) error {
	if (h.task != FlutTask{}) {
		// @incomplete: stop old task if new task is received
		fmt.Println("[hevring] already have a task")
		reply.Ok = false
		return nil
	}

	fmt.Printf("[hevring] Rán gave us /w o r k/! %v\n", task)
	h.task = task
	// @incomplete: async errorhandling
	pixelflut.Flut(task.Img, task.Offset, task.Shuffle, task.Address, task.MaxConns)
	reply.Ok = true
	return nil
}

func (h *Hevring) Status(x int, reply *FlutAck) error {
	// @incomplete: provide performance metrics
	reply.Ok = true
	return nil
}

func (h *Hevring) Stop(x int, reply *FlutAck) error {
	// @incomplete
	if (h.task != FlutTask{}) {
		fmt.Println("[hevring] stopping task")
		h.task = FlutTask{}
		reply.Ok = true
	}
	return nil
}

func (h *Hevring) Die(x int, reply *FlutAck) error {
	go func() { // @cleanup: hacky
		time.Sleep(100 * time.Millisecond)
		log.Fatal("[hevring] Rán disconnected, stopping")
	}()
	reply.Ok = true
	return nil
}

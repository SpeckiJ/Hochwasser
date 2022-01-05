package rpc

import (
	"fmt"
	"image"
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"

	"github.com/SpeckiJ/Hochwasser/pixelflut"
	"github.com/SpeckiJ/Hochwasser/render"
)

func ConnectHevring(ránAddress string, stop chan bool, wg *sync.WaitGroup) *Hevring {
	h := new(Hevring)
	rpc.Register(h)

	fmt.Printf("[hevring] greeting Rán at %s\n", ránAddress)
	conn, err := net.Dial("tcp", ránAddress)
	if err != nil {
		log.Fatal(err)
	}
	go rpc.ServeConn(conn)
	fmt.Printf("[hevring] awaiting task from Rán\n")

	// print performance
	go func() {
		for {
			time.Sleep(5 * time.Second)
			if pixelflut.PerformanceReporter.Enabled {
				fmt.Println(pixelflut.PerformanceReporter)
			}
		}
	}()

	// add listener to stop the task, if this hevring should stop
	// (either because Rán told us so, or we received an interrupt)
	h.quit = stop
	h.wg = wg
	h.wg.Add(1)
	go func() {
		<-h.quit
		h.quit = nil
		if h.taskQuit != nil {
			close(h.taskQuit)
			h.taskQuit = nil
		}
		h.wg.Done()
	}()

	return h
}

type Hevring struct {
	PreviewPath string
	task        pixelflut.FlutTask
	taskQuit    chan bool // if closed, task is stopped.
	quit        chan bool // if closed, kills this hevring
	wg          *sync.WaitGroup
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
	go h.savePreview(task.Img)

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
		if h.quit != nil {
			close(h.quit)
		}
	}()
	reply.Ok = true
	return nil
}

func (h Hevring) savePreview(img image.Image) {
	if h.PreviewPath != "" && img != nil {
		err := render.WriteImage(h.PreviewPath, img)
		if err != nil {
			fmt.Printf("[hevring] unable to write preview: %s\n", err)
		}
	}

}

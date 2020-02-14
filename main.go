package main

import (
	"flag"
	"image"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime/pprof"
	"sync"

	"github.com/SpeckiJ/Hochwasser/pixelflut"
	"github.com/SpeckiJ/Hochwasser/rpc"
)

var err error
var cpuprofile = flag.String("cpuprofile", "", "Destination file for CPU Profile")
var imgPath = flag.String("image", "", "Absolute Path to image")
var x = flag.Int("x", 0, "Offset of posted image from left border")
var y = flag.Int("y", 0, "Offset of posted image from top border")
var connections = flag.Int("connections", 4, "Number of simultaneous connections. Each connection posts a subimage")
var address = flag.String("host", "127.0.0.1:1337", "Server address")
var shuffle = flag.Bool("shuffle", false, "pixel send ordering")
var fetchImgPath = flag.String("fetch-image", "", "path to save the fetched pixel state to")
var ránAddr = flag.String("rán", "", "enable rpc server to distribute jobs, listening on the given address/port")
var hevringAddr = flag.String("hevring", "", "connect to rán rpc server at given address")

func main() {
	flag.Parse()

	// Start cpu profiling if wanted
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// :cleanExit setup
	//   stop chan is closed at end of main process, telling async tasks to stop.
	//   wg waits until async tasks gracefully stopped
	wg := sync.WaitGroup{}
	stopChan := make(chan bool)
	interruptChan := make(chan os.Signal)
	signal.Notify(interruptChan, os.Interrupt)

	if *imgPath != "" {
		offset := image.Pt(*x, *y)
		img := imgToNRGBA(readImage(*imgPath))

		// check connectivity by opening one test connection
		conn, err := net.Dial("tcp", *address)
		if err != nil {
			log.Fatal(err)
		}
		conn.Close()

		if *ránAddr != "" {
			// run RPC server, tasking clients to flut
			wg.Add(1)
			r := rpc.SummonRán(*ránAddr, stopChan, &wg)
			r.SetTask(img, offset, *address, *connections) // @incomplete

		} else {
			// fetch server state and save to file
			// @incomplete: make this available also when not fluting?
			if *fetchImgPath != "" {
				fetchedImg := pixelflut.FetchImage(img.Bounds().Add(offset), *address, 1, stopChan)
				*connections -= 1
				defer writeImage(*fetchImgPath, fetchedImg)
			}

			// local 🌊🌊🌊🌊🌊🌊🌊🌊🌊🌊🌊🌊🌊🌊🌊
			wg.Add(1)
			go pixelflut.Flut(img, offset, *shuffle, *address, *connections, stopChan, &wg)
		}

	} else if *hevringAddr != "" {
		// connect to RPC server and execute their tasks
		rpc.ConnectHevring(*hevringAddr)

	} else {
		log.Fatal("must specify -image or -hevring")
	}

	// :cleanExit logic:
	//   notify all async tasks to stop on interrupt
	//   then wait for clean shutdown of all tasks before exiting
	//   TODO: make this available to all invocation types
	select {
	case <-interruptChan:
	}
	close(stopChan)
	wg.Wait()
}

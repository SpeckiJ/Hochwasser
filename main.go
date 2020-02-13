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
	"time"

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
var runtime = flag.String("runtime", "60s", "exit after timeout")
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

	if *imgPath != "" {
		offset := image.Pt(*x, *y)
		img := readImage(*imgPath)

		// check connectivity by opening one test connection
		conn, err := net.Dial("tcp", *address)
		if err != nil {
			log.Fatal(err)
		}
		conn.Close()

		if *ránAddr != "" {
			// run RPC server, tasking clients to flut
			r := rpc.SummonRán(*ránAddr)
			r.SetTask(img, offset, *address, *connections) // @incomplete
			select {}                                      // block forever

		} else {

			// local 🌊🌊🌊🌊🌊🌊🌊🌊🌊🌊🌊🌊🌊🌊🌊
			var wg sync.WaitGroup
			defer wg.Wait()
			stopChan := make(chan bool)
			defer close(stopChan)

			wg.Add(1) // :cleanExit: is this WG needed? we only have one task running at a time?
			go pixelflut.Flut(img, offset, *shuffle, *address, *connections, stopChan, &wg)

			// fetch server state and save to file
			// @incomplete: make this available also when not fluting?
			if *fetchImgPath != "" {
				fetchedImg := pixelflut.FetchImage(img.Bounds().Add(offset), *address, 1, stopChan)
				*connections -= 1
				defer writeImage(*fetchImgPath, fetchedImg)
			}

			// :cleanExit logic:
			//   notify all async tasks to stop on interrupt or after timeout,
			//   then wait for clean shutdown of all tasks before exiting
			//   TODO: make this available to all invocation types

			timer, err := time.ParseDuration(*runtime)
			if err != nil {
				log.Fatal("Invalid runtime specified: " + err.Error())
			}

			interruptChan := make(chan os.Signal)
			signal.Notify(interruptChan, os.Interrupt)
			select {
			case <-time.After(timer):
			case <-interruptChan:
			}
		}

	} else if *hevringAddr != "" {
		// connect to RPC server and execute their tasks
		rpc.ConnectHevring(*hevringAddr)
		select {} // block forever

	} else {
		log.Fatal("must specify -image or -hevring")
	}
}

package main

import (
	"flag"
	"image"
	"log"
	"net"
	"os"
	"runtime/pprof"
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
			pixelflut.Flut(img, offset, *shuffle, *address, *connections)

			// fetch server state and save to file
			// @incomplete: make this available also when not fluting?
			if *fetchImgPath != "" {
				fetchedImg := pixelflut.FetchImage(img.Bounds().Add(offset), *address, 1)
				*connections -= 1
				defer writeImage(*fetchImgPath, fetchedImg)
			}

			// Terminate after timeout to save resources
			timer, err := time.ParseDuration(*runtime)
			if err != nil {
				log.Fatal("Invalid runtime specified: " + err.Error())
			}
			time.Sleep(timer)
		}

	} else if *hevringAddr != "" {
		// connect to RPC server and execute their tasks
		rpc.ConnectHevring(*hevringAddr)
		select {} // block forever

	} else {
		log.Fatal("must specify -image or -hevring")
	}
}

/**
 * @incomplete: clean exit
 * to ensure cleanup is done (rpc disconnects, cpuprof, image writing, ...),
 * we should catch signals and force-exit all goroutines (bomb, rpc). via channel?
 */

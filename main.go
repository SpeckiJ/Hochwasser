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
var image_path = flag.String("image", "", "Absolute Path to image")
var image_offsetx = flag.Int("xoffset", 0, "Offset of posted image from left border")
var image_offsety = flag.Int("yoffset", 0, "Offset of posted image from top border")
var connections = flag.Int("connections", 4, "Number of simultaneous connections. Each connection posts a subimage")
var address = flag.String("host", "127.0.0.1:1337", "Server address")
var runtime = flag.String("runtime", "60s", "exit after timeout")
var shuffle = flag.Bool("shuffle", false, "pixel send ordering")
var fetchImgPath = flag.String("fetch-image", "", "path to save the fetched pixel state to")
var rán = flag.String("rán", "", "enable rpc server to distribute jobs, listening on the given address/port")
var hevring = flag.String("hevring", "", "connect to rán rpc server at given address")

func main() {
	flag.Parse()
	if *image_path == "" {
		log.Fatal("No image provided")
	}

	// check connectivity by opening one test connection
	conn, err := net.Dial("tcp", *address)
	if err != nil {
		log.Fatal(err)
	}
	conn.Close()

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

	if *rán != "" { // @fixme: should validate proper address?
		rpc.SummonRán(*rán)
	}
	if *hevring != "" { // @fixme: should validate proper address?
		rpc.ConnectHevring(*hevring)
	}

	offset := image.Pt(*image_offsetx, *image_offsety)
	img := readImage(*image_path)

	if *fetchImgPath != "" {
		fetchedImg := pixelflut.FetchImage(img.Bounds().Add(offset), *address, 1)
		*connections -= 1
		defer writeImage(*fetchImgPath, fetchedImg)
	}

	// 🌊🌊🌊🌊🌊🌊🌊🌊🌊🌊🌊🌊🌊🌊🌊
	pixelflut.Flut(img, offset, *shuffle, *address, *connections)

	// Terminate after timeout to save resources
	timer, err := time.ParseDuration(*runtime)
	if err != nil {
		log.Fatal("Invalid runtime specified: " + err.Error())
	}
	time.Sleep(timer)
}

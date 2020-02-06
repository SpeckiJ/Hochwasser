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

	offset := image.Pt(*image_offsetx, *image_offsety)
	img := readImage(*image_path)

	var fetchedImg *image.NRGBA
	if *fetchImgPath != "" {
		fetchedImg = pixelflut.FetchImage(img.Bounds().Add(offset), *address, 1)
		*connections -= 1
	}

	// 🌊🌊🌊🌊🌊🌊🌊🌊🌊🌊🌊🌊🌊🌊🌊
	pixelflut.Flut(img, offset, *shuffle, *address, *connections)

	// Terminate after timeout to save resources
	timer, err := time.ParseDuration(*runtime)
	if err != nil {
		log.Fatal("Invalid runtime specified: " + err.Error())
	}
	time.Sleep(timer)

	if *fetchImgPath != "" {
		writeImage(*fetchImgPath, fetchedImg)
	}
}

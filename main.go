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
	fetchedImg := image.NewNRGBA(img.Bounds().Add(offset))

	if *fetchImgPath != "" {
		// using img.SubImage to distribute tasks is nice, as we can also parallelize command generation easily!
		// @cleanup: use a box tiling algo instead of hardcoding
		b := fetchedImg.Bounds()
		b1 := image.Rectangle{b.Min, b.Size().Div(2)}
		b2 := b1.Add(image.Pt(b1.Dx(), 0))
		b3 := b1.Add(image.Pt(0, b1.Dy()))
		b4 := b1.Add(b1.Size())
		go pixelflut.FetchImage(fetchedImg.SubImage(b1).(*image.NRGBA), *address)
		go pixelflut.FetchImage(fetchedImg.SubImage(b2).(*image.NRGBA), *address)
		go pixelflut.FetchImage(fetchedImg.SubImage(b3).(*image.NRGBA), *address)
		go pixelflut.FetchImage(fetchedImg.SubImage(b4).(*image.NRGBA), *address)
		*connections -= 4
	}

	// Generate and split messages into equal chunks
	commands := pixelflut.CommandsFromImage(img, offset)
	if *shuffle {
		commands.Shuffle()
	}

	if *connections > 0 {
		commandGroups := commands.Chunk(*connections)
		for _, messages := range commandGroups {
			go pixelflut.Bomb(messages, *address)
		}
	}

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

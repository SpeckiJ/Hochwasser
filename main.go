package main

import (
	"flag"
	"net"
	"runtime/pprof"
	"time"
)

import "fmt"
import "image/png"
import "log"
import "os"
import "strconv"
import _ "net/http/pprof"

var err error
var cpuprofile = flag.String("cpuprofile", "", "Destination file for CPU Profile")
var image = flag.String("image", "", "Absolute Path to image")
var canvas_xsize = flag.Int("xsize", 800, "Width of the canvas in px")
var canvas_ysize = flag.Int("ysize", 600, "Height of the canvas in px")
var image_offsetx = flag.Int("xoffset", 0, "Offset of posted image from left border")
var image_offsety = flag.Int("yoffset", 0, "Offset of posted image from top border")
var connections = flag.Int("connections", 10, "Number of simultaneous connections/threads. Each Thread posts a subimage")
var address = flag.String("host", "127.0.0.1:1337", "Server address")
var runtime = flag.String("runtime", "1", "Runtime in Minutes")

func main() {
	flag.Parse()
	if *image == "" || *address == "" {
		log.Fatal("No image or no server address provided")
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
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	// Generate and split messages into equal chunks
	msg := splitmessages(genMessages())
	for _, message := range msg {
		go bomb(message)
	}

	// Terminate after 1 Minute to save resources
	timer, err := time.ParseDuration(*runtime + "m")
	if err != nil {
		log.Fatal("Invalid runtime specified: " + err.Error())
	}
	time.Sleep(time.Minute * timer)
}

func bomb(messages []byte) {
	conn, err := net.Dial("tcp", *address)

	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Start bombardement
	for {
		_, err := conn.Write(messages)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// Creates message based on given image
func genMessages() (output []byte) {
	reader, err := os.Open(*image)
	if err != nil {
		log.Fatal(err)
	}

	img, err2 := png.Decode(reader)
	if err2 != nil {
		log.Fatal(err2)
	}

	for x := img.Bounds().Max.X; x != 0; x-- {
		for y := img.Bounds().Max.Y; y != 0; y-- {
			col := img.At(x, y)
			r, g, b, _ := col.RGBA()

			rStr := strconv.FormatInt(int64(r), 16)
			if len(rStr) == 1 {
				rStr = "0" + rStr
			}

			gStr := strconv.FormatInt(int64(g), 16)
			if len(gStr) == 1 {
				gStr = "0" + gStr
			}

			bStr := strconv.FormatInt(int64(b), 16)
			if len(bStr) == 1 {
				bStr = "0" + bStr
			}

			colStr := rStr[0:2]
			colStr += gStr[0:2]
			colStr += bStr[0:2]

			//Do not draw transparent pixels
			if colStr == "000000" {
				continue
			}
			pxStr := fmt.Sprintf("PX %d %d %s\n", x+*image_offsetx, y+*image_offsety, colStr)
			output = append(output, []byte(pxStr)...)
		}
	}
	return output
}

// Splits messages into chunks, splitting on complete commands only
func splitmessages(in []byte) [][]byte {
	index := 0
	equalsplit := len(in) / *connections
	output := make([][]byte, *connections)
	for i := 0; i < *connections; i++ {
		if index+equalsplit > len(in) {
			output[i] = in[index:]
			break
		}

		tmp := index
		for in[index+equalsplit] != 80 {
			index++
		}
		output[i] = in[tmp : index+equalsplit]
		index += equalsplit
	}
	return output
}

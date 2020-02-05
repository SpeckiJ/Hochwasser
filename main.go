package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math/rand"
	"net"
	"os"
	"runtime/pprof"
	"time"
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
	// Generate and split messages into equal chunks
	commands := genCommands(img, offset)
	if *shuffle {
		shuffleCommands(commands)
	}

	commandGroups := chunkCommands(commands, *connections)
	for _, messages := range commandGroups {
		go bomb(messages)
	}

	// Terminate after timeout to save resources
	timer, err := time.ParseDuration(*runtime)
	if err != nil {
		log.Fatal("Invalid runtime specified: " + err.Error())
	}
	time.Sleep(timer)
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

func readImage(path string) (img image.Image) {
	reader, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	img, _, err2 := image.Decode(reader)
	if err2 != nil {
		log.Fatal(err2)
	}

	return img
}

// Creates message based on given image
func genCommands(img image.Image, offset image.Point) (commands [][]byte) {
	b := img.Bounds()
	commands = make([][]byte, b.Size().X*b.Size().Y)
	numCmds := 0

	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			// ensure we're working with RGBA colors (non-alpha-pre-multiplied)
			c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)

			// ignore transparent pixels
			if c.A == 0 {
				continue
			}
			// @incomplete: also send alpha? -> bandwidth tradeoff
			// @speed: this sprintf call is quite slow..
			cmd := fmt.Sprintf("PX %d %d %.2x%.2x%.2x\n",
				x + offset.X, y + offset.Y, c.R, c.G, c.B)
			commands[numCmds] = []byte(cmd)
			numCmds++
		}
	}

	return commands[:numCmds]
}

// Splits messages into equally sized chunks
func chunkCommands(commands [][]byte, numChunks int) [][]byte {
	chunks := make([][]byte, numChunks)

	chunkLength := len(commands) / numChunks
	for i := 0; i < numChunks; i++ {
		cmdOffset := i * chunkLength
		for j := 0; j < chunkLength; j++ {
			chunks[i] = append(chunks[i], commands[cmdOffset+j]...)
		}
	}
	return chunks
}

func shuffleCommands(slice [][]byte) {
	for i := range slice {
		j := rand.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}

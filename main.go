package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"
	"strconv"
	"time"
)

var err error
var cpuprofile = flag.String("cpuprofile", "", "Destination file for CPU Profile")
var image_path = flag.String("image", "", "Absolute Path to image")
var image_offsetx = flag.Int("xoffset", 0, "Offset of posted image from left border")
var image_offsety = flag.Int("yoffset", 0, "Offset of posted image from top border")
var connections = flag.Int("connections", 10, "Number of simultaneous connections/threads. Each Thread posts a subimage")
var address = flag.String("host", "127.0.0.1:1337", "Server address")
var runtime = flag.String("runtime", "1", "Runtime in Minutes")

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
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// Generate and split messages into equal chunks
	commands := genCommands(readImage(*image_path), *image_offsetx, *image_offsety)
	commandGroups := chunkCommands(commands, *connections)
	for _, messages := range commandGroups {
		go bomb(messages)
	}

	// Terminate after 1 Minute to save resources
	timer, err := time.ParseDuration(*runtime + "m")
	if err != nil {
		log.Fatal("Invalid runtime specified: " + err.Error())
	}
	time.Sleep(time.Minute * timer)
}

func bomb(messages [][]byte) {
	conn, err := net.Dial("tcp", *address)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	// Start bombardement
	for {
		for _, message := range messages {
			_, err := conn.Write(message)
			if err != nil {
				log.Fatal(err)
			}
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

func intToHex(x uint32) string {
	str := strconv.FormatInt(int64(x), 16)
	if len(str) == 1 {
		str = "0" + str
	}
	return str[0:2]
}

// Creates message based on given image
func genCommands(img image.Image, offset_x, offset_y int) (commands [][]byte) {
	max_x := img.Bounds().Max.X
	max_y := img.Bounds().Max.Y
	commands = make([][]byte, max_x*max_y)

	for x := 0; x < max_x; x++ {
		for y := 0; y < max_y; y++ {
			r, g, b, _ := img.At(x, y).RGBA()
			colStr := intToHex(r) + intToHex(g) + intToHex(b)
			cmd := fmt.Sprintf("PX %d %d %s\n", x+offset_x, y+offset_y, colStr)
			commands[x*max_y+y] = []byte(cmd)
		}
	}

	return commands
}

// Splits messages into equally sized chunks
func chunkCommands(commands [][]byte, numChunks int) [][][]byte {
	chunks := make([][][]byte, numChunks)

	chunkLength := len(commands) / numChunks
	for i := 0; i < numChunks; i++ {
		cmdOffset := i * chunkLength

		if cmdOffset+chunkLength > len(commands) {
			chunks[i] = commands[cmdOffset:]
			break
		}
		chunks[i] = commands[cmdOffset : cmdOffset+chunkLength]
	}
	return chunks
}

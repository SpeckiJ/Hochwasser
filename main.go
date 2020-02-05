package main

import (
	"bufio"
	"encoding/hex"
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"log"
	"math/rand"
	"net"
	"net/textproto"
	"os"
	"runtime/pprof"
	"strings"
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
		go fetchImage(fetchedImg.SubImage(b1).(*image.NRGBA))
		go fetchImage(fetchedImg.SubImage(b2).(*image.NRGBA))
		go fetchImage(fetchedImg.SubImage(b3).(*image.NRGBA))
		go fetchImage(fetchedImg.SubImage(b4).(*image.NRGBA))
		*connections -= 4
	}

	// Generate and split messages into equal chunks
	commands := genCommands(img, offset)
	if *shuffle {
		shuffleCommands(commands)
	}

	if *connections > 0 {
		commandGroups := chunkCommands(commands, *connections)
		for _, messages := range commandGroups {
			go bomb(messages)
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

func writeImage(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		log.Fatal(err)
	}
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

func fetchImage(img *image.NRGBA) {
	// FIXME @speed: this is unusably s l o w w w
	// bottleneck seems to be our pixel reading/parsing code. cpuprofile!
	// -> should buffer it just as in bomb()

	conn, err := net.Dial("tcp", *address)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)

	b := img.Bounds()
	for {
		for x := b.Min.X; x < b.Max.X; x++ {
			for y := b.Min.Y; y < b.Max.Y; y++ {
				// request pixel
				fmt.Fprintf(conn, "PX %d %d\n", x, y)
				if err != nil {
					log.Fatal(err)
				}

				// read pixel
				// @speed try to run this in a separate goroutine?
				// we probably need to buffer the responses then
				res, err := tp.ReadLine()
				if err != nil {
					log.Fatal(err)
				}
				res2 := strings.Split(res, " ")
				col, _ := hex.DecodeString(res2[3])
				img.Set(x, y, color.NRGBA{
					uint8(col[0]),
					uint8(col[1]),
					uint8(col[2]),
					255,
				})
			}
		}
	}
}

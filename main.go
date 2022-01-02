package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/SpeckiJ/Hochwasser/pixelflut"
	"github.com/SpeckiJ/Hochwasser/render"
	"github.com/SpeckiJ/Hochwasser/rpc"
)

var (
	imgPath      = flag.String("image", "", "Filepath of an image to flut")
	ránAddr      = flag.String("rán", "", "Start RPC server to distribute jobs, listening on the given address/port")
	hevringAddr  = flag.String("hevring", "", "Connect to PRC server at given address/port")
	address      = flag.String("host", ":1234", "Target server address")
	connections  = flag.Int("connections", 4, "Number of simultaneous connections. Each connection posts a subimage")
	x            = flag.Int("x", 0, "Offset of posted image from left border")
	y            = flag.Int("y", 0, "Offset of posted image from top border")
	order        = flag.String("order", "rtl", "Draw order (shuffle, ltr, rtl, ttb, btt)")
	fetchImgPath = flag.String("fetch", "", "Enable fetching the screen area to the given local file, updating it each second")
	cpuprofile   = flag.String("cpuprofile", "", "Destination file for CPU Profile")
)

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	task := runWithExitHandler(taskFromFlags)
	if *cpuprofile != "" {
		runWithProfiler(*cpuprofile, task)
	} else {
		task()
	}
}

func taskFromFlags(stop chan bool, wg *sync.WaitGroup) {
	rán := *ránAddr
	hev := *hevringAddr

	startServer := rán != "" || (hev == "" && *imgPath != "")
	startClient := hev != "" || (rán == "" && *imgPath != "")
	fetchImg := *fetchImgPath != ""

	if !(startServer || startClient || fetchImg) {
		fmt.Println("Error: At least one of the following flags is needed:\n	-image -rán -hevring\n")
		flag.Usage()
		os.Exit(1)
	}

	if startServer && startClient && rán == "" && hev == "" {
		rán = fmt.Sprintf(":%d", rand.Intn(30000)+1000)
		hev = rán
	}

	if startServer {
		r := rpc.SummonRán(rán, stop, wg)

		var img *image.NRGBA
		if *imgPath != "" {
			var err error
			if img, err = render.ReadImage(*imgPath); err != nil {
				log.Fatal(err)
			}
		}

		r.SetTask(pixelflut.FlutTask{
			FlutTaskOpts: pixelflut.FlutTaskOpts{
				Address:     *address,
				MaxConns:    *connections,
				Offset:      pixelflut.RandOffsetter{Point: image.Pt(*x, *y)},
				RenderOrder: pixelflut.NewOrder(*order),
			},
			Img: img,
		})
	}

	if startClient {
		rpc.ConnectHevring(hev, stop, wg)
	}

	if fetchImg {
		canvasToFile(*fetchImgPath, *address, time.Second, stop, wg)
	}
}

func canvasToFile(filepath, server string, interval time.Duration, stop chan bool, wg *sync.WaitGroup) {
	// async fetch the image
	fetchedImg := pixelflut.FetchImage(nil, server, 1, stop)

	// write it in a fixed interval
	go func() {
		wg.Add(1)
		defer wg.Done()

		for loop := true; loop; {
			select {
			case <-stop:
				loop = false
			case <-time.Tick(interval):
			}
			render.WriteImage(filepath, fetchedImg)
		}
	}()
}

// Takes a non-blocking function, and provides it an interface for graceful shutdown:
// stop chan is closed if the routine should be stopped. before quitting, wg is awaited.
func runWithExitHandler(task func(stop chan bool, wg *sync.WaitGroup)) func() {
	return func() {
		wg := sync.WaitGroup{}
		stopChan := make(chan bool)
		interruptChan := make(chan os.Signal)
		signal.Notify(interruptChan, os.Interrupt)

		task(stopChan, &wg)

		// block until we get an interrupt, or somebody says we need to quit (by closing stopChan)
		select {
		case <-interruptChan:
		case <-stopChan:
			stopChan = nil
		}

		if stopChan != nil {
			// notify all async tasks to stop on interrupt, if channel wasn't closed already
			close(stopChan)
		}

		// then wait for clean shutdown of all tasks before exiting
		wg.Wait()
	}
}

func runWithProfiler(outfile string, task func()) {
	f, err := os.Create(outfile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	task()
}

package pixelflut

import (
	"fmt"
	"image"
	"math/rand"
	"net"
	"sync"
	"time"
)

const (
	timeoutMin = 100 * time.Millisecond
	timeoutMax = 10 * time.Second
)

type Offsetter interface {
	Next() (x, y int)
	SetMaximumOffset(max image.Point)
}

type RandOffsetter struct {
	image.Point
	Mask   *image.NRGBA
	Random bool
	Max    image.Point
}

func (r RandOffsetter) String() string {
	mask := "nil"
	if r.Mask != nil {
		mask = r.Mask.Bounds().Size().String()
	}
	return fmt.Sprintf("[%v + random %v, mask %v]", r.Point, r.Random, mask)
}

// override image.Point interface
func (r RandOffsetter) Add(p image.Point) RandOffsetter {
	r.Point = r.Point.Add(p)
	return r
}

func (r *RandOffsetter) SetMaximumOffset(max image.Point) {
	if r.Mask != nil {
		mask := r.Mask.Bounds().Canon().Max.Sub(r.Point)
		r.Max.X = clamp(mask.X, 1, max.X)
		r.Max.Y = clamp(mask.Y, 1, max.Y)
	} else {
		r.Max.X = clamp(max.X, 1, max.X)
		r.Max.Y = clamp(max.Y, 1, max.Y)
	}
}

func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func (r RandOffsetter) Next() (x, y int) {
	if r.Random {
		for i := 0; i < 1000; i++ {
			x := rand.Intn(r.Max.X)
			y := rand.Intn(r.Max.Y)
			if r.Mask != nil {
				if _, _, _, a := r.Mask.At(x, y).RGBA(); a != 0 {
					return r.X + x, r.Y + y
				}
			} else {
				return r.X + x, r.Y + y
			}
		}
	}

	return r.X, r.Y
}

// Performance contains pixelflut metrics
type Performance struct {
	Enabled     bool
	Conns       int
	BytesPerSec int
	BytesTotal  int

	connsReporter chan int
	bytesReporter chan int
	bytes         int
}

func (p Performance) String() string {
	return fmt.Sprintf("%v conns\t%v\t%v/s",
		p.Conns, fmtBytes(p.BytesTotal), fmtBit(p.BytesPerSec))
}

// https://yourbasic.org/golang/byte-count.go
func fmtBytes(b int) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func fmtBit(b int) string {
	const unit = 1000
	b *= 8
	if b < unit {
		return fmt.Sprintf("%d b", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cb",
		float64(b)/float64(div), "kMGTPE"[exp])
}

// PerformanceReporter provides pixelflut performance metrics, when Enabled is true.
//   @speed: Note that enabling  costs ~9% bomb performance under high throughput.
var PerformanceReporter = initPerfReporter()

// should be called only once
func initPerfReporter() *Performance {
	r := new(Performance)
	r.bytesReporter = make(chan int, 512)
	r.connsReporter = make(chan int, 512)

	go func() {
		for {
			select {
			case b := <-r.bytesReporter:
				r.bytes += b
				r.BytesTotal += b
			case c := <-r.connsReporter:
				r.Conns += c
			}
		}
	}()
	go func() {
		for {
			time.Sleep(time.Second)
			r.BytesPerSec = r.bytes
			r.bytes = 0
		}
	}()

	return r
}

// bombAddress opens a TCP connection to `address`, and writes `message` repeatedly, until `stop` is closed.
// It retries with exponential backoff on network errors.
func bombAddress(message []byte, address string, offsetter Offsetter, stop chan bool, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	timeout := timeoutMin

	for {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			// this was a network error, retry!
			fmt.Printf("[net] error: %s. retrying in %s\n", err, timeout)
			time.Sleep(timeout)
			timeout *= 2
			if timeout > timeoutMax {
				timeout = timeoutMax
			}
			continue
		}

		fmt.Printf("[net] bombing %s with new connection\n", address)

		err = bombConn(message, offsetter, conn, stop)
		conn.Close()
		timeout = timeoutMin
		if err == nil {
			break // we're supposed to exit
		}
		fmt.Printf("[net] error: %s\n", err)
	}
}

// bombConn writes the given message to the given connection in a tight loop, until `stop` is closed.
// Does no transformation on the given message, so make sure packet splitting / nagle works.
func bombConn(message []byte, offsetter Offsetter, conn net.Conn, stop chan bool) error {
	PerformanceReporter.connsReporter <- 1
	defer func() { PerformanceReporter.connsReporter <- -1 }()

	var msg = make([]byte, len(message)+16) // leave some space for offset cmd
	msg = message
	// randOffset := maxOffsetX > 0 && maxOffsetY > 0

	for {
		select {
		case <-stop:
			return nil
		default:
			// if randOffset {
			msg = append(
				OffsetCmd(offsetter.Next()),
				message...,
			)
			// }
			b, err := conn.Write(msg)
			if err != nil {
				return err
			}
			if PerformanceReporter.Enabled {
				PerformanceReporter.bytesReporter <- b
			}
		}
	}
}

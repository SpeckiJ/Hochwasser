package pixelflut

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
)

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
		p.Conns, fmtBytes(p.BytesTotal), fmtBytes(p.BytesPerSec))
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

// bombAddress writes the given message via plain TCP to the given address,
// as fast as possible, until stop is closed.
func bombAddress(message []byte, address string, maxOffsetX, maxOffsetY int, stop chan bool, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	bombConn(message, maxOffsetX, maxOffsetY, conn, stop)
}

func bombConn(message []byte, maxOffsetX, maxOffsetY int, conn net.Conn, stop chan bool) {
	PerformanceReporter.connsReporter <- 1
	defer func() { PerformanceReporter.connsReporter <- -1 }()

	var msg = make([]byte, len(message)+16) // leave some space for offset cmd
	msg = message
	randOffset := maxOffsetX > 1 && maxOffsetY > 0

	for {
		select {
		case <-stop:
			return
		default:
			if randOffset {
				msg = append(
					OffsetCmd(rand.Intn(maxOffsetX), rand.Intn(maxOffsetY)),
					message...,
				)
			}
			b, err := conn.Write(msg)
			if err != nil {
				log.Fatal(err)
			}
			if PerformanceReporter.Enabled {
				PerformanceReporter.bytesReporter <- b
			}
		}
	}
}

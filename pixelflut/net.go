package pixelflut

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
)

const (
	timeoutMin = 100 * time.Millisecond
	timeoutMax = 10 * time.Second
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

// bombAddress writes the given message via plain TCP to the given address,
// as fast as possible, until stop is closed. retries with exponential backoff on network errors.
func bombAddress(message []byte, address string, maxOffsetX, maxOffsetY int, stop chan bool, wg *sync.WaitGroup) {
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

		err = bombConn(message, maxOffsetX, maxOffsetY, conn, stop)
		conn.Close()
		timeout = timeoutMin
		if err == nil {
			break // we're supposed to exit
		}
	}
}

func bombConn(message []byte, maxOffsetX, maxOffsetY int, conn net.Conn, stop chan bool) error {
	PerformanceReporter.connsReporter <- 1
	defer func() { PerformanceReporter.connsReporter <- -1 }()

	var msg = make([]byte, len(message)+16) // leave some space for offset cmd
	msg = message
	randOffset := maxOffsetX > 0 && maxOffsetY > 0

	for {
		select {
		case <-stop:
			return nil
		default:
			if randOffset {
				msg = append(
					OffsetCmd(rand.Intn(maxOffsetX), rand.Intn(maxOffsetY)),
					message...,
				)
			}
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

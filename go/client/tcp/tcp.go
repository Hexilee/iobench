package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
	"golang.org/x/sync/errgroup"
)

var (
	duration time.Duration
	port     uint64
	workers  uint64
)

func init() {
	var err error
	if dur := os.Getenv("TIME"); dur != "" {
		duration, err = time.ParseDuration(dur)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if pstr := os.Getenv("TCP_PORT"); pstr != "" {
		port, err = strconv.ParseUint(pstr, 10, 16)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if w := os.Getenv("WORKERS"); w != "" {
		workers, err = strconv.ParseUint(w, 10, 16)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func main() {
	fmt.Printf("Dialing server :%d with %d workers...\n", port, workers)
	stream, err := net.Dial("tcp", "localhost:"+strconv.Itoa(int(port)))
	if err != nil {
		log.Fatalln(err)
	}
	start := time.Now()
	var eg errgroup.Group
	bytes := new(atomic.Uint64)

	for i := 0; i < int(workers); i++ {
		eg.Go(func() error {
			defer stream.Close()
			buffer := make([]byte, 128<<10)
			for {
				n, err := stream.Read(buffer)
				if err != nil {
					return err
				}
				bytes.Add(uint64(n))
				if time.Since(start) >= duration {
					return err
				}
			}
		})
	}

	if err := eg.Wait(); err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("read %s in %s, throughput: %s/s\n", humanize.Bytes(bytes.Load()), duration, humanize.Bytes(uint64(float64(bytes.Load())/duration.Seconds())))
}

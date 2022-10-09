package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
	"golang.org/x/sync/errgroup"
)

var (
	duration     time.Duration
	workers      uint64
	randomOffset bool
)

func init() {
	var err error
	if dur := os.Getenv("TIME"); dur != "" {
		duration, err = time.ParseDuration(dur)
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

	if r := os.Getenv("RANDOM"); r != "" {
		randomOffset, err = strconv.ParseBool(r)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func main() {
	fmt.Printf("Read file to memory with %d workers...\n", workers)

	start := time.Now()
	var eg errgroup.Group
	throughputs := uint64(0)
	reads := uint64(0)
	for i := 0; i < int(workers); i++ {
		eg.Go(func() error {
			file, err := os.OpenFile("../../../data/tmp/bigdata", os.O_RDONLY, 0)
			if err != nil {
				return err
			}
			buffer := bytes.NewBuffer(make([]byte, 0))
			for {
				buffer.Reset()
				offset := uint64(0)
				if randomOffset {
					offset = rand.Uint64() % (60 * 1024 * 1024 * 1024)
				}
				_, err = file.Seek(int64(offset), 0)
				if err != nil {
					return err
				}

				n, err := io.Copy(buffer, io.LimitReader(file, 128*1024))
				if err != nil {
					return err
				}

				atomic.AddUint64(&throughputs, uint64(n))
				atomic.AddUint64(&reads, uint64(1))

				if time.Since(start) >= duration {
					return nil
				}
			}
		})
	}

	if err := eg.Wait(); err != nil {
		log.Fatalln(err)
	}

	cost := time.Since(start)

	fmt.Printf(`
Summary: %d reads in %s
  Total throughput: %s/s
`,
		reads, cost,
		humanize.IBytes(uint64(float64(throughputs)/cost.Seconds())),
	)
}

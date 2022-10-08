package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
	iorpcbench "github.com/hexilee/iobench/go/iorpc"
	"github.com/hexilee/iorpc"
	"golang.org/x/sync/errgroup"
)

type Mode string

const (
	modeWithHeaders  = "with-headers"
	modeRandomOffset = "random"
)

var (
	duration time.Duration
	port     uint64
	workers  uint64
	sessions uint64

	mode Mode
)

func init() {
	var err error
	if dur := os.Getenv("TIME"); dur != "" {
		duration, err = time.ParseDuration(dur)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if pstr := os.Getenv("IORPC_PORT"); pstr != "" {
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

	if s := os.Getenv("SESSIONS"); s != "" {
		sessions, err = strconv.ParseUint(s, 10, 16)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if m := os.Getenv("MODE"); m != "" {
		mode = Mode(m)
	}
}

func main() {
	fmt.Printf("Dialing server :%d with %d x workers(%d) sessions...\n", port, sessions, workers)
	clients := make([]*iorpc.Client, 0, workers)
	for i := 0; i < int(workers); i++ {
		clients = append(clients, iorpcbench.NewClient("localhost:"+strconv.Itoa(int(port)), 1))
	}
	start := time.Now()
	var eg errgroup.Group

	for c := range clients {
		client := clients[c]
		for i := 0; i < int(sessions); i++ {
			eg.Go(func() error {
				for {
					req := iorpc.Request{}
					if mode == modeWithHeaders {
						req.Headers = map[string]any{
							"Size":   uint64(128 * 1024),
							"Offset": uint64(0),
						}
					}

					if mode == modeRandomOffset {
						req.Headers = map[string]any{
							"Size":   uint64(128 * 1024),
							"Offset": rand.Uint64() % (60 * 1024 * 1024 * 1024),
						}
					}
					_, err := client.Call(req)
					if err != nil {
						return err
					}

					if time.Since(start) >= duration {
						return nil
					}
				}
			})
		}
	}

	if err := eg.Wait(); err != nil {
		log.Fatalln(err)
	}

	stats := make([]*iorpc.ConnStats, 0, len(clients))

	for _, client := range clients {
		stats = append(stats, client.Stats.Snapshot())
	}
	cost := time.Since(start)
	for _, client := range clients {
		client.Stop()
	}

	headBytes := uint64(0)
	bodyBytes := uint64(0)
	for _, s := range stats {
		headBytes += s.HeadRead
		bodyBytes += s.BodyRead
	}

	fmt.Printf(`
Summary:
    Read %s of head and %s of body in %s, total throughput: %s/s	
    `,
		humanize.Bytes(headBytes), humanize.Bytes(bodyBytes), cost, humanize.Bytes(uint64(float64(bodyBytes+headBytes)/cost.Seconds())),
	)
}

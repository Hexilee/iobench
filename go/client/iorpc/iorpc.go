package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
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
	go func() {
		// http://localhost:8888/debug/pprof/
		log.Println(http.ListenAndServe(":8888", nil))
	}()
	fmt.Printf("Dialing server :%d with %d x workers(%d) sessions...\n", port, sessions, workers)
	clients := make([]*iorpc.DispatchClient[iorpcbench.ReadDataHeaders, iorpc.NoopHeaders], 0, workers)
	for i := 0; i < int(workers); i++ {
		c := iorpcbench.NewClient("localhost:"+strconv.Itoa(int(port)), 1)
		c.RequestTimeout = 10 * time.Hour
		clients = append(clients, iorpc.NewDispatchClient[iorpcbench.ReadDataHeaders, iorpc.NoopHeaders](iorpcbench.ServiceReadData, c))
	}
	start := time.Now()
	var eg errgroup.Group

	ctx, cancel := context.WithCancel(context.Background())

	for c := range clients {
		client := clients[c]
		for i := 0; i < int(sessions); i++ {
			eg.Go(func() error {
				for {
					select {
					case <-ctx.Done():
						return nil
					default:
						headers := iorpcbench.ReadDataHeaders{}
						if mode == modeWithHeaders {
							headers.Size = 128 * 1024
						}

						if mode == modeRandomOffset {
							headers.Size = 128 * 1024
							headers.Offset = rand.Uint64() % (60 * 1024 * 1024 * 1024)
						}
						resp, err := client.Call(headers, 0, nil)
						if err != nil {
							return err
						}
						if resp.Body != nil {
							resp.Body.Close()
						}
					}
				}
			})
		}
	}

	time.Sleep(duration)

	stats := make([]*iorpc.ConnStats, 0, len(clients))
	for _, client := range clients {
		stats = append(stats, client.RawClient().Stats.Snapshot())
	}

	cost := time.Since(start)

	cancel()
	if err := eg.Wait(); err != nil {
		log.Fatalln(err)
	}

	for _, client := range clients {
		client.RawClient().Stop()
	}

	headBytes := uint64(0)
	bodyBytes := uint64(0)
	responseNum := uint64(0)
	failNum := uint64(0)
	for _, s := range stats {
		responseNum += s.RPCCalls
		failNum += s.ReadErrors
		headBytes += s.HeadRead
		bodyBytes += s.BodyRead
	}

	fmt.Printf(`
Summary: %d calls in %s
  Read head: %s
  Read body: %s
  Total throughput: %s/s
  Rps: %d/s
  Failures: %d
  Head/resp: %s
  Body/resp: %s 
`,
		responseNum, cost,
		humanize.IBytes(headBytes),
		humanize.IBytes(bodyBytes),
		humanize.IBytes(uint64(float64(bodyBytes+headBytes)/cost.Seconds())),
		responseNum/uint64(cost.Seconds()),
		failNum,
		humanize.IBytes(headBytes/responseNum),
		humanize.IBytes(bodyBytes/responseNum),
	)
}

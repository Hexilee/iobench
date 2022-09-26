package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
	gorpcbench "github.com/hexilee/iobench/go/gorpc"
	"github.com/valyala/gorpc"
	"golang.org/x/sync/errgroup"
)

var (
	duration time.Duration
	port     uint64
	workers  uint64
	sessions uint64
)

func init() {
	var err error
	if dur := os.Getenv("TIME"); dur != "" {
		duration, err = time.ParseDuration(dur)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if pstr := os.Getenv("GORPC_PORT"); pstr != "" {
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
}

func main() {
	fmt.Printf("Dialing server :%d with %d x workers(%d) sessions...\n", port, sessions, workers)
	clients := make([]*gorpc.DispatcherClient, 0, workers)
	for i := 0; i < int(workers); i++ {
		clients = append(clients, gorpcbench.NewClient("localhost:"+strconv.Itoa(int(port))))
	}
	start := time.Now()
	var eg errgroup.Group
	bytes := new(atomic.Uint64)

	for c := 0; c < len(clients); c++ {
		client := clients[c]
		for i := 0; i < int(sessions); i++ {
			eg.Go(func() error {
				for {
					data, err := client.Call("Get", nil)
					if err != nil {
						return err
					}
					bytes.Add(uint64(len(data.([]byte))))
					if time.Since(start) >= duration {
						return err
					}
				}
			})
		}
	}

	if err := eg.Wait(); err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("read %s in %s, throughput: %s/s\n", humanize.Bytes(bytes.Load()), duration, humanize.Bytes(uint64(float64(bytes.Load())/duration.Seconds())))
}

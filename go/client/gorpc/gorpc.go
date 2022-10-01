package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
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
	clients := make([]*gorpc.Client, 0, workers)
	for i := 0; i < int(workers); i++ {
		clients = append(clients, gorpcbench.NewClient("localhost:"+strconv.Itoa(int(port)), 1))
	}
	start := time.Now()
	var eg errgroup.Group

	for c := 0; c < len(clients); c++ {
		client := clients[c]
		for i := 0; i < int(sessions); i++ {
			eg.Go(func() error {
				for {
					_, err := client.Call(gorpc.Request{})
					if err != nil {
						return err
					}
					// if body.Body != nil {
					// 	data, err := io.ReadAll(body.Body)
					// 	if err != nil {
					// 		return err
					// 	}
					// 	bytes.Add(uint64(len(data)))
					// }
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

	bytes := uint64(0)
	for _, c := range clients {
		bytes += c.Stats.BytesRead
	}

	fmt.Printf("read %s in %s, throughput: %s/s\n", humanize.Bytes(bytes), duration, humanize.Bytes(uint64(float64(bytes)/duration.Seconds())))
}

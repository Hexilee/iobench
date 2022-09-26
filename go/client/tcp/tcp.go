package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
)

var (
	duration time.Duration
	port     uint16
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
		p, err := strconv.ParseUint(pstr, 10, 16)
		if err != nil {
			log.Fatalln(err)
		}
		port = uint16(p)
	}
}

func main() {
	stream, err := net.Dial("tcp", "localhost:"+strconv.Itoa(int(port)))
	if err != nil {
		log.Fatalln(err)
	}
	defer stream.Close()
	start := time.Now()
	buffer := make([]byte, 128<<10)
	bytes := 0
	for {
		n, err := stream.Read(buffer)
		if err != nil {
			duration = time.Since(start)
			break
		}
		bytes += n
		if elapsed := time.Since(start); elapsed > duration {
			duration = elapsed
			break
		}
	}
	fmt.Printf("read %s in %s, throughput: %s/s\n", humanize.Bytes(uint64(bytes)), duration, humanize.Bytes(uint64(float64(bytes)/duration.Seconds())))
}

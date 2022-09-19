package main

import (
	"os"
	"time"

	"github.com/dustin/go-humanize"
)

func mustLoadData() []byte {
	data, err := os.ReadFile("../data/data")
	if err != nil {
		panic(err)
	}
	return data
}

const (
	MOCK_LATENCY   = "MOCK_LATENCY"
	MOCK_BANDWIDTH = "MOCK_BANDWIDTH"
)

var (
	data             = mustLoadData()
	latency          = time.Millisecond
	bandwidth uint64 = 1 << 30

	bucket chan []byte
)

func init() {
	if v := os.Getenv(MOCK_LATENCY); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			panic(err)
		}
		latency = d
	}

	if v := os.Getenv(MOCK_BANDWIDTH); v != "" {
		b, err := humanize.ParseBytes(v)
		if err != nil {
			panic(err)
		}
		bandwidth = b
	}

	items := 1 + bandwidth*uint64(latency)/(uint64(len(data))*uint64(time.Second))
	bucket = make(chan []byte, items)
	for i := 0; i < int(items); i++ {
		bucket <- data
	}
}

func mockRead() []byte {
	d := <-bucket
	time.Sleep(latency)
	bucket <- d
	return d
}

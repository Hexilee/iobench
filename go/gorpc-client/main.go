package main

import (
	"log"

	"github.com/hexilee/iobench/go/gorpc"
)

var (
	clients  uint
	sessions uint
)

func main() {
	client := gorpc.NewClient("localhost:8003")
	data, err := client.Call("Get", nil)
	if err != nil {
		log.Fatal("call Get failed: ", err)
	}
	log.Printf("data length: %d", len(data.([]byte)))
}

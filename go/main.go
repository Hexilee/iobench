package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/dustin/go-humanize"
	"github.com/felixge/fgprof"
	"github.com/valyala/fasthttp"
)

func main() {
	fmt.Printf("Fake: latency(%s), bandwidth(%s), bucket-size(%d)\n", latency, humanize.IBytes(bandwidth), len(bucket))

	go func() {
		fmt.Println("Starting fasthttp server on :8001...")
		if err := fasthttp.ListenAndServe(":8001", fastHttpHandler); err != nil {
			log.Fatalf("Error in fasthttp ListenAndServe: %v", err)
		}
	}()

	go func() {
		fmt.Println("Starting http2 server on :8443...")
		if err := http.ListenAndServeTLS(":8443", "../output/server.crt", "../output/server.key", nil); err != nil {
			log.Fatalf("Error in ListenAndServeTLS: %v", err)
		}
	}()

	http.HandleFunc("/fast", fastHandler)
	http.HandleFunc("/slow", slowHandler)
	http.HandleFunc("/mock", mockHandler)
	http.HandleFunc("/stat/fast", statHandler(fastStat))
	http.HandleFunc("/stat/slow", statHandler(slowStat))
	http.HandleFunc("/stat/mock", statHandler(mockStat))
	http.Handle("/debug/fgprof", fgprof.Handler())

	fmt.Println("Starting http1 server on :8000...")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatalf("Error in ListenAndServe: %v", err)
	}
}

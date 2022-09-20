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
	fmt.Println("Starting server...")

	go func() {
		if err := fasthttp.ListenAndServe(":8001", fastHttpHandler); err != nil {
			log.Fatalf("Error in ListenAndServe: %v", err)
		}
	}()

	http.HandleFunc("/fast", fastHandler)
	http.HandleFunc("/slow", slowHandler)
	http.HandleFunc("/mock", mockHandler)
	http.HandleFunc("/stat/fast", statHandler(fastStat))
	http.HandleFunc("/stat/slow", statHandler(slowStat))
	http.HandleFunc("/stat/mock", statHandler(mockStat))
	http.Handle("/debug/fgprof", fgprof.Handler())
	if err := http.ListenAndServeTLS(":8000", "../output/server.crt", "../output/server.key", nil); err != nil {
		log.Fatalf("Error in ListenAndServeTLS: %v", err)
	}
}

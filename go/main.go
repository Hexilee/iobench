package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"

	"github.com/dustin/go-humanize"
	"github.com/felixge/fgprof"
	"github.com/valyala/fasthttp"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	fmt.Printf("Fake: latency(%s), bandwidth(%s), bucket-size(%d)\n", latency, humanize.IBytes(bandwidth), len(bucket))

	mux := http.NewServeMux()
	mux.HandleFunc("/fast", fastHandler)
	mux.HandleFunc("/slow", slowHandler)
	mux.HandleFunc("/mock", mockHandler)
	mux.HandleFunc("/stat/fast", statHandler(fastStat))
	mux.HandleFunc("/stat/slow", statHandler(slowStat))
	mux.HandleFunc("/stat/mock", statHandler(mockStat))
	mux.Handle("/debug/fgprof", fgprof.Handler())
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)

	go func() {
		fmt.Println("Starting fasthttp server on :8001...")
		if err := fasthttp.ListenAndServe(":8001", fastHttpHandler); err != nil {
			log.Fatalf("Error in fasthttp ListenAndServe: %v", err)
		}
	}()

	go func() {
		h2s := &http2.Server{}
		server := http.Server{
			Addr:    ":8002",
			Handler: h2c.NewHandler(mux, h2s),
		}
		fmt.Println("Starting h2c server on :8002...")
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Error in ListenAndServeTLS: %v", err)
		}
	}()

	fmt.Println("Starting http1 server on :8000...")
	server := http.Server{
		Addr:    ":8000",
		Handler: mux,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Error in ListenAndServe: %v", err)
	}
}

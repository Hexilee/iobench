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

	iorpcbench "github.com/hexilee/iobench/go/iorpc"
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
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

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
			log.Fatalf("Error in h2c ListenAndServe: %v", err)
		}
	}()

	go func() {
		server := http.Server{
			Addr:    ":8443",
			Handler: mux,
		}
		fmt.Println("Starting http2 server on :8443...")
		if err := server.ListenAndServeTLS("../output/server.crt", "../output/server.key"); err != nil {
			log.Fatalf("Error in http2 ListenAndServeTLS: %v", err)
		}
	}()

	go func() {
		fmt.Println("Starting iorpc server on :8003...")
		if err := iorpcbench.ListenAndServe(":8003"); err != nil {
			log.Fatal("iorpc server failed: ", err)
		}
	}()

	go func() {
		fmt.Println("Starting tcp server on :8004...")
		if err := NewTCPFileServer("../data/data").ListenAndServe(":8004"); err != nil {
			log.Fatal("tcp server failed: ", err)
		}
	}()

	go func() {
		fmt.Println("Starting sendfile server on :8005...")
		if err := NewSendFileServer("../data/data").ListenAndServe(":8005"); err != nil {
			log.Fatal("sendfile server failed: ", err)
		}
	}()

	go func() {
		fmt.Println("Starting splice server on :8006...")
		if err := NewSendFileServer("../data/data").ListenAndServe(":8006"); err != nil {
			log.Fatal("splice server failed: ", err)
		}
	}()

	fmt.Println("Starting http1 server on :8000...")
	server := http.Server{
		Addr:    ":8000",
		Handler: mux,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Error in http1 ListenAndServe: %v", err)
	}
}

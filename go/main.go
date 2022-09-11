package main

import (
	"net/http"
)

var data = make([]byte, 4096)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(data)
}

func main() {
	for i := range data {
		data[i] = '0'
	}
	http.HandleFunc("/", indexHandler)
	http.ListenAndServe(":8000", nil)
}

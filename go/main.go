package main

import (
	"context"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/google/uuid"
)

var (
	slowStat = NewIOStat(context.TODO(), 10000)
	fastStat = NewIOStat(context.TODO(), 100000)
)

func fastHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	data, err := os.ReadFile("../data/data.txt")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	fastStat.Collect(time.Since(start))
	w.Write(data)
}

func slowHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	data, err := func() ([]byte, error) {
		filepath := path.Join("../data/tmp/", uuid.New().String())
		file, err := os.Create(filepath)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		data, err := os.ReadFile("../data/data.txt")
		if err != nil {
			return nil, err
		}

		_, err = file.Write(data)
		if err != nil {
			return nil, err
		}
		err = file.Sync()
		if err != nil {
			return nil, err
		}
		return data, os.Remove(filepath)
	}()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	slowStat.Collect(time.Since(start))
	w.Write(data)
}

func statHandler(ioStat *IOStat) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		stat := ioStat.Stat()
		w.Write([]byte(stat.String()))
	}
}

func main() {
	http.HandleFunc("/fast", fastHandler)
	http.HandleFunc("/slow", slowHandler)
	http.HandleFunc("/stat/fast", statHandler(fastStat))
	http.HandleFunc("/stat/slow", statHandler(slowStat))
	http.ListenAndServe(":8000", nil)
}

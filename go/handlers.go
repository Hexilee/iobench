package main

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path"
	"time"

	"github.com/google/uuid"
)

var (
	slowStat = NewIOStat(context.TODO(), 10000)
	fastStat = NewIOStat(context.TODO(), 100000)
	mockStat = NewIOStat(context.TODO(), 100000)
)

func mockHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	data := mockRead()
	mockStat.Collect(time.Since(start))
	w.Write(data)
}

func fastHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	data, err := os.ReadFile("../data/data")
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
		data, err := os.ReadFile("../data/data")
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

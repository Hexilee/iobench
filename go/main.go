package main

import (
	"net/http"
	"os"
	"path"

	"github.com/google/uuid"
)

func fastHandler(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile("../data/data.txt")
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func slowHandler(w http.ResponseWriter, r *http.Request) {
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
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func main() {
	http.HandleFunc("/fast", fastHandler)
	http.HandleFunc("/slow", slowHandler)
	http.ListenAndServe(":8000", nil)
}

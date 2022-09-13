package main

import (
	"net/http"
	"os"
	"path"

	"github.com/google/uuid"
)

var data = make([]byte, 4096)

func fastHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(data)
}

func slowHandler(w http.ResponseWriter, r *http.Request) {
	err := func() error {
		filepath := path.Join("../data", uuid.New().String())
		file, err := os.Create(filepath)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = file.Write(data)
		if err != nil {
			return err
		}
		err = file.Sync()
		if err != nil {
			return err
		}
		return os.Remove(filepath)
	}()

	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func main() {
	for i := range data {
		data[i] = '0'
	}
	http.HandleFunc("/fast", fastHandler)
	http.HandleFunc("/slow", slowHandler)
	http.ListenAndServe(":8000", nil)
}

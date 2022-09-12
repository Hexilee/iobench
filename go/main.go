package main

import (
	"net/http"
	"os"

	"github.com/google/uuid"
)

var data = make([]byte, 4096)

func fastHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(data)
}

func slowHandler(w http.ResponseWriter, r *http.Request) {
	err := func() error {
		filename := uuid.New().String()
		file, err := os.Create(filename)
		if err != nil {
			return err
		}

		_, err = file.Write(data)
		if err != nil {
			return err
		}
		err = file.Sync()
		if err != nil {
			return err
		}
		err = file.Close()
		if err != nil {
			return err
		}
		return os.Remove(filename)
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

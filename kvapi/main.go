package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

var store = struct {
	sync.RWMutex
	m map[string]string
}{m: make(map[string]string)}

func Put(key string, value string) error {
	store.Lock()
	store.m[key] = value
	store.Unlock()

	return nil
}

var ErrorNoSuchKey = errors.New("no such key")

func Get(key string) (string, error) {
	store.RLock()
	value, ok := store.m[key]
	store.RUnlock()

	if !ok {
		return "", ErrorNoSuchKey
	}

	return value, nil
}

func Delete(key string) error {
	store.Lock()
	delete(store.m, key)
	store.Unlock()

	return nil
}

func helloMuxHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello gorilla/mux\n"))
}

func putHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = Put(key, string(value))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := Get(key)

	if err != nil {
		if errors.Is(err, ErrorNoSuchKey) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return

		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(value))
	return
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	err := Delete(key)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", helloMuxHandler)
	r.HandleFunc("/v1/{key}", putHandler).Methods("PUT")
	r.HandleFunc("/v1/{key}", getHandler).Methods("GET")
	r.HandleFunc("/v1/{key}", deleteHandler).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8080", r))
}

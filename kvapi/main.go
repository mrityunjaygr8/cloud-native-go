package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/mrityunjaygr8/cloud-native-go/kvapi/store"
	transactionlogger "github.com/mrityunjaygr8/cloud-native-go/kvapi/transactionLogger"

	"github.com/gorilla/mux"
)

var logger transactionlogger.TransactionLogger

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

	err = store.Put(key, string(value))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.WritePut(key, string(value))

	w.WriteHeader(http.StatusCreated)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := store.Get(key)

	if err != nil {
		if errors.Is(err, store.ErrorNoSuchKey) {
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

	err := store.Delete(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.WriteDelete(key)

	w.WriteHeader(http.StatusOK)

}

func initializeTransactionLog() error {
	var err error

	logger, err = transactionlogger.NewFileTransactionLogger("transaction.log")

	if err != nil {
		return fmt.Errorf("failed to create event logger: %w", err)
	}

	events, errors := logger.ReadEvents()

	e, ok := transactionlogger.Event{}, true

	for ok && err == nil {
		select {
		case err, ok = <-errors:
		case e, ok = <-events:
			switch e.EventType {
			case transactionlogger.EventDelete:
				err = store.Delete(e.Key)
			case transactionlogger.EventPut:
				err = store.Put(e.Key, e.Value)
			}
		}
	}

	logger.Run()

	return err
}

func main() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	defer func() {
		signal.Stop(signalChan)
		logger.Close()
	}()

	go func() {
		select {
		case <-signalChan:
			log.Println("Exiting Gracefully")
			logger.Close()
		}

		<-signalChan
		log.Println("Exiting Immediately")
		os.Exit(1)
	}()

	err := initializeTransactionLog()
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/v1/{key}", putHandler).Methods("PUT")
	r.HandleFunc("/v1/{key}", getHandler).Methods("GET")
	r.HandleFunc("/v1/{key}", deleteHandler).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8080", r))
}

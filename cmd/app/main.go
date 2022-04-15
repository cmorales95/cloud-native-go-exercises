package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/cmorales95/cloud-native-go-exercises/entities"
	transactionLogger "github.com/cmorales95/cloud-native-go-exercises/logger"
	"github.com/cmorales95/cloud-native-go-exercises/service"
	"github.com/gorilla/mux"
)

var logger transactionLogger.TransactionLogger

func initializeTransactionLog() error {
	var err error
	logger, err = transactionLogger.NewFileTransactionLogger("transaction.log")
	if err != nil {
		return fmt.Errorf("failed to crate event logger: %w", err)
	}
	events, _errors := logger.ReadEvents()
	e, ok := entities.Event{}, true

	for ok && err == nil {
		select {
		case err, ok = <-_errors:
		case e, ok = <-events:
			switch e.EventType {
			case entities.EventDelete:
				err = service.Delete(e.Key)
			case entities.EventPut:
				err = service.Put(e.Key, e.Value)
			}
		}
	}

	logger.Run()

	return err
}

func helloMuxHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello gorilla/mux!\n"))
}

// KeyValuePutHandler expects to be called with a put request for
// the "/v1/key/{key}"
func KeyValuePutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w,
			err.Error(),
			http.StatusInternalServerError)
		return
	}

	err = service.Put(key, string(value))
	if err != nil {
		http.Error(w,
			err.Error(),
			http.StatusInternalServerError)
		return
	}

	logger.WritePut(key, string(value))
	w.WriteHeader(http.StatusCreated)
}

func KeyValueGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := service.Get(key)
	if errors.Is(err, service.ErrorNoSuchKey) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(value))
}

func KeyValueDeleteHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	key := vars["key"]

	err := service.Delete(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.WriteDelete(key)
	w.WriteHeader(http.StatusOK)
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", helloMuxHandler)
	r.HandleFunc("/v1/{key}", KeyValuePutHandler).Methods(http.MethodPut)
	r.HandleFunc("/v1/{key}", KeyValueGetHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/{key}", KeyValueDeleteHandler).Methods(http.MethodDelete)
	if err := initializeTransactionLog(); err != nil {
		log.Fatalf("error initializing transaction log: %w", err)
	}

	fmt.Println()
	log.Fatal(http.ListenAndServe(":8080", r))

}

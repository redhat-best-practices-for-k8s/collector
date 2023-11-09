package main

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/collector/actions"

	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		actions.ResultsHandler(w, r)
	case http.MethodPost:
		actions.ParserHandler(w, r)
	default:
		_, writeErr := w.Write([]byte(actions.InvalidRequestErr + "\n"))
		if writeErr != nil {
			logrus.Errorf(actions.WritingResponseErr, writeErr)
		}
		logrus.Errorf(actions.InvalidRequestErr)
	}
}

func main() {
	http.HandleFunc("/", handler)
	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

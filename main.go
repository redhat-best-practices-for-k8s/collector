package main

import (
	"database/sql"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/collector/actions"

	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

func connectToDB() (*sql.DB, error) {
	DBUsername := os.Getenv("DB_USER")
	DBPassword := os.Getenv("DB_PASSWORD")
	DBURL := os.Getenv("DB_URL")
	DBPort := os.Getenv("DB_PORT")

	DBConnStr := DBUsername + ":" + DBPassword + "@tcp(" + DBURL + ":" + DBPort + ")/"
	db, err := sql.Open("mysql", DBConnStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	db, err := connectToDB()
	if err != nil {
		_, writeErr := w.Write([]byte(err.Error() + "\n"))
		if writeErr != nil {
			logrus.Errorf(actions.WritingResponseErr, writeErr)
		}
		logrus.Errorf(actions.FailedToConnectDBErr, err)
		return
	}
	defer db.Close()

	switch r.Method {
	case http.MethodGet:
		actions.ResultsHandler(w, r, db)
	case http.MethodPost:
		actions.ParserHandler(w, r, db)
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

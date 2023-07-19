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

func connectToDB() *sql.DB {
	DBUsername := os.Getenv("DB_USER")
	DBPassword := os.Getenv("DB_PASSWORD")
	DBURL := os.Getenv("DB_URL")
	DBPort := os.Getenv("DB_PORT")

	DBConnStr := DBUsername + ":" + DBPassword + "@tcp(" + DBURL + ":" + DBPort + ")/"
	db, err := sql.Open("mysql", DBConnStr)
	if err != nil {
		return nil
	}

	err = db.Ping()
	if err != nil {
		return nil
	}
	return db
}

func handler(w http.ResponseWriter, r *http.Request) {
	db := connectToDB()
	if db == nil {
		_, writeErr := w.Write([]byte(actions.FailedToConnectDBErr))
		if writeErr != nil {
			logrus.Errorln(writeErr)
		}
		logrus.Error(actions.FailedToConnectDBErr)
		return
	}
	defer db.Close()

	switch r.Method {
	case http.MethodGet:
		actions.ResultsHandler(w, db)
	case http.MethodPost:
		actions.ParserHandler(w, r, db)
	default:
		_, err := w.Write([]byte(actions.InvalidRequest))
		if err != nil {
			logrus.Errorln(err)
		}
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

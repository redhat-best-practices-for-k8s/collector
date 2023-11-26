package main

import (
	"database/sql"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/collector/actions"

	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

type CollectorApp struct {
	Database *sql.DB
}

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

func (collector *CollectorApp) handler(w http.ResponseWriter, r *http.Request) {
	logrus.Info(actions.ServerIsUpMsg)
	_, writeErr := w.Write([]byte(actions.ServerIsUpMsg + "\n"))
	if writeErr != nil {
		logrus.Errorf(actions.WritingResponseErr, writeErr)
	}

	switch r.Method {
	case http.MethodGet:
		actions.ResultsHandler(w, r, collector.Database)
	case http.MethodPost:
		actions.ParserHandler(w, r, collector.Database)
	default:
		_, writeErr := w.Write([]byte(actions.InvalidRequestErr + "\n"))
		if writeErr != nil {
			logrus.Errorf(actions.WritingResponseErr, writeErr)
		}
		logrus.Errorf(actions.InvalidRequestErr)
	}
}

func main() {
	readTimeOut, _ := strconv.Atoi(os.Getenv("SERVER_READ_TIMEOUT"))
	writeTimeOut, _ := strconv.Atoi(os.Getenv("SERVER_WRITE_TIMEOUT"))
	adrr := os.Getenv("SERVER_ADDR")

	// connect to DB
	db, _ := connectToDB()

	collector := &CollectorApp{Database: db}

	http.HandleFunc("/", collector.handler)
	server := &http.Server{
		Addr:         adrr,
		ReadTimeout:  time.Duration(readTimeOut) * time.Second,
		WriteTimeout: time.Duration(writeTimeOut) * time.Second,
	}
	err := server.ListenAndServe()
	if err != nil {
		db.Close()
		log.Fatal(err)
	}
	db.Close()
}

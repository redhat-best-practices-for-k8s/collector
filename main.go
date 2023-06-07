package main

import (
	"database/sql"
	"os"
	"time"

	"github.com/test-network-function/collector/actions"

	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

func handler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", os.Getenv("DB_CONN_STR"))
	if err != nil {
		fmt.Println(err)
	}

	if r.Method == http.MethodGet {
		actions.ResultsHandler(w, db)
	}
	if r.Method == http.MethodPost {
		actions.ParserHandler(w, r, db)
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

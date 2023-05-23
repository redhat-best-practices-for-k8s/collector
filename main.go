package main

import (
	"database/sql"

	"github.com/test-network-function/collector/actions"

	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

const DB_CONN_STR = "root:@tcp(127.0.0.1:3306)/"

func handler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", DB_CONN_STR)
	if err != nil {
		fmt.Println(err)
	}

	if r.Method == http.MethodGet {
		actions.ResultsHandler(w, r, db)
	}
	if r.Method == http.MethodPost {
		actions.ParserHandler(w, r, db)
	}
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

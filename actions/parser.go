package actions

import (
	"database/sql"
	"encoding/json"
	"time"

	"fmt"
	"io/ioutil"

	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

const CLAIM_TAG = "claim"
const VERSIONS_TAG = "versions"
const METADATA_TAG = "metadata"
const RESULTS_TAG = "results"
const CLAIM_FILE_INOUT_NAME = "claimFile"
const CREATED_BY_INPUT_NAME = "created_by"
const PARTNER_NAME_INPUT_NAME = "partner_name"

const USE_COLLECTOR_SQL_CMD = `USE cnf; `
const INSERT_TO_CLAIM_SQL_CMD = `INSERT INTO claim 
								(cnf_version, created_by, upload_time, partner_name)
								VALUES (?, ?, ?, ?);`
const INSERT_TO_CLAIM_RES_SQL_CMD = `INSERT INTO claim_result
							(claim_id, suite_name, test_id, test_status)
							VALUES (?, ?, ?, ?);`
const EXTRACT_LAST_CLAIM_ID = `SELECT id FROM cnf.claim ORDER BY id DESC LIMIT 1;`
const UPLOAD_TIME_LAYOUT = "2006-01-02T15:04:05-07:00"


func uploadAndConvertClaimFile(r *http.Request) map[string]interface{} {

	r.ParseMultipartForm(10 << 20)
	
	claimFile, _, err := r.FormFile(CLAIM_FILE_INOUT_NAME)
	if err != nil {
		fmt.Println(err)
	}
	defer claimFile.Close()

	claimFileBytes, err := ioutil.ReadAll(claimFile)
	if err != nil {
		fmt.Println(err)
	}
	var claimFileMap map[string]interface{}
	json.Unmarshal([]byte(claimFileBytes), &claimFileMap)
	return claimFileMap[CLAIM_TAG].(map[string]interface{})
}

func insertToClaimTable(r *http.Request, db *sql.DB, claimFileMap map[string]interface{}) {

	versions := claimFileMap[VERSIONS_TAG].(map[string]interface{})
	metadata := claimFileMap[METADATA_TAG].(map[string]interface{})

	// saving users input reffering to who created claim file and partner's name
	created_by := r.FormValue(CREATED_BY_INPUT_NAME)
	partner_name := r.FormValue(PARTNER_NAME_INPUT_NAME)

	// converting tests end time to time object
	datetime, err := time.Parse(UPLOAD_TIME_LAYOUT, metadata["endTime"].(string))
	if err != nil {
		fmt.Println(err)
	}

	_, err = db.Exec(INSERT_TO_CLAIM_SQL_CMD, versions["ocp"].(string),
		created_by, datetime, partner_name)
	if err != nil {
		fmt.Println(err)
	}
}

func insertToClaimResultTable(db *sql.DB, claimFileMap map[string]interface{}) {
	results := claimFileMap[RESULTS_TAG].(map[string]interface{})

	var claimId string
	err := db.QueryRow(EXTRACT_LAST_CLAIM_ID).Scan(&claimId)
	if err != nil {
		fmt.Println(err)
	}

	for testName := range results {
		testData := results[testName].([]interface{})[0].(map[string]interface{})
		testID := testData["testID"].(map[string]interface{})
		_, err = db.Exec(INSERT_TO_CLAIM_RES_SQL_CMD, claimId, testID["suite"].(string),
			testID["id"].(string), testData["state"].(string))
		if err != nil {
			fmt.Println(err)
		}
	}
}

func parseClaimFile(r *http.Request, db *sql.DB, claimFileMap map[string]interface{}) {
	_, err := db.Exec(USE_COLLECTOR_SQL_CMD)
	if err != nil {
		fmt.Println(err)
	}

	insertToClaimTable(r, db, claimFileMap)
	insertToClaimResultTable(db, claimFileMap)
}

func ParserHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	claimFileMap := uploadAndConvertClaimFile(r)
	parseClaimFile(r, db, claimFileMap)

	fmt.Fprintf(w, "File was uploaded succesfully!")
}
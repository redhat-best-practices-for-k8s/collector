package actions

import (
	"database/sql"
	"encoding/json"
	"io"
	"time"

	"fmt"

	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

func uploadAndConvertClaimFile(r *http.Request) map[string]interface{} {

	_ = r.ParseMultipartForm(10 << 20)

	claimFile, _, err := r.FormFile(CLAIM_FILE_INPUT_NAME)
	if err != nil {
		fmt.Println(err)
	}
	defer claimFile.Close()

	claimFileBytes, err := io.ReadAll(claimFile)
	if err != nil {
		fmt.Println(err)
	}
	var claimFileMap map[string]interface{}
	err = json.Unmarshal([]byte(claimFileBytes), &claimFileMap)
	if err != nil {
		fmt.Println(err)
	}
	return claimFileMap[CLAIM_TAG].(map[string]interface{})
}

func insertToClaimTable(r *http.Request, db *sql.DB, claimFileMap map[string]interface{}) {

	versions := claimFileMap[VERSIONS_TAG].(map[string]interface{})

	// saving users input referring to who created claim file and partner's name
	created_by := r.FormValue(CREATED_BY_INPUT_NAME)
	partner_name := r.FormValue(PARTNER_NAME_INPUT_NAME)

	_, err := db.Exec(INSERT_TO_CLAIM_SQL_CMD, versions["ocp"].(string),
		created_by, time.Now(), partner_name)
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

	fmt.Fprintf(w, "File was uploaded successfully!")
}

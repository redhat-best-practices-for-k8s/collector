package actions

import (
	"database/sql"
	"encoding/json"
	"io"
	"time"

	"fmt"

	"net/http"
)

func uploadAndConvertClaimFile(r *http.Request) map[string]interface{} {
	_ = r.ParseMultipartForm(ParseLowerBound << ParseUpperBound)

	claimFile, _, err := r.FormFile(ClaimFileInputName)
	if err != nil {
		fmt.Println(err)
	}
	defer claimFile.Close()

	claimFileBytes, err := io.ReadAll(claimFile)
	if err != nil {
		fmt.Println(err)
	}
	var claimFileMap map[string]interface{}
	err = json.Unmarshal(claimFileBytes, &claimFileMap)
	if err != nil {
		fmt.Println(err)
	}
	return claimFileMap[ClaimTag].(map[string]interface{})
}

func insertToClaimTable(r *http.Request, db *sql.DB, claimFileMap map[string]interface{}) {
	versions := claimFileMap[VersionsTag].(map[string]interface{})

	// saving users input referring to who created claim file and partner's name
	createdBy := r.FormValue(CreatedByInputName)
	partnerName := r.FormValue(PartnerNameInputName)

	_, err := db.Exec(InsertToClaimSQLCmd, versions["ocp"].(string),
		createdBy, time.Now(), partnerName)
	if err != nil {
		fmt.Println(err)
	}
}

func insertToClaimResultTable(db *sql.DB, claimFileMap map[string]interface{}) {
	results := claimFileMap[ResultsTag].(map[string]interface{})

	var claimID string
	err := db.QueryRow(ExtractLastClaimID).Scan(&claimID)
	if err != nil {
		fmt.Println(err)
	}

	for testName := range results {
		testData := results[testName].([]interface{})[0].(map[string]interface{})
		testID := testData["testID"].(map[string]interface{})
		_, err = db.Exec(InsertToClaimResSQLCmd, claimID, testID["suite"].(string),
			testID["id"].(string), testData["state"].(string))
		if err != nil {
			fmt.Println(err)
		}
	}
}

func parseClaimFile(r *http.Request, db *sql.DB, claimFileMap map[string]interface{}) {
	_, err := db.Exec(UseCollectorSQLCmd)
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

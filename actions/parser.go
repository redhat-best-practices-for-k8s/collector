package actions

import (
	"database/sql"
	"encoding/json"
	"io"
	"time"

	"fmt"

	"net/http"
)

func writeResponse(w http.ResponseWriter, response string) {
	_, err := w.Write([]byte(response))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(response)
}

func readClaimFile(w http.ResponseWriter, r *http.Request) []byte {
	err := r.ParseMultipartForm(ParseLowerBound << ParseUpperBound)
	if err != nil {
		writeResponse(w, err.Error())
		return nil
	}

	claimFile, _, err := r.FormFile(ClaimFileInputName)
	if err != nil {
		writeResponse(w, err.Error())
		return nil
	}
	defer claimFile.Close()

	claimFileBytes, err := io.ReadAll(claimFile)
	if err != nil {
		writeResponse(w, err.Error())
		return nil
	}

	return claimFileBytes
}

func uploadAndConvertClaimFile(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	claimFileBytes := readClaimFile(w, r)
	if claimFileBytes == nil {
		return nil
	}

	var claimFileMap map[string]interface{}
	err := json.Unmarshal(claimFileBytes, &claimFileMap)
	if err != nil {
		writeResponse(w, err.Error())
		return nil
	}

	_, keyExists := claimFileMap[ClaimTag]
	if !keyExists {
		writeResponse(w, err.Error())
		return nil
	}
	return claimFileMap[ClaimTag].(map[string]interface{})
}

func insertToClaimTable(r *http.Request, db *sql.DB, claimFileMap map[string]interface{}) bool {
	versions, keyExists := claimFileMap[VersionsTag].(map[string]interface{})
	if !keyExists {
		return false
	}

	// saving users input referring to who created claim file and partner's name
	createdBy := r.FormValue(CreatedByInputName)
	partnerName := r.FormValue(PartnerNameInputName)

	// created_by field can't be null
	if createdBy == "" {
		return false
	}

	_, keyExists = versions["ocp"]
	if !keyExists {
		return false
	}
	_, err := db.Exec(InsertToClaimSQLCmd, versions["ocp"].(string), createdBy, time.Now(), partnerName)
	if err != nil {
		fmt.Println(err)
	}
	return true
}

func validateInnerResultsKeys(results map[string]interface{}, testName string) (
	keysExists bool, testData map[string]interface{}, testID map[string]interface{}) {
	testData, keyExists := results[testName].([]interface{})[0].(map[string]interface{})
	if !keyExists {
		return false, nil, nil
	}
	testID, keyExists = testData["testID"].(map[string]interface{})
	if !keyExists {
		return false, nil, nil
	}
	_, stateKeyExists := testData["state"]
	_, suiteKeyExists := testID["suite"]
	_, idKeyExists := testID["id"]
	if !stateKeyExists || !suiteKeyExists || !idKeyExists {
		return false, nil, nil
	}
	return true, testData, testID
}

func insertToClaimResultTable(db *sql.DB, claimFileMap map[string]interface{}) bool {
	results, keyExists := claimFileMap[ResultsTag].(map[string]interface{})
	if !keyExists {
		return false
	}

	var claimID string
	err := db.QueryRow(ExtractLastClaimID).Scan(&claimID)
	if err != nil {
		fmt.Println(err)
	}

	for testName := range results {
		keysExists, testData, testID := validateInnerResultsKeys(results, testName)
		if !keysExists {
			return false
		}
		_, err = db.Exec(InsertToClaimResSQLCmd, claimID, testID["suite"].(string),
			testID["id"].(string), testData["state"].(string))
		if err != nil {
			fmt.Println(err)
		}
	}
	return true
}

func parseClaimFile(r *http.Request, db *sql.DB, claimFileMap map[string]interface{}) bool {
	_, err := db.Exec(UseCollectorSQLCmd)
	if err != nil {
		fmt.Println(err)
	}

	if insertToClaimTable(r, db, claimFileMap) && insertToClaimResultTable(db, claimFileMap) {
		return true
	}
	return false
}

func ParserHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	claimFileMap := uploadAndConvertClaimFile(w, r)
	if claimFileMap == nil {
		return
	}
	if !parseClaimFile(r, db, claimFileMap) {
		writeResponse(w, MalformedClaimFileErr)
		return
	}
	writeResponse(w, SuccessUploadingFileMSG)
}

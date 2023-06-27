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

func validateClaimKeys(w http.ResponseWriter, claimFileMap map[string]interface{}) map[string]interface{}{
	versions, keyExists := claimFileMap[VersionsTag].(map[string]interface{})
	if !keyExists {
		return nil
	}

	_, keyExists = versions["ocp"]
	if !keyExists {
		return nil
	}
	
	return versions
}

func insertToClaimTable(w http.ResponseWriter, r *http.Request, tx *sql.Tx, claimFileMap map[string]interface{}) bool {
	versions := validateClaimKeys(w, claimFileMap)

	// saving users input referring to who created claim file and partner's name
	createdBy := r.FormValue(CreatedByInputName)
	partnerName := r.FormValue(PartnerNameInputName)

	// missing fields in claim file or created_by field is null
	if versions == nil || createdBy == "" {
		writeResponse(w, MalformedClaimFileErr)
		return false
	}

	_, err := tx.Exec(InsertToClaimSQLCmd, versions["ocp"].(string), createdBy, time.Now(), partnerName)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return false
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

func insertToClaimResultTable(w http.ResponseWriter, tx *sql.Tx, claimFileMap map[string]interface{}) bool {
	results, keyExists := claimFileMap[ResultsTag].(map[string]interface{})
	if !keyExists {
		writeResponse(w, MalformedClaimFileErr)
		return false
	}

	var claimID string
	err := tx.QueryRow(ExtractLastClaimID).Scan(&claimID)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return false
	}

	for testName := range results {
		keysExists, testData, testID := validateInnerResultsKeys(results, testName)
		if !keysExists {
			writeResponse(w, MalformedClaimFileErr)
			return false
		}
		_, err = tx.Exec(InsertToClaimResSQLCmd, claimID, testID["suite"].(string),
			testID["id"].(string), testData["state"].(string))
		if err != nil {
			tx.Rollback()
			fmt.Println(err)
			return false
		}
	}
	return true
}

func parseClaimFile(w http.ResponseWriter, r *http.Request, tx *sql.Tx, claimFileMap map[string]interface{}) bool {
	_, err := tx.Exec(UseCollectorSQLCmd)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return false
	}

	if insertToClaimTable(w, r, tx, claimFileMap) && insertToClaimResultTable(w, tx, claimFileMap) {
		return true
	}
	return false
}

func ParserHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	claimFileMap := uploadAndConvertClaimFile(w, r)
	if claimFileMap == nil {
		return
	}
	tx, err := db.Begin()
	if err != nil {
		fmt.Println(err)
		return
	}
	if !parseClaimFile(w, r, tx, claimFileMap) {
		return
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return
	}
	writeResponse(w, SuccessUploadingFileMSG)
}

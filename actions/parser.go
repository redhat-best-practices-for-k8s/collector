package actions

import (
	"database/sql"
	"encoding/json"
	"io"
	"time"

	"fmt"

	"net/http"
)

func uploadAndConvertClaimFile(w http.ResponseWriter, r *http.Request) map[string]interface{} {
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
		_, writeErr := w.Write([]byte(MalformedJSONFileErr))
		if writeErr != nil {
			fmt.Println(writeErr)
		}
		fmt.Println(err)
		return nil
	}

	_, keyExists := claimFileMap[ClaimTag]
	if !keyExists {
		_, writeErr := w.Write([]byte(MalformedClaimFileErr))
		if writeErr != nil {
			fmt.Println(writeErr)
		}
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

	_, keyExists = versions["ocp"]
	if !keyExists {
		return false
	}
	_, err := db.Exec(InsertToClaimSQLCmd, versions["ocp"].(string),
		createdBy, time.Now(), partnerName)
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
		_, writeErr := w.Write([]byte(MalformedClaimFileErr))
		if writeErr != nil {
			fmt.Println(writeErr)
		}
		return
	}
	_, writeErr := w.Write([]byte(SuccessUploadingFileMSG))
	if writeErr != nil {
		fmt.Println(writeErr)
	}
}

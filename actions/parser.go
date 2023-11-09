package actions

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"net/http"

	"github.com/sirupsen/logrus"
)

func handleTransactionRollback(tx *sql.Tx, context string, err error) {
	txErr := tx.Rollback()
	if txErr != nil {
		logrus.Errorf(RollbackErr, txErr)
	}
	logrus.Errorf(context, err)
}

func writeError(w http.ResponseWriter, context, err string) {
	_, writeErr := w.Write([]byte(err + "\n"))
	if writeErr != nil {
		logrus.Errorf(WritingResponseErr, writeErr)
	}
	logrus.Errorf(context, err)
}

func readClaimFile(w http.ResponseWriter, r *http.Request) []byte {
	err := r.ParseMultipartForm(ParseLowerBound << ParseUpperBound)
	if err != nil {
		writeError(w, RequestContentTypeErr, err.Error())
		return nil
	}

	claimFile, _, err := r.FormFile(ClaimFileInputName)
	if err != nil {
		writeError(w, FormFileErr, err.Error())
		return nil
	}
	defer claimFile.Close()

	claimFileBytes, err := io.ReadAll(claimFile)
	if err != nil {
		writeError(w, ReadingFileErr, err.Error())
		return nil
	}

	return claimFileBytes
}

func uploadAndConvertClaimFile(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	claimFileBytes := readClaimFile(w, r)
	if claimFileBytes == nil {
		// error occurred while reading claim file
		return nil
	}

	var claimFileMap map[string]interface{}
	err := json.Unmarshal(claimFileBytes, &claimFileMap)
	if err != nil {
		writeError(w, UnmarshalErr, err.Error())
		return nil
	}

	_, keyExists := claimFileMap[ClaimTag]
	if !keyExists {
		writeError(w, MalformedClaimFileErr, ClaimFieldMissingErr)
		return nil
	}
	return claimFileMap[ClaimTag].(map[string]interface{})
}

func validateClaimKeys(w http.ResponseWriter, claimFileMap map[string]interface{}) map[string]interface{} {
	versions, keyExists := claimFileMap[VersionsTag].(map[string]interface{})
	if !keyExists {
		writeError(w, MalformedClaimFileErr, VersionsFieldMissingErr)
		return nil
	}

	_, keyExists = versions["ocp"]
	if !keyExists {
		writeError(w, MalformedClaimFileErr, OcpFieldMissingErr)
		return nil
	}

	return versions
}

func insertToClaimTable(w http.ResponseWriter, r *http.Request, tx *sql.Tx, claimFileMap map[string]interface{}, partnerName string) bool {
	versions := validateClaimKeys(w, claimFileMap)
	if versions == nil {
		return false
	}

	// saving users input referring to who executed claim file and partner's name
	executedBy := r.FormValue(ExecutedByInputName)

	if executedBy == "" {
		writeError(w, "%s", ExecutedByMissingErr)
		return false
	}

	_, err := tx.Exec(InsertToClaimSQLCmd, versions["ocp"].(string), executedBy, time.Now(), partnerName)
	if err != nil {
		handleTransactionRollback(tx, ExecQueryErr, err)
		return false
	}
	return true
}

func validateInnerResultsKeys(results map[string]interface{}, testName string) (
	testData map[string]interface{}, testID map[string]interface{}, err string) {
	testData, _ = results[testName].([]interface{})[0].(map[string]interface{})

	testID, keyExists := testData["testID"].(map[string]interface{})
	if !keyExists {
		return nil, nil, fmt.Sprintf(TestTestIDMissingErr, testName)
	}

	_, stateKeyExists := testData["state"]
	if !stateKeyExists {
		return nil, nil, fmt.Sprintf(TestStateMissingErr, testName)
	}

	_, suiteKeyExists := testID["suite"]
	if !suiteKeyExists {
		return nil, nil, fmt.Sprintf(TestIDSuiteMissingErr, testName)
	}

	_, idKeyExists := testID["id"]
	if !idKeyExists {
		return nil, nil, fmt.Sprintf(TestIDIDMissingErr, testName)
	}
	return testData, testID, ""
}

func insertToClaimResultTable(w http.ResponseWriter, tx *sql.Tx, claimFileMap map[string]interface{}) bool {
	results, keyExists := claimFileMap[ResultsTag].(map[string]interface{})
	if !keyExists {
		writeError(w, MalformedClaimFileErr, ResultsFieldMissingErr)
		return false
	}

	var claimID string
	err := tx.QueryRow(ExtractLastClaimID).Scan(&claimID)
	if err != nil {
		handleTransactionRollback(tx, ScanDBFieldErr, err)
		return false
	}

	for testName := range results {
		testData, testID, keyErr := validateInnerResultsKeys(results, testName)
		if keyErr != "" {
			writeError(w, MalformedClaimFileErr, keyErr)
			return false
		}
		_, err = tx.Exec(InsertToClaimResSQLCmd, claimID, testID["suite"].(string),
			testID["id"].(string), testData["state"].(string))
		if err != nil {
			handleTransactionRollback(tx, ExecQueryErr, err)
			return false
		}
	}
	return true
}

func parseClaimFile(w http.ResponseWriter, r *http.Request, tx *sql.Tx, claimFileMap map[string]interface{}, partnerName string) bool {
	_, err := tx.Exec(UseCollectorSQLCmd)
	if err != nil {
		handleTransactionRollback(tx, ExecQueryErr, err)
		return false
	}

	if insertToClaimTable(w, r, tx, claimFileMap, partnerName) && insertToClaimResultTable(w, tx, claimFileMap) {
		return true
	}
	return false
}

func ParserHandler(w http.ResponseWriter, r *http.Request) {
	// connect to DB
	db, err := connectToDB()
	if err != nil {
		writeError(w, FailedToConnectDBErr, err.Error())
		return
	}
	defer db.Close()

	// Beginning the transaction.
	tx, err := db.Begin()
	if err != nil {
		logrus.Errorf(BeginTxErr, err)
		return
	}

	// Authonticate partner name and password
	partnerName, err := authenticatePostRequest(r, tx)
	if err != nil {
		writeError(w, AuthError, err.Error())
		return
	}

	claimFileMap := uploadAndConvertClaimFile(w, r)
	if claimFileMap == nil {
		// error occurred while uploading\converting claim file.
		return
	}

	// Check if an error occurred while parsing (which caused a Rollback).
	if !parseClaimFile(w, r, tx, claimFileMap, partnerName) {
		return
	}

	// If no error occurred, commit the transaction to make database changes.
	err = tx.Commit()
	if err != nil {
		handleTransactionRollback(tx, CommitTxErr, err)
		return
	}

	// Succfully uploaded file
	_, writeErr := w.Write([]byte(SuccessUploadingFileMSG + "\n"))
	if writeErr != nil {
		logrus.Errorf(WritingResponseErr, writeErr)
	}
	logrus.Info(SuccessUploadingFileMSG)
}

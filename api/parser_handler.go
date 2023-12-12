package api

import (
	"net/http"
	"time"

	"database/sql"
	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/collector/storage"
	"github.com/test-network-function/collector/util"
)

func ParserHandler(w http.ResponseWriter, r *http.Request, storage *storage.MySqlStorage) {
	db := storage.MySql
	defer db.Close()

	// 1. Validate the request (includes validation of the claim file format)
	claimFileMap, params, isValid := validatePostRequest(w, r)
	if !isValid {
		return
	}

	// Valid parameters
	partnerName := params[0]
	decodedPassword := params[1]
	ocpVersion := params[2]
	executedBy := params[3]

	// 2. Begin transaction to make entries into the table

	tx, err := db.Begin()
	if err != nil {
		logrus.Errorf(util.BeginTxErr, err)
		return
	}

	// 3. Validate partner's credentials, for non-existent partner create an entry in the database

	err = CreateCredentialsIfNotExists(partnerName, decodedPassword, tx)
	if err != nil {
		util.WriteError(w, util.AuthError, err.Error())
		return
	}

	// Check if an error occurred while parsing (which caused a Rollback).
	if !parseClaimFile(w, r, tx, claimFileMap, partnerName, ocpVersion, executedBy) {
		return
	}

	// If no error occurred, commit the transaction to make database changes.
	err = tx.Commit()
	if err != nil {
		util.HandleTransactionRollback(tx, util.CommitTxErr, err)
		return
	}
	/*
	   // Successfully write to S3
	   claimFile := getClaimFile(w, r)
	   uploadFileToS3(claimFile, partnerName)
	   // Succfully uploaded file
	   _, writeErr := w.Write([]byte(SuccessUploadingFileMSG + "\n"))

	   	if writeErr != nil {
	   		logrus.Errorf(WritingResponseErr, writeErr)
	   	}

	   logrus.Info(SuccessUploadingFileMSG)
	*/
}

// Done
func uploadAndConvertClaimFile(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	claimFileBytes := util.ReadClaimFile(w, r)
	if claimFileBytes == nil {
		// error occurred while reading claim file
		return nil
	}

	var claimFileMap map[string]interface{}
	err := json.Unmarshal(claimFileBytes, &claimFileMap)
	if err != nil {
		util.WriteError(w, util.UnmarshalErr, err.Error())
		return nil
	}

	_, keyExists := claimFileMap[util.ClaimTag]
	if !keyExists {
		util.WriteError(w, util.MalformedClaimFileErr, util.ClaimFieldMissingErr)
		return nil
	}
	return claimFileMap[util.ClaimTag].(map[string]interface{})
}

// Done
func validateClaimKeys(w http.ResponseWriter, claimFileMap map[string]interface{}) map[string]interface{} {
	versions, keyExists := claimFileMap[util.VersionsTag].(map[string]interface{})
	if !keyExists {
		util.WriteError(w, util.MalformedClaimFileErr, util.VersionsFieldMissingErr)
		return nil
	}

	_, keyExists = versions["ocp"]
	if !keyExists {
		util.WriteError(w, util.MalformedClaimFileErr, util.OcpFieldMissingErr)
		return nil
	}

	return versions
}

// Bad naming
func parseClaimFile(w http.ResponseWriter, r *http.Request, tx *sql.Tx, claimFileMap map[string]interface{}, partnerName, executedBy, ocpVersion string) bool {
	_, err := tx.Exec(util.UseCollectorSQLCmd)
	if err != nil {
		util.HandleTransactionRollback(tx, util.ExecQueryErr, err)
		return false
	}

	if insertToClaimTable(w, tx, claimFileMap, partnerName, executedBy, ocpVersion) && insertToClaimResultTable(w, tx, claimFileMap) {
		return true
	}
	return false
}

// Done
func insertToClaimTable(w http.ResponseWriter, tx *sql.Tx, claimFileMap map[string]interface{}, partnerName, executedBy, ocpVersion string) bool {

	_, err := tx.Exec(util.InsertToClaimSQLCmd, ocpVersion, executedBy, time.Now(), partnerName)
	if err != nil {
		util.HandleTransactionRollback(tx, util.ExecQueryErr, err)
		return false
	}
	return true
}

func insertToClaimResultTable(w http.ResponseWriter, tx *sql.Tx, claimFileMap map[string]interface{}) bool {
	results, keyExists := claimFileMap[util.ResultsTag].(map[string]interface{})
	if !keyExists {
		util.WriteError(w, util.MalformedClaimFileErr, util.ResultsFieldMissingErr)
		return false
	}

	var claimID string
	err := tx.QueryRow(util.ExtractLastClaimID).Scan(&claimID)
	if err != nil {
		util.HandleTransactionRollback(tx, util.ScanDBFieldErr, err)
		return false
	}

	for testName := range results {
		testData, testID, keyErr := validateInnerResultsKeys(results, testName)
		if keyErr != "" {
			util.WriteError(w, util.MalformedClaimFileErr, keyErr)
			return false
		}
		_, err = tx.Exec(util.InsertToClaimResSQLCmd, claimID, testID["suite"].(string),
			testID["id"].(string), testData["state"].(string))
		if err != nil {
			util.HandleTransactionRollback(tx, util.ExecQueryErr, err)
			return false
		}
	}
	return true
}

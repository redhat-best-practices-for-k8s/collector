package util

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/collector/types"
)

func HandleTransactionRollback(tx *sql.Tx) {
	txErr := tx.Rollback()
	if txErr != nil {
		logrus.Errorf(RollbackErr, txErr)
	}
}

func GetEntireCollectorTable(db *sql.DB) (claimRows, claimResultsRows *sql.Rows) {
	claimRows, err := db.Query(SelectAllFromClaim)
	if err != nil {
		logrus.Errorf(ExecQueryErr, err)
	}

	claimResultsRows, err = db.Query(SelectAllFromClaimResult)
	if err != nil {
		logrus.Errorf(ExecQueryErr, err)
	}

	return claimRows, claimResultsRows
}

func GetCollectorTablesByPartner(db *sql.DB, partnerName string) (claimRows, claimResultsRows *sql.Rows) {
	claimRows, err := db.Query(SelectAllFromClaimByPartner, partnerName)
	if err != nil {
		logrus.Errorf(ExecQueryErr, err)
	}

	// Extract claim IDs of given partner
	claimIDsRows, err := db.Query(SelectAllClaimIDsByPartner, partnerName)
	if err != nil {
		logrus.Errorf(ExecQueryErr, err)
	}
	defer claimIDsRows.Close()

	var claimIDsList []string
	for claimIDsRows.Next() {
		var claimID string
		claimIDErr := claimIDsRows.Scan(&claimID)
		if err != nil {
			logrus.Errorf(ScanDBFieldErr, claimIDErr)
		}
		claimIDsList = append(claimIDsList, claimID)
	}

	// Extract claim results of found claim IDs
	claimResultsQuery := fmt.Sprintf(SelectAllFromClaimResultByClaimIDs, strings.Join(claimIDsList, ","))
	claimResultsRows, err = db.Query(claimResultsQuery)
	if err != nil {
		logrus.Errorf(ExecQueryErr, err)
	}

	return claimRows, claimResultsRows
}

// This function stores the claim and claim result into the database in a transaction
func StoreClaimFileInDatabase(db *sql.DB, claimResult []types.ClaimResult, partnerName, executedBy, ocpVersion, s3FileKey string) error {
	// Begin transaction here
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(UseCollectorSQLCmd)
	if err != nil {
		return err
	}

	// store claim
	claimID, err := storeClaimIntoDatabase(partnerName, executedBy, ocpVersion, s3FileKey, tx)
	if err != nil {
		HandleTransactionRollback(tx)
		return err
	}

	err = storeClaimResultIntoDatabase(claimResult, claimID, tx)
	if err != nil {
		HandleTransactionRollback(tx)
		return err
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		HandleTransactionRollback(tx)
		return err
	}
	logrus.Info("Claim file is entirely stored into the database.")

	return nil
}

func storeClaimResultIntoDatabase(claimResults []types.ClaimResult, claimID int64, tx *sql.Tx) error {
	for _, cr := range claimResults {
		_, err := tx.Exec(InsertToClaimResSQLCmd, claimID, cr.SuiteName, cr.TestID, cr.TestStatus)
		if err != nil {
			return err
		}
	}
	logrus.Info(FileStoredIntoClaimResultTableSuccessfully)
	return nil
}

// Inserts into claim table and returns the id
func storeClaimIntoDatabase(partnerName, executedBy, ocpVersion, s3FileKey string, tx *sql.Tx) (claimID int64, err error) {
	result, err := tx.Exec(InsertToClaimSQLCmd, ocpVersion, executedBy, time.Now(), partnerName, s3FileKey)
	if err != nil {
		return -1, err
	}
	logrus.Info(FileStoredIntoClaimTableSuccessfully)
	claimID, err = result.LastInsertId()
	if err != nil {
		return -1, err
	}
	return claimID, nil
}

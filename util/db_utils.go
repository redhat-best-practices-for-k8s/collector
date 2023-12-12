package util

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

func HandleTransactionRollback(tx *sql.Tx, context string, err error) {
	txErr := tx.Rollback()
	if txErr != nil {
		logrus.Errorf(RollbackErr, txErr)
	}
	logrus.Errorf(context, err)
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

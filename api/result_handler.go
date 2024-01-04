package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/collector/storage"
	"github.com/test-network-function/collector/types"
	"github.com/test-network-function/collector/util"
)

func ResultsHandler(w http.ResponseWriter, r *http.Request, mysqlStorage *storage.MySQLStorage) {
	logrus.Info("Handling the GET request")
	db := mysqlStorage.MySQL
	partnerName, isValid := validateGetRequest(w, r, db)

	if !isValid {
		logrus.Error("Authentication is not successful. No result returned")
		return
	}

	collector := processResults(partnerName, db)

	printCollectorJSONFile(w, collector)
}

func processResults(partnerName string, db *sql.DB) []types.ClaimCollector {
	var claimRows, claimResultsRows *sql.Rows
	if partnerName == util.AdminUserName {
		claimRows, claimResultsRows = util.GetEntireCollectorTable(db)
	} else {
		claimRows, claimResultsRows = util.GetCollectorTablesByPartner(db, partnerName)
	}
	defer claimRows.Close()
	defer claimResultsRows.Close()

	claims := mapClaimsToStruct(claimRows)
	claimResults := mapClaimResultsToStruct(claimResultsRows)
	collector := combineClaimAndResultsToStruct(claims, claimResults)
	return collector
}

func printCollectorJSONFile(w http.ResponseWriter, collector []types.ClaimCollector) {
	claimFileJSON, err := json.MarshalIndent(collector, "", "	")
	if err != nil {
		logrus.Errorf(util.MarshalErr, err)
	}
	_, err = w.Write(append(claimFileJSON, '\n'))
	if err != nil {
		logrus.Errorf(util.WritingResponseErr, err)
	}
}

func mapClaimsToStruct(claimRows *sql.Rows) []types.Claim {
	var claims []types.Claim
	for claimRows.Next() {
		var row types.Claim
		err := claimRows.Scan(&row.ID, &row.CnfVersion, &row.ExecutedBy, &row.UploadTime, &row.PartnerName, &row.MarkForDelete)
		if err != nil {
			logrus.Errorf(util.ScanDBFieldErr, err)
		}
		claims = append(claims, row)
	}
	return claims
}

func mapClaimResultsToStruct(claimResultsRows *sql.Rows) []types.ClaimResult {
	var claimResults []types.ClaimResult
	for claimResultsRows.Next() {
		var row types.ClaimResult
		err := claimResultsRows.Scan(&row.ID, &row.ClaimID, &row.SuiteName, &row.TestID, &row.TesStatus)
		if err != nil {
			logrus.Errorf(util.ScanDBFieldErr, err)
		}
		claimResults = append(claimResults, row)
	}
	return claimResults
}

func combineClaimAndResultsToStruct(claims []types.Claim, claimResults []types.ClaimResult) []types.ClaimCollector {
	var collector []types.ClaimCollector
	for _, claim := range claims {
		var curClaim types.ClaimCollector
		curClaim.Claim = claim
		for _, res := range claimResults {
			if res.ClaimID == claim.ID {
				curClaim.ClaimResults = append(curClaim.ClaimResults, res)
			}
		}
		collector = append(collector, curClaim)
	}
	return collector
}

package actions

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"net/http"

	"github.com/sirupsen/logrus"
)

type Claim struct {
	ID            int    `json:"id"`
	CnfVersion    string `json:"cnf_version"`
	ExecutedBy    string `json:"executed_by"`
	UploadTime    string `json:"upload_time"`
	PartnerName   string `json:"partner_name"`
	MarkForDelete bool   `json:"mark_for_delete"`
}

type ClaimResults struct {
	ID        int    `json:"id"`
	ClaimID   int    `json:"claim_id"`
	SuiteName string `json:"suite_name"`
	TestID    string `json:"test_id"`
	TesStatus string `json:"test_status"`
}

type ClaimCollector struct {
	Claim        Claim
	ClaimResults []ClaimResults
}

func getEntireCollectorTable(db *sql.DB) (claimRows, claimResultsRows *sql.Rows) {
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

func getCollectorTablesByPartner(db *sql.DB, partnerName string) (claimRows, claimResultsRows *sql.Rows) {
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

func mapClaimsToStruct(claimRows *sql.Rows) []Claim {
	var claims []Claim
	for claimRows.Next() {
		var row Claim
		err := claimRows.Scan(&row.ID, &row.CnfVersion, &row.ExecutedBy, &row.UploadTime, &row.PartnerName, &row.MarkForDelete)
		if err != nil {
			logrus.Errorf(ScanDBFieldErr, err)
		}
		claims = append(claims, row)
	}
	return claims
}

func mapClaimResultsToStruct(claimResultsRows *sql.Rows) []ClaimResults {
	var claimResults []ClaimResults
	for claimResultsRows.Next() {
		var row ClaimResults
		err := claimResultsRows.Scan(&row.ID, &row.ClaimID, &row.SuiteName, &row.TestID, &row.TesStatus)
		if err != nil {
			logrus.Errorf(ScanDBFieldErr, err)
		}
		claimResults = append(claimResults, row)
	}
	return claimResults
}

func combineClaimAndResultsToStruct(claims []Claim, claimResults []ClaimResults) []ClaimCollector {
	var collector []ClaimCollector
	for _, claim := range claims {
		var curClaim ClaimCollector
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

func printCollectorJSONFile(w http.ResponseWriter, collector []ClaimCollector) {
	claimFileJSON, err := json.MarshalIndent(collector, "", "	")
	if err != nil {
		logrus.Errorf(MarshalErr, err)
	}
	_, err = w.Write(append(claimFileJSON, '\n'))
	if err != nil {
		logrus.Errorf(WritingResponseErr, err)
	}
}

func ResultsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	partnerName, err := authenticateGetRequest(r, db)
	if err != nil {
		// authentication failed
		_, err = w.Write([]byte(err.Error() + "\n"))
		if err != nil {
			logrus.Errorf(WritingResponseErr, err)
		}
		return
	}
	if partnerName == "" {
		// partner name and password were not given
		return
	}

	var claimRows, claimResultsRows *sql.Rows
	if partnerName == AdminUserName {
		claimRows, claimResultsRows = getEntireCollectorTable(db)
	} else {
		claimRows, claimResultsRows = getCollectorTablesByPartner(db, partnerName)
	}
	defer claimRows.Close()
	defer claimResultsRows.Close()

	claims := mapClaimsToStruct(claimRows)
	claimResults := mapClaimResultsToStruct(claimResultsRows)
	collector := combineClaimAndResultsToStruct(claims, claimResults)
	printCollectorJSONFile(w, collector)
}

package actions

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"net/http"
)

type Claim struct {
	ID            int    `json:"id"`
	CnfVersion    string `json:"cnf_version"`
	CreatedBy     string `json:"created_by"`
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

func getCollectorTables(db *sql.DB) (claimRows, claimResultsRows *sql.Rows) {
	claimRows, err := db.Query(SelectAllFromClaim)
	if err != nil {
		fmt.Println(err)
	}

	claimResultsRows, err = db.Query(SelectAllFromClaimResult)
	if err != nil {
		fmt.Println(err)
	}

	return claimRows, claimResultsRows
}

func mapClaimsToStruct(claimRows *sql.Rows) []Claim {
	var claims []Claim
	for claimRows.Next() {
		var row Claim
		err := claimRows.Scan(&row.ID, &row.CnfVersion, &row.CreatedBy, &row.UploadTime, &row.PartnerName, &row.MarkForDelete)
		if err != nil {
			fmt.Println(err)
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
			fmt.Println(err)
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
		fmt.Println(err)
	}
	_, err = w.Write(claimFileJSON)
	if err != nil {
		fmt.Println(err)
	}
}

func ResultsHandler(w http.ResponseWriter, db *sql.DB) {
	claimRows, claimResultsRows := getCollectorTables(db)
	defer claimRows.Close()
	defer claimResultsRows.Close()

	claims := mapClaimsToStruct(claimRows)
	claimResults := mapClaimResultsToStruct(claimResultsRows)
	collector := combineClaimAndResultsToStruct(claims, claimResults)
	printCollectorJSONFile(w, collector)
}

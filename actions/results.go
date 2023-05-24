package actions

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"

	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
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
	ClaimId   int    `json:"claim_id"`
	SuiteName string `json:"suite_name"`
	TestId    string `json:"test_id"`
	TesStatus string `json:"test_status"`
}

type ClaimCollector struct {
	Claim        Claim
	ClaimResults []ClaimResults
}

func getCollectorTables(db *sql.DB) (*sql.Rows, *sql.Rows) {
	claimRows, err := db.Query(SELECT_ALL_FROM_CLAIM)
	if err != nil {
		fmt.Println(err)
	}

	claimResultsRows, err := db.Query(SELECT_ALL_FROM_CLAIM_RESULT)
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
		err := claimResultsRows.Scan(&row.ID, &row.ClaimId, &row.SuiteName, &row.TestId, &row.TesStatus)
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
			if res.ClaimId == claim.ID {
				curClaim.ClaimResults = append(curClaim.ClaimResults, res)
			}
		}
		collector = append(collector, curClaim)
	}
	return collector
}

func createCollectorJsonFile(collector []ClaimCollector) {
	claimFile, err := json.MarshalIndent(collector, "", "	")
	if err != nil {
		fmt.Println(err)
	}

	file, err := os.Create(RESULT_JSON_PATH)
	if err != nil {
		fmt.Println(err)
	}

	_, err = file.Write(claimFile)
	if err != nil {
		fmt.Println(err)
	}
}

func printCollectorJsonFile(w http.ResponseWriter) {
	file, err := os.Open(RESULT_JSON_PATH)
	if err != nil {
		fmt.Println(err)
	}

	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Fprint(w, string(content))
}

func ResultsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	claimRows, claimResultsRows := getCollectorTables(db)
	defer claimRows.Close()
	defer claimResultsRows.Close()

	claims := mapClaimsToStruct(claimRows)
	claimResults := mapClaimResultsToStruct(claimResultsRows)
	collector := combineClaimAndResultsToStruct(claims, claimResults)
	createCollectorJsonFile(collector)
	printCollectorJsonFile(w)
}

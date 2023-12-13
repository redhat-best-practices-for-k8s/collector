package types

import "database/sql"

type Claim struct {
	ID            int    `json:"id"`
	CnfVersion    string `json:"cnf_version"`
	ExecutedBy    string `json:"executed_by"`
	UploadTime    string `json:"upload_time"`
	PartnerName   string `json:"partner_name"`
	MarkForDelete bool   `json:"mark_for_delete"`
}

type ClaimResult struct {
	ID        int    `json:"id"`
	ClaimID   int    `json:"claim_id"`
	SuiteName string `json:"suite_name"`
	TestID    string `json:"test_id"`
	TesStatus string `json:"test_status"`
}

type ClaimCollector struct {
	Claim        Claim
	ClaimResults []ClaimResult
}

type CollectorApp struct {
	Database *sql.DB
}

package actions

const FailedToConnectDBErr = "Unable to connect database.\n"
const InvalidRequest = "Invalid request.\n"
const InvalidRequestContentType = "request Content-Type isn't multipart/form-data.\n"

// parser.go constants
const ClaimTag = "claim"
const VersionsTag = "versions"
const ResultsTag = "results"
const ClaimFileInputName = "claimFile"
const CreatedByInputName = "created_by"
const PartnerNameInputName = "partner_name"

const UseCollectorSQLCmd = `USE cnf; `
const InsertToClaimSQLCmd = `INSERT INTO claim 
								(cnf_version, created_by, upload_time, partner_name)
								VALUES (?, ?, ?, ?);`
const InsertToClaimResSQLCmd = `INSERT INTO claim_result
							(claim_id, suite_name, test_id, test_status)
							VALUES (?, ?, ?, ?);`
const ExtractLastClaimID = `SELECT id FROM cnf.claim ORDER BY id DESC LIMIT 1;`
const ParseLowerBound = 10
const ParseUpperBound = 20

const SuccessUploadingFileMSG = "File was uploaded successfully!\n"
const MalformedClaimFileErr = "Malformed claim file: missing field.\n"
const MalformedJSONFileErr = "Malformed json file.\n"

// results.go constants
const SelectAllFromClaim = "SELECT * FROM cnf.claim"
const SelectAllFromClaimResult = "SELECT * FROM cnf.claim_result"

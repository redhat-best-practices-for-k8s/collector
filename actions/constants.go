package actions

const FailedToConnectDBErr = "Error found while trying to connect db: %s"
const InvalidRequestErr = "Invalid request."
const RequestContentTypeErr = "Error found while Parsing multipart form: %s"
const WritingResponseErr = "Error found while writing response: %s"
const FormFileErr = "Error found while forming file: %s"
const ReadingFileErr = "Error found while reading claim file: %s"
const UnmarshalErr = "Error found while trying to unmarshal claim file: %s"
const MarshalErr = "Error found while marshaling claim file: %s"
const MalformedClaimFileErr = "Malformed claim file: "
const ClaimFieldMissingErr = MalformedClaimFileErr + "claim field is missing."
const VersionsFieldMissingErr = MalformedClaimFileErr + "versions field is missing."
const OcpFieldMissingErr = MalformedClaimFileErr + "ocp subfield of versions field is missing."
const TestMissingErr = MalformedClaimFileErr + "%s subfield of results field is missing."
const TestTestIDMissingErr = MalformedClaimFileErr + "testID subfield of %s test is missing."
const TestStateMissingErr = MalformedClaimFileErr + "state subfield of %s test is missing."
const TestIDSuiteMissingErr = MalformedClaimFileErr + "suite subfield of %s's testID field is missing."
const TestIDIDMissingErr = MalformedClaimFileErr + "id subfield of %s's testID field is missing."
const ResultsFieldMissingErr = MalformedClaimFileErr + "results field is missing."
const CreatedByMissingErr = MalformedClaimFileErr + "created by value is missing."
const MalformedJSONFileErr = "Malformed json file."
const RollbackErr = "Error found while Rollbacking transaction: %s"
const ExecQueryErr = "Error found while executing a mysql query: %s"
const ScanDBFieldErr = "Error found while scanning db field: %s"
const BeginTxErr = "Error found while beginning transaction: %s"
const CommitTxErr = "Error found while committing transaction: %s"

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

const SuccessUploadingFileMSG = "File was uploaded successfully!"

// results.go constants
const SelectAllFromClaim = "SELECT * FROM cnf.claim"
const SelectAllFromClaimResult = "SELECT * FROM cnf.claim_result"

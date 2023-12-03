package actions

const FailedToConnectDBErr = "Error found while trying to connect db: %s"
const InvalidRequestErr = "Invalid request."
const RequestContentTypeErr = "Error found while Parsing multipart form: %s"
const WritingResponseErr = "Error found while writing response: %s"
const FormFileErr = "Error found while forming file: %s"
const ReadingFileErr = "Error found while reading claim file: %s"
const UnmarshalErr = "Error found while trying to unmarshal claim file: %s"
const MarshalErr = "Error found while marshaling claim file: %s"
const MalformedClaimFileErr = "Malformed claim file: %s"
const ClaimFieldMissingErr = "claim field is missing."
const VersionsFieldMissingErr = "versions field is missing."
const OcpFieldMissingErr = "ocp subfield of versions field is missing."
const TestTestIDMissingErr = "testID subfield of %s test is missing."
const TestStateMissingErr = "state subfield of %s test is missing."
const TestIDSuiteMissingErr = "suite subfield of %s's testID field is missing."
const TestIDIDMissingErr = "id subfield of %s's testID field is missing."
const ResultsFieldMissingErr = "results field is missing."
const ExecutedByMissingErr = "Executed by value is missing."
const MalformedJSONFileErr = "Malformed json file."
const InvalidPasswordErr = "invalid password to given partner's name"
const InvalidUsernameErr = "invalid partner name"
const RollbackErr = "Error found while Rollbacking transaction: %s"
const ExecQueryErr = "Error found while executing a mysql query: %s"
const ScanDBFieldErr = "Error found while scanning db field: %s"
const BeginTxErr = "Error found while beginning transaction: %s"
const CommitTxErr = "Error found while committing transaction: %s"
const AuthError = "Error found while authenticating partner's password: %s"
const EncodingPasswordError = "Failed encoding password." // #nosec
const ServerIsUpMsg = "Server is up."
const ServerReadTimeOutEnvVarErr = "SERVER_READ_TIMEOUT environment variable must be set."
const ServerWriteTimeOutEnvVarErr = "SERVER_WRITE_TIMEOUT environment variable must be set."
const ServerAddrEnvVarErr = "SERVER_ADDR environment variable must be set."
const ServerEnvVarsError = "Error found while extracting environment variables realted to the server: %s"

// parser.go constants
const ClaimTag = "claim"
const VersionsTag = "versions"
const ResultsTag = "results"
const ClaimFileInputName = "claimFile"
const ExecutedByInputName = "executed_by"
const PartnerNameInputName = "partner_name"
const DedcodedPasswordInputName = "decoded_password"

const UseCollectorSQLCmd = `USE cnf; `
const InsertToClaimSQLCmd = `INSERT INTO claim 
								(cnf_version, executed_by, upload_time, partner_name)
								VALUES (?, ?, ?, ?);`
const InsertToClaimResSQLCmd = `INSERT INTO claim_result
							(claim_id, suite_name, test_id, test_status)
							VALUES (?, ?, ?, ?);`
const ExtractLastClaimID = `SELECT id FROM cnf.claim ORDER BY id DESC LIMIT 1;`
const ExtractPartnerAndPasswordCmd = `SELECT encoded_password FROM cnf.authenticator WHERE partner_name = ?`
const InsertPartnerToAuthSQLCmd = `INSERT INTO cnf.authenticator (partner_name, encoded_password) VALUES (?, ?)`
const ParseLowerBound = 10
const ParseUpperBound = 20

const SuccessUploadingFileMSG = "File was uploaded successfully!"

// results.go constants
const SelectAllFromClaimByPartner = "SELECT * FROM cnf.claim WHERE partner_name = ?"
const SelectAllFromClaim = "SELECT * FROM cnf.claim"
const SelectAllClaimIDsByPartner = "SELECT id FROM cnf.claim WHERE partner_name = ?"
const SelectAllFromClaimResultByClaimIDs = "SELECT * FROM cnf.claim_result WHERE claim_id IN (%s)"
const SelectAllFromClaimResult = "SELECT * FROM cnf.claim_result"
const AdminUserName = "admin"

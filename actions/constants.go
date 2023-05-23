package actions

// parser.go constants
const CLAIM_TAG = "claim"
const VERSIONS_TAG = "versions"
const METADATA_TAG = "metadata"
const RESULTS_TAG = "results"
const CLAIM_FILE_INPUT_NAME = "claimFile"
const CREATED_BY_INPUT_NAME = "created_by"
const PARTNER_NAME_INPUT_NAME = "partner_name"

const USE_COLLECTOR_SQL_CMD = `USE cnf; `
const INSERT_TO_CLAIM_SQL_CMD = `INSERT INTO claim 
								(cnf_version, created_by, upload_time, partner_name)
								VALUES (?, ?, ?, ?);`
const INSERT_TO_CLAIM_RES_SQL_CMD = `INSERT INTO claim_result
							(claim_id, suite_name, test_id, test_status)
							VALUES (?, ?, ?, ?);`
const EXTRACT_LAST_CLAIM_ID = `SELECT id FROM cnf.claim ORDER BY id DESC LIMIT 1;`

// results.go constants
const SELECT_ALL_FROM_CLAIM = "SELECT * FROM cnf.claim"
const SELECT_ALL_FROM_CLAIM_RESULT = "SELECT * FROM cnf.claim_result"
const RESULT_JSON_PATH = "results.json"
package api

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/test-network-function/collector/types"
	"github.com/test-network-function/collector/util"
)

func validatePostRequest(w http.ResponseWriter, r *http.Request) ([]types.ClaimResult, []string, error) {
	partnerName := r.FormValue(util.PartnerNameInputName)
	decodedPassword := r.FormValue(util.DedcodedPasswordInputName)

	executedBy := r.FormValue(util.ExecutedByInputName)

	if executedBy == "" {
		return nil, nil, fmt.Errorf(util.ExecutedByMissingErr)
	}

	claimFileMap, err := parseClaimFile(w, r)
	if err != nil {
		// error occurred while uploading\converting claim file.
		return nil, nil, err
	}

	versions, err := validateClaimKeys(claimFileMap)
	if err != nil {
		return nil, nil, err
	}

	// validate results in claim results in JSON
	claimResults, err := verifyClaimResultInJSON(claimFileMap)
	if err != nil {
		return nil, nil, err
	}

	return claimResults, []string{partnerName, decodedPassword, executedBy, versions["ocp"].(string)}, nil
}

func validateGetRequest(r *http.Request, db *sql.DB) (string, error) {
	partnerName := r.FormValue(util.PartnerNameInputName)
	decodedPassword := r.FormValue(util.DedcodedPasswordInputName)

	// If partner name and password are not given return
	if partnerName == "" || decodedPassword == "" {
		return "", fmt.Errorf(util.PartnerOrPasswordArgsMissingErr)
	}

	err := CheckIfValidCredentials(partnerName, decodedPassword, db)
	if err != nil {
		// authentication failed
		return "", err
	}

	return partnerName, nil
}

/*
The validation is done at first, so that we do not keep database transaction open for validation
to reduce network bandwidth to the remote RDS. Also, since the claim file structure can be changed in
the future, validation is kept separate from database insert which gives us less effort to make code
changes.
*/
func verifyClaimResultInJSON(claimFileMap map[string]interface{}) ([]types.ClaimResult, error) {
	results, keyExists := claimFileMap[util.ResultsTag].(map[string]interface{})
	if !keyExists {
		return nil, fmt.Errorf(util.ResultsFieldMissingErr)
	}

	claimResults := []types.ClaimResult{}
	for testName := range results {
		testData, testID, keyErr := validateInnerResultsKeys(results, testName)
		if keyErr != nil {
			return nil, keyErr
		}

		claimResult := types.ClaimResult{
			SuiteName:  testID["suite"].(string),
			TestID:     testID["id"].(string),
			TestStatus: testData["state"].(string),
		}

		claimResults = append(claimResults, claimResult)
	}

	return claimResults, nil
}

func validateInnerResultsKeys(results map[string]interface{}, testName string) (
	testData map[string]interface{}, testID map[string]interface{}, err error) {
	testData, _ = results[testName].(map[string]interface{})

	testID, keyExists := testData["testID"].(map[string]interface{})
	if !keyExists {
		return nil, nil, fmt.Errorf(util.TestTestIDMissingErr, testName)
	}

	_, stateKeyExists := testData["state"]
	if !stateKeyExists {
		return nil, nil, fmt.Errorf(util.TestStateMissingErr, testName)
	}

	_, suiteKeyExists := testID["suite"]
	if !suiteKeyExists {
		return nil, nil, fmt.Errorf(util.TestIDSuiteMissingErr, testName)
	}

	_, idKeyExists := testID["id"]
	if !idKeyExists {
		return nil, nil, fmt.Errorf(util.TestIDIDMissingErr, testName)
	}
	return testData, testID, nil
}

func validateClaimKeys(claimFileMap map[string]interface{}) (map[string]interface{}, error) {
	versions, keyExists := claimFileMap[util.VersionsTag].(map[string]interface{})
	if !keyExists {
		return nil, fmt.Errorf(util.VersionsFieldMissingErr)
	}

	_, keyExists = versions["ocp"]
	if !keyExists {
		return nil, fmt.Errorf(util.OcpFieldMissingErr)
	}

	return versions, nil
}

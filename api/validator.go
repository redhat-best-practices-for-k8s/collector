package api

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/collector/types"
	"github.com/test-network-function/collector/util"
)

func validatePostRequest(w http.ResponseWriter, r *http.Request) ([]types.ClaimResult, []string, bool) {
	partnerName := r.FormValue(util.PartnerNameInputName)
	decodedPassword := r.FormValue(util.DedcodedPasswordInputName)
	if partnerName == "" || decodedPassword == "" {
		return nil, nil, false
	}

	executedBy := r.FormValue(util.ExecutedByInputName)

	if executedBy == "" {
		util.WriteError(w, "%s", util.ExecutedByMissingErr)
		return nil, nil, false
	}

	claimFileMap := parseClaimFile(w, r)
	if claimFileMap == nil {
		// error occurred while uploading\converting claim file.
		return nil, nil, false
	}

	versions := validateClaimKeys(w, claimFileMap)
	if versions == nil {
		return nil, nil, false
	}

	ocpVersion := versions["ocp"].(string)

	// validate results in claim results in JSON
	isValid, claimResults := verifyClaimResultInJson(w, claimFileMap)
	if !isValid {
		return nil, nil, false
	}

	return claimResults, []string{partnerName, decodedPassword, executedBy, ocpVersion}, true
}

func validateGetRequest(w http.ResponseWriter, r *http.Request, db *sql.DB) (string, bool) {
	partnerName := r.FormValue(util.PartnerNameInputName)
	decodedPassword := r.FormValue(util.DedcodedPasswordInputName)

	// If partner name and password are not given return
	if partnerName == "" || decodedPassword == "" {
		return "", false
	}

	err := CheckIfValidCredentials(partnerName, decodedPassword, db)
	if err != nil {
		// authentication failed
		_, err = w.Write([]byte(err.Error() + "\n"))
		if err != nil {
			logrus.Errorf(util.WritingResponseErr, err)
		}
		return "", false
	}

	return partnerName, true
}

/*
The validation is done at first, so that we do not keep database transaction open for validation
to reduce network bandwith to the remote RDS. Also, since the claim file structure can be changed in
the future, validation is kept separate from database insert which gives us less effort to make code
changes.
*/
func verifyClaimResultInJson(w http.ResponseWriter, claimFileMap map[string]interface{}) (bool, []types.ClaimResult) {
	results, keyExists := claimFileMap[util.ResultsTag].(map[string]interface{})
	if !keyExists {
		util.WriteError(w, util.MalformedClaimFileErr, util.ResultsFieldMissingErr)
		return false, nil
	}

	claimResults := []types.ClaimResult{}
	for testName := range results {
		testData, testID, keyErr := validateInnerResultsKeys(results, testName)
		if keyErr != "" {
			util.WriteError(w, util.MalformedClaimFileErr, keyErr)
			return false, nil
		}

		claimResult := types.ClaimResult{
			SuiteName: testID["suite"].(string),
			TestID:    testID["id"].(string),
			TesStatus: testData["state"].(string),
		}

		claimResults = append(claimResults, claimResult)

	}

	return true, claimResults
}

func validateInnerResultsKeys(results map[string]interface{}, testName string) (
	testData map[string]interface{}, testID map[string]interface{}, err string) {
	testData, _ = results[testName].([]interface{})[0].(map[string]interface{})

	testID, keyExists := testData["testID"].(map[string]interface{})
	if !keyExists {
		return nil, nil, fmt.Sprintf(util.TestTestIDMissingErr, testName)
	}

	_, stateKeyExists := testData["state"]
	if !stateKeyExists {
		return nil, nil, fmt.Sprintf(util.TestStateMissingErr, testName)
	}

	_, suiteKeyExists := testID["suite"]
	if !suiteKeyExists {
		return nil, nil, fmt.Sprintf(util.TestIDSuiteMissingErr, testName)
	}

	_, idKeyExists := testID["id"]
	if !idKeyExists {
		return nil, nil, fmt.Sprintf(util.TestIDIDMissingErr, testName)
	}
	return testData, testID, ""
}

func validateClaimKeys(w http.ResponseWriter, claimFileMap map[string]interface{}) map[string]interface{} {
	versions, keyExists := claimFileMap[util.VersionsTag].(map[string]interface{})
	if !keyExists {
		util.WriteError(w, util.MalformedClaimFileErr, util.VersionsFieldMissingErr)
		return nil
	}

	_, keyExists = versions["ocp"]
	if !keyExists {
		util.WriteError(w, util.MalformedClaimFileErr, util.OcpFieldMissingErr)
		return nil
	}

	return versions
}

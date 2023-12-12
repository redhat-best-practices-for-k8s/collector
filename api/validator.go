package api

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/collector/util"
)

func validatePostRequest(w http.ResponseWriter, r *http.Request) (map[string]interface{}, []string, bool) {

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

	claimFileMap := uploadAndConvertClaimFile(w, r)
	if claimFileMap == nil {
		// error occurred while uploading\converting claim file.
		return nil, nil, false
	}

	versions := validateClaimKeys(w, claimFileMap)
	if versions == nil {
		return nil, nil, false
	}

	ocpVersion := versions["ocp"].(string)

	return claimFileMap, []string{partnerName, decodedPassword, executedBy, ocpVersion}, true

}

func validateGetRequest(w http.ResponseWriter, r *http.Request, db *sql.DB) (string, bool) {

	partnerName := r.FormValue(util.PartnerNameInputName)
	decodedPassword := r.FormValue(util.DedcodedPasswordInputName)

	// If partner name and password are not given return
	if partnerName == "" && decodedPassword == "" {
		return "", false
	}

	if partnerName == "" {
		// partner name and password were not given
		return partnerName, false
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

// Done
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

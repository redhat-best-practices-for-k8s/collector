package api

import (
	"encoding/json"
	"net/http"

	"github.com/test-network-function/collector/util"
)

func parseClaimFile(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	claimFileBytes := util.ReadClaimFile(w, r)
	if claimFileBytes == nil {
		// error occurred while reading claim file
		return nil
	}

	var claimFileMap map[string]interface{}
	err := json.Unmarshal(claimFileBytes, &claimFileMap)
	if err != nil {
		util.WriteError(w, util.UnmarshalErr, err.Error())
		return nil
	}

	_, keyExists := claimFileMap[util.ClaimTag]
	if !keyExists {
		util.WriteError(w, util.MalformedClaimFileErr, util.ClaimFieldMissingErr)
		return nil
	}
	return claimFileMap[util.ClaimTag].(map[string]interface{})
}

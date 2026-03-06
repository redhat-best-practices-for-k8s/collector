package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/redhat-best-practices-for-k8s/collector/util"
)

func parseClaimFile(w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	claimFileBytes, err := util.ReadClaimFile(w, r)
	if err != nil {
		// error occurred while reading claim file
		return nil, err
	}

	var claimFileMap map[string]interface{}
	err = json.Unmarshal(claimFileBytes, &claimFileMap)
	if err != nil {
		return nil, err
	}

	claimData, ok := claimFileMap[util.ClaimTag].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf(util.ClaimFieldMissingErr)
	}
	return claimData, nil
}

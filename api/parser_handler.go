package api

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/collector/storage"
	"github.com/test-network-function/collector/util"
)

func ParserHandler(w http.ResponseWriter, r *http.Request, mysqlStorage *storage.MySQLStorage) {
	db := mysqlStorage.MySQL
	defer db.Close()

	// 1. Validate the request (includes validation of the claim file format)
	claimResults, params, isValid := validatePostRequest(w, r)
	if !isValid {
		return
	}

	// Valid parameters for database calls
	partnerName := params[0]
	decodedPassword := params[1]
	ocpVersion := params[2]
	executedBy := params[3]

	// 2. Validate partner's credentials, for non-existent partner create an entry in the database
	// which he has to use each time even when the claim file error happens
	err := VerifyCredentialsAndCreateIfNotExists(partnerName, decodedPassword, db)
	if err != nil {
		util.WriteError(w, util.AuthError, err.Error())
		return
	}

	// 3. Store claim + claim result into the database
	if !util.StoreClaimFileInDatabase(db, claimResults, partnerName, ocpVersion, executedBy) {
		util.WriteError(w, util.ClaimFileError, err.Error())
		return
	}

	// 4. Store file to S3
	claimFile := util.GetClaimFile(w, r)
	if !uploadFileToS3(claimFile, partnerName) {
		return
	}

	// Succfully uploaded file
	_, writeErr := w.Write([]byte(util.SuccessUploadingFileMSG + "\n"))
	if writeErr != nil {
		logrus.Errorf(util.WritingResponseErr, writeErr)
	}
}

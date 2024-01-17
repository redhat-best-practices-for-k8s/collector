package api

import (
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/collector/storage"
	"github.com/test-network-function/collector/util"
)

func ParserHandler(w http.ResponseWriter, r *http.Request, mysqlStorage *storage.MySQLStorage) {
	db := mysqlStorage.MySQL

	// 1. Validate the request (includes validation of the claim file format)
	claimResults, params, isValid := validatePostRequest(w, r)
	if !isValid {
		return
	}

	// Valid parameters for database calls
	partnerName := strings.ToLower(params[0])
	decodedPassword := params[1]
	executedBy := strings.ToLower(params[2])
	ocpVersion := params[3]

	// 2. Validate partner's credentials, for non-existent partner create an entry in the database
	// which he has to use each time even when the claim file error happens
	err := VerifyCredentialsAndCreateIfNotExists(partnerName, decodedPassword, db)
	if err != nil {
		util.WriteError(w, util.AuthError, err.Error())
		return
	}

	// 3. Store file to S3
	claimFile := util.GetClaimFile(w, r)
	s3BucketName, region, accessKey, secretAccessKey := util.GetS3ConnectEnvVars()
	awsS3Client := configS3(region, accessKey, secretAccessKey)
	s3FileKey, success := uploadFileToS3(awsS3Client, claimFile, executedBy, partnerName, s3BucketName)
	if !success {
		return
	}

	// 4. Store claim + claim result into the database
	if !util.StoreClaimFileInDatabase(db, claimResults, partnerName, executedBy, ocpVersion, s3FileKey) {
		deleteFileFromS3(awsS3Client, s3FileKey, s3BucketName)
		util.WriteError(w, util.ClaimFileError, err.Error())
		return
	}

	// Successfully uploaded file
	_, writeErr := w.Write([]byte(util.SuccessUploadingFileMSG + "\n"))
	if writeErr != nil {
		logrus.Errorf(util.WritingResponseErr, writeErr)
	}
}

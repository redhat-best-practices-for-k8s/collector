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
		util.WriteMsg(w, err.Error())
		logrus.Errorf(util.AuthError, err)
		return
	}

	// 3. Store file to S3
	claimFile := util.GetClaimFile(w, r)
	s3BucketName, region, accessKey, secretAccessKey := util.GetS3ConnectEnvVars()
	awsS3Client := configS3(region, accessKey, secretAccessKey)
	s3FileKey, err := uploadFileToS3(awsS3Client, claimFile, executedBy, partnerName, s3BucketName)
	if err != nil {
		util.WriteMsg(w, err.Error())
		logrus.Errorf(util.FailedToUploadFileToS3Err, err)
		return
	}

	// 4. Store claim + claim result into the database
	err = util.StoreClaimFileInDatabase(db, claimResults, partnerName, executedBy, ocpVersion, s3FileKey)
	if err != nil {
		deleteFileFromS3(awsS3Client, s3FileKey, s3BucketName)
		util.WriteMsg(w, err.Error())
		logrus.Errorf(util.ClaimFileError, err)
		return
	}

	// Successfully uploaded file
	util.WriteMsg(w, util.SuccessUploadingFileMSG)
}

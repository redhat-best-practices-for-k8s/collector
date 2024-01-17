package api

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/collector/util"
)

func configS3(region, accessKey, secretAccessKey string) *s3.Client {
	creds := credentials.NewStaticCredentialsProvider(accessKey, secretAccessKey, "")

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithCredentialsProvider(creds), config.WithRegion(region))
	if err != nil {
		logrus.Errorf("error: %v", err)
		return nil
	}

	return s3.NewFromConfig(cfg)
}

func uploadFileToS3(file multipart.File, executedBy, partner string) bool {
	s3bucketName, region, accessKey, secretAccessKey := util.GetS3ConnectEnvVars()
	awsS3Client := configS3(region, accessKey, secretAccessKey)
	uploader := manager.NewUploader(awsS3Client)
	_, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s3bucketName),
		Key:    aws.String(executedBy + "/" + partner + "/claim_" + time.Now().Format("2006-01-02-15:04:05")),
		Body:   file,
	})
	if err != nil {
		logrus.Errorf("error: %v", err)
		return false
	}

	logrus.Infof(util.FileUploadedSuccessfullyToBucket, s3bucketName)
	return true
}

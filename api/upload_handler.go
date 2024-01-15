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

const (
	Region       = "us-east-1" // Region
	S3BucketName = "cnf-suite" // Bucket
)

func configS3() *s3.Client {
	accessKey, secretAccessKey := util.GetS3ConnectEnvVars()
	creds := credentials.NewStaticCredentialsProvider(accessKey, secretAccessKey, "")

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithCredentialsProvider(creds), config.WithRegion(Region))
	if err != nil {
		logrus.Errorf("error: %v", err)
		return nil
	}

	return s3.NewFromConfig(cfg)
}

func uploadFileToS3(file multipart.File, partner string) bool {
	awsS3Client := configS3()
	uploader := manager.NewUploader(awsS3Client)
	_, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(S3BucketName),
		Key:    aws.String(partner + "/claim_" + time.Now().Format("2006-01-02-15:04:05")),
		Body:   file,
	})
	if err != nil {
		logrus.Errorf("error: %v", err)
		return false
	}

	logrus.Infof(util.FileUploadedSuccessfullyToBucket, S3BucketName)
	return true
}

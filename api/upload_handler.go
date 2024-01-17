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

func deleteFileFromS3(awsS3Client *s3.Client, s3FileKey string) {
	deleteFileInput := s3.DeleteObjectInput{
		Bucket: aws.String(S3BucketName),
		Key:    aws.String(s3FileKey),
	}

	_, err := awsS3Client.DeleteObject(context.TODO(), &deleteFileInput)
	if err != nil {
		logrus.Errorf(util.FailedToDeleteFileFromS3Err, err)
	}

	logrus.Infof(util.FileHasBeenDeletedFromBucket, S3BucketName)
}

func uploadFileToS3(awsS3Client *s3.Client, file multipart.File, executedBy, partner string) (string, bool) {
	s3BucketName, region, accessKey, secretAccessKey := util.GetS3ConnectEnvVars()
	awsS3Client := configS3(region, accessKey, secretAccessKey)
	uploader := manager.NewUploader(awsS3Client)
	s3FileKey := executedBy + "/" + partner + "/claim_" + time.Now().Format("2006-01-02-15:04:05")
	_, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(S3BucketName),
		Key:    aws.String(s3FileKey),
		Body:   file,
	})
	if err != nil {
		logrus.Errorf("error: %v", err)
		return "", false
	}

	logrus.Infof(util.FileUploadedSuccessfullyToBucket, S3BucketName)
	return s3FileKey, true
}

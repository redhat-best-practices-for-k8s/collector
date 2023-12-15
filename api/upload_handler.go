package api

import (
	"context"
	"log"
	"mime/multipart"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/test-network-function/collector/util"
)

func HandleUpload() {

}

const (
	Region       = "us-east-1"     // Region
	S3BucketName = "cnf-collector" // Bucket
)

func configS3() *s3.Client {
	accessKey, secretAccessKey := util.GetS3ConnectEnvVars()
	creds := credentials.NewStaticCredentialsProvider(accessKey, secretAccessKey, "")

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithCredentialsProvider(creds), config.WithRegion(S3BucketName))
	if err != nil {
		log.Printf("error: %v", err)
		return nil
	}

	return s3.NewFromConfig(cfg)
}

func uploadFileToS3(file multipart.File, partner string) {
	awsS3Client := configS3()
	uploader := manager.NewUploader(awsS3Client)
	_, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(S3BucketName),
		Key:    aws.String("claim_" + partner + "_" + time.Now().Format("YYYY-MM-DD")),
		Body:   file,
	})
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

	log.Printf("Sucecssfully uploaded to %q\n", S3BucketName)
}

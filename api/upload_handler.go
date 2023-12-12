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
)

func HandleUpload() {

}

const (
	AWS_S3_REGION  = "us-east-1"     // Region
	AWS_S3_BUCKET  = "cnf-collector" // Bucket
	AWS_ACCESS_KEY = ""
	AWS_SECRET_KEY = ""
)

// We will be using this client everywhere in our code
var awsS3Client *s3.Client

func configS3() {
	creds := credentials.NewStaticCredentialsProvider(AWS_ACCESS_KEY, AWS_SECRET_KEY, "")
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithCredentialsProvider(creds), config.WithRegion(AWS_S3_BUCKET))
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

	awsS3Client = s3.NewFromConfig(cfg)
}

func uploadFileToS3(file multipart.File, partner string) {
	uploader := manager.NewUploader(awsS3Client)
	_, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(AWS_S3_BUCKET),
		Key:    aws.String("claim_" + partner + "_" + time.Now().Format("YYYY-MM-DD")),
		Body:   file,
	})
	if err != nil {
		// Do your error handling here
		log.Printf("error: %v", err)
		return
	}

	log.Printf("Sucecssfully uploaded to %q\n", AWS_S3_BUCKET)
}

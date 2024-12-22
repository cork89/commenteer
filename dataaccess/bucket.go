package dataaccess

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
)

var accountId string
var accessKeyId string
var accessKeySecret string
var bucketName string
var client *s3.Client

type RealBucketUploader struct{}

func (r RealBucketUploader) UploadImage(img io.Reader, fileName string) (*string, error) {
	uploader := manager.NewUploader(client)
	output, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket:          aws.String(bucketName),
		Key:             aws.String(fileName),
		Body:            img,
		ContentType:     aws.String("image/webp"),
		ContentEncoding: aws.String("base64"),
	})

	if err != nil {
		log.Printf("failed to save image to bucket: %v\n", err)
		return nil, err
	}

	return &output.Location, nil
}

func (r RealBucketUploader) InitializeBucket() {
	err := godotenv.Load("/run/secrets/.env.local")
	if err != nil {
		log.Println(err)
	}

	accountId = os.Getenv("R2_ACCOUNT_ID")
	accessKeyId = os.Getenv("R2_ACCESS_KEY_ID")
	accessKeySecret = os.Getenv("R2_ACCESS_KEY_SECRET")
	bucketName = os.Getenv("R2_BUCKET_NAME")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, "")),
		config.WithRegion("auto"))
	if err != nil {
		slog.Error("failed to create config", "error", err.Error())
	} else {
		client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId))
		})
	}
}

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
var client *s3.Client

func UploadImage(img io.Reader, fileName string) (*string, error) {
	// n, err := io.Copy(io.Discard, img)
	// if err != nil {
	// 	log.Printf("failed to get image size, %v\n", err)
	// 	return err
	// }

	// noRetry := func(options *s3.Options) {
	// 	options.RetryMaxAttempts = 1
	// }

	uploader := manager.NewUploader(client)
	output, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket:          aws.String("commenteer-beta"),
		Key:             aws.String(fileName),
		Body:            img,
		ContentType:     aws.String("image/webp"),
		ContentEncoding: aws.String("base64"),
	})

	// Get the first page of results for ListObjectsV2 for a bucket
	// _, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
	// 	Bucket:        aws.String("commenteer-beta"),
	// 	Key:           aws.String(fileName),
	// 	Body:          img,
	// 	ContentLength: aws.Int64(n),
	// }, noRetry)

	if err != nil {
		log.Printf("failed to save image to bucket: %v\n", err)
		return nil, err
	}

	// log.Println("first page results:")
	// for _, object := range output.Contents {
	// 	log.Printf("key=%s size=%d", aws.ToString(object.Key), object.Size)
	// }
	// output.
	return &output.Location, nil
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	accountId = os.Getenv("R2_ACCOUNT_ID")
	accessKeyId = os.Getenv("R2_ACCESS_KEY_ID")
	accessKeySecret = os.Getenv("R2_ACCESS_KEY_SECRET")

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

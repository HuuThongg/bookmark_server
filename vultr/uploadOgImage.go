package vultr

import (
	"bookmark/util"
	"bytes"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

func UploadOgImage(config util.Config, imgBuffer []byte, imgFormat string) (string, error) {

	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(config.CLOUDFLARE_ACCESS_KEY_ID, config.CLOUDFLARE_SECRET_ACCESS_KEY, ""),
		Endpoint:         aws.String(config.CLOUDFLARE_ENDPOINT),
		S3ForcePathStyle: aws.Bool(true),
		Region:           aws.String(config.CLOUDFLARE_REGION),
	}
	sess, err := session.NewSession(s3Config)
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	// Create an S3 service client
	svc := s3.New(sess)
	key := fmt.Sprintf("%s/%s", config.CLOUDFLARE_OG_BUCKET, uuid.NewString())
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:       aws.String(config.CLOUDFLARE_OG_BUCKET),
		Key:          aws.String(key),
		Body:         bytes.NewReader(imgBuffer),
		ACL:          aws.String("public-read"),
		CacheControl: aws.String("public, max-age=31536000"),

		ContentLength: aws.Int64(int64(len(imgBuffer))),
		ContentType:   aws.String(imgFormat),
		// Expires:      aws.Time(time.Now().AddDate(1, 0, 0)),
	})
	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}

	fmt.Println("File uploaded successfully!")
	return fmt.Sprintf("%s/%s", "https://bookmarking.app", key), nil
}

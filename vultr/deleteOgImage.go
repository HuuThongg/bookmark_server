package vultr

import (
	"bookmark/util"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// DeleteOgImage deletes an image from Cloudflare R2
func DeleteOgImage(config util.Config, key string) error {
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

	// Delete the object from Cloudflare R2
	_, err = svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(config.CLOUDFLARE_OG_BUCKET),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	// Optionally, you can call WaitUntilObjectNotExists to wait for the deletion to complete
	// err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
	// 	Bucket: aws.String(config.CLOUDFLARE_OG_BUCKET),
	// 	Key:    aws.String(key),
	// })
	// if err != nil {
	// 	return fmt.Errorf("failed to wait for deletion: %v", err)
	// }

	fmt.Println("File deleted successfully!")
	return nil
}

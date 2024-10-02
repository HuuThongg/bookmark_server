package vultr

import (
	"fmt"
	"log"
	"os"

	"bookmark/util"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

func UploadHeroImage(heroImage *os.File) string {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Panicf("could not load conig file: %v", err)
	}
	fmt.Println("Endpoint", config.BlackBlazeHostName)
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(config.BlackBlazeKeyId, config.BlackBlazeSecretKey, ""),
		Endpoint:         aws.String(config.BlackBlazeHostName),
		S3ForcePathStyle: aws.Bool(false),
		Region:           aws.String("us-west-002"),
	}

	newSession, err := session.NewSession(s3Config)
	if err != nil {
		log.Panicf("could not create new vultr s3 session: %v", err)
	}

	s3Client := s3.New(newSession)

	object := s3.PutObjectInput{
		Bucket: aws.String("/app-assets"),
		Key:    aws.String(uuid.NewString()),
		Body:   heroImage,
		// ACL:    aws.String("public-read"),
	}

	_, err = s3Client.PutObject(&object)
	if err != nil {
		log.Panicf("could not upload hero image to app-assets bucket: %v", err)
	}

	return fmt.Sprintf("%s/app-assets/%s", config.BlackBlazeHostName, *object.Key)
}

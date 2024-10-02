package vultr

import (
	"bookmark/util"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func DeleteObjectFromBucket(bucket, key string) {

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Panicf("could not load config file: %v", err)
	}

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
	object := &s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}

	_, err = s3Client.DeleteObject(object)
	if err != nil {
		log.Panicf("could not delete object from vultr: %v", err)
	}
}

package vultr

import (
	"bookmark/util"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

func UploadLinkFavicon(linkFaviconChannel chan string) {

	log.Println("uploading link favicon...")
	imgFileChan := make(chan *os.File, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := util.LoadImage(imgFileChan, "favicon.icon"); err != nil {
			log.Panicf("could not load link favicon :%v", err)
		}
	}()

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Panicf("could not load config file: %v", err)
	}
	fmt.Println("UploadLinkFavicon: Endpoint: ", config.BlackBlazeHostName)
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
		Bucket:       aws.String("/link-favicons"),
		Key:          aws.String(uuid.NewString()),
		Body:         <-imgFileChan,
		ACL:          aws.String("public-read"),
		CacheControl: aws.String("public, max-age=31536000"),
		Expires:      aws.Time(time.Now().Add(365 * 24 * time.Hour)),
	}

	_, err = s3Client.PutObject(&object)
	if err != nil {
		log.Panicf("could not upload link favicon to vultr: %v", err)
	}
	log.Printf("link favicon url: %s", fmt.Sprintf("%s/link-favicons/%s", config.BlackBlazeHostName, *object.Key))
	linkFaviconChannel <- fmt.Sprintf("%s/link-favicons/%s", config.BlackBlazeHostName, *object.Key)

	wg.Wait()
	log.Println("successfully uploaded link favicon")
}

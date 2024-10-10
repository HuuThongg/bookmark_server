package vultr

import (
	"bookmark/util"
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
)

func UploadLinkThumbnail(linkThumbnailChannel chan string) {
	imgFileChan := make(chan *os.File, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := util.LoadImage(imgFileChan, "a.jpeg"); err != nil {
			log.Panicf("could not load link thumbnail: %v", err)
		}
	}()

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
		log.Panicf("could not create a new vultr s3 session: %v", err)
	}
	s3Client := s3.New(newSession)
	object := s3.PutObjectInput{
		Bucket: aws.String("/link-thumbnails"),
		Key:    aws.String(uuid.NewString()),
		Body:   <-imgFileChan,
		// ACL:    aws.String("public-reaç"),
	}
	_, err = s3Client.PutObject(&object)
	if err != nil {
		log.Panicf("could not upload link thumbnail to vultr: %v", err)
	}

	log.Printf("link thumbnail url: %s", fmt.Sprintf("%s/%s", config.BlackBlazeHostName, *object.Key))
	linkThumbnailChannel <- fmt.Sprintf("%s/link-thumbnails/%s", config.BlackBlazeHostName, *object.Key)
	wg.Wait()
}

func UploadLinkThumbnail1(linkThumbnailChannel chan string, screenShotBytes []byte) {
	resizedImgBytes, imageErr := util.ResizeImage(screenShotBytes, 350, 0)
	if imageErr != nil {
		log.Println("image err", imageErr)
	}

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
		log.Panicf("could not create a new vultr s3 session: %v", err)
	}
	s3Client := s3.New(newSession)
	object := s3.PutObjectInput{
		Bucket:        aws.String("/link-thumbnails"),
		Key:           aws.String(uuid.NewString()),
		Body:          bytes.NewReader(resizedImgBytes),
		ContentLength: aws.Int64(int64(len(resizedImgBytes))),
		// ContentType:   aws.String("image/jpeg"),
	}
	_, err = s3Client.PutObject(&object)
	if err != nil {
		log.Panicf("could not upload link thumbnail to vultr: %v", err)
	}

	log.Printf("link thumbnail url: %s", fmt.Sprintf("%s/%s", config.BlackBlazeHostName, *object.Key))
	linkThumbnailChannel <- fmt.Sprintf("%s/link-thumbnails/%s", config.BlackBlazeHostName, *object.Key)
}

func UploadLinkThumbnail2(linkThumbnailChannel chan string, ogImage, host string) {

	if ogImage == "" {
		fmt.Println("ogImage is empty")
		return
	}
	// Call the function to download the image
	imageBuffer, imgType, err1 := DownloadImage(ogImage, host)
	if err1 != nil {
		fmt.Println("Error downloading the image:", err1)
		return
	}
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
		log.Panicf("could not create a new vultr s3 session: %v", err)
	}
	imgFormat := fmt.Sprintf("image/%v", imgType)

	s3Client := s3.New(newSession)
	object := s3.PutObjectInput{
		Bucket:        aws.String("/link-thumbnails"),
		Key:           aws.String(uuid.NewString()),
		Body:          bytes.NewReader(imageBuffer),
		ContentLength: aws.Int64(int64(len(imageBuffer))),
		ContentType:   aws.String(imgFormat),
	}
	_, err = s3Client.PutObject(&object)
	if err != nil {
		log.Panicf("could not upload link thumbnail to vultr: %v", err)
	}

	log.Printf("link thumbnail url: %s", fmt.Sprintf("%s/%s", config.BlackBlazeHostName, *object.Key))
	linkThumbnailChannel <- fmt.Sprintf("%s/link-thumbnails/%s", config.BlackBlazeHostName, *object.Key)
}

func DownloadImage(url, host string) ([]byte, string, error) {
	// Create the HTTP client with a timeout
	client := &http.Client{}

	// Create an HTTP GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}

	// Optional: Add headers if necessary (e.g., User-Agent)
	req.Header.Set("User-Agent", "Go Image Downloader")

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	contentType := resp.Header.Get("Content-Type")
	fmt.Printf("Content-Type of the image: %s\n", contentType)

	// Check if the HTTP status is OK
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed to download image: status code %d", resp.StatusCode)
	}

	// // Create the file where the image will be saved

	parts := strings.Split(contentType, "/")
	var imageFormat string
	if len(parts) <= 1 {

		fmt.Println("Unable to determine the image format")
	}
	imageFormat = parts[1]
	var img image.Image
	var err1 error
	fmt.Println("imageFormat", imageFormat)
	filePath := fmt.Sprintf("%s.og.%s", host, imageFormat)
	outFile, err := os.Create(filePath)
	if err != nil {
		return nil, "", err
	}
	defer outFile.Close()
	if imageFormat == "jpeg" || imageFormat == "jpg" {
		img, err1 = jpeg.Decode(resp.Body)
		if err1 != nil {
			log.Fatalf("Failed to decode JPEG image: %v", err1)
		}
		fmt.Println("JPEG image decoded successfully:", img.Bounds())
	} else if imageFormat == "png" {
		img, err1 = png.Decode(resp.Body)
		if err1 != nil {
			log.Fatalf("Failed to decode JPEG image: %v", err)
		}
		fmt.Println("Image format is not JPEG, skipping decode.")
	}

	m := resize.Resize(300, 0, img, resize.Lanczos2)

	if imageFormat == "jpeg" || imageFormat == "jpg" {

		jpeg.Encode(outFile, m, nil)
	}
	if imageFormat == "png" {
		png.Encode(outFile, m)
	}
	// return nil
	var buf bytes.Buffer
	if imageFormat == "jpeg" || imageFormat == "jpg" {
		err = jpeg.Encode(&buf, m, nil)
		if err != nil {
			return nil, "", err
		}
	} else if imageFormat == "png" {
		err = png.Encode(&buf, m)
		if err != nil {
			return nil, "", err
		}
	}

	// Return the buffer as a byte slice and the format
	return buf.Bytes(), imageFormat, nil
}

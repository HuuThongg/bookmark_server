package util

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
)

func DownloadFavicon(URL, fileName string) error {
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		log.Printf("status code: %v", response.StatusCode)
		return errors.New("received non 200 response code")
	}

	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}
	return nil
}

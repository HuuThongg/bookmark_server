package util

import (
	"log"
	"os"
)

func LoadImage(imgFileChan chan *os.File, f string) error {
	imgFile, err := os.Open(f)
	if err != nil {
		log.Printf("LoadImage: failed to open file: %v", err)
		return err
	}
	imgFileChan <- imgFile
	return nil
}

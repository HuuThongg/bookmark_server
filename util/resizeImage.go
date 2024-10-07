package util

import (
	"bytes"
	"image"
	"image/jpeg"
	_ "image/jpeg" // Import to support JPEG decoding
	_ "image/png"  // Import to support PNG decoding

	"github.com/nfnt/resize"
)

func ResizeImage(imgBytes []byte, width, height uint) ([]byte, error) {
	// Decode image from bytes (automatically detects format: JPEG, PNG, etc.)
	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return nil, err
	}

	// Resize the image
	resizedImg := resize.Resize(width, height, img, resize.Lanczos3)

	// Encode resized image to a new byte buffer (saving as JPEG)
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, resizedImg, nil)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

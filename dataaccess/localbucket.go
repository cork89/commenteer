package dataaccess

import (
	"errors"
	"io"
)

type LocalBucketUploader struct{}

var bucket map[string]string

func (r LocalBucketUploader) UploadImage(img io.Reader, fileName string) (*string, error) {
	if fileName == "invalid-invalid-invalid.webp" {
		return nil, errors.New("failed to upload")
	}
	imgBytes, _ := io.ReadAll(img)
	bucket[fileName] = string(imgBytes)
	return &fileName, nil
}

func (r LocalBucketUploader) InitializeBucket() {
	bucket = make(map[string]string, 0)
}

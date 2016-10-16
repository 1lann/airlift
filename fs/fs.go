package fs

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/minio/minio-go"
)

var client *minio.Client

// Errors that can occur during uploading
var (
	ErrInvalidType = errors.New("fs: invalid detected file type")
	ErrTooBig      = errors.New("fs: file too big")
	ErrNotFound    = errors.New("fs: file not found")
)

// MaxUpload represents the maximum upload in bytes for an individual file
const MaxUpload = 210000000

// const MaxUpload = 10

// File represents a file
type File struct {
	*minio.Object
	minio.ObjectInfo
}

// Connect connects to a Minio instance
func Connect(address, accessKey, secretKey string) error {
	var err error
	// TODO: Add SSL
	client, err = minio.New(address, accessKey, secretKey, false)
	return err
}

// RetrieveFile retrieves a file in a bucket. It is recommended to use a
// presigned file instead.
func RetrieveFile(bucket, name string) (File, error) {
	obj, err := client.GetObject(bucket, name)
	if err != nil {
		errResp, ok := err.(minio.ErrorResponse)
		if ok {
			if errResp.Code == "NoSuchKey" {
				return File{}, ErrNotFound
			}
		}

		return File{}, err
	}

	stat, err := obj.Stat()
	if err != nil {
		return File{}, err
	}

	return File{
		Object:     obj,
		ObjectInfo: stat,
	}, nil
}

// StatFile retrieves only the ObjectInfo component of a file in a bucket.
func StatFile(bucket, name string) (File, error) {
	stat, err := client.StatObject(bucket, name)
	if err != nil {
		errResp, ok := err.(minio.ErrorResponse)
		if ok {
			if errResp.Code == "NoSuchKey" {
				return File{}, ErrNotFound
			}
		}

		return File{}, err
	}

	return File{
		ObjectInfo: stat,
	}, nil
}

// UploadFile uploads a file to a bucket
func UploadFile(bucket, name, requiredType string, rd io.Reader) error {
	buf := make([]byte, 1024)
	_, err := io.ReadFull(rd, buf)
	if err != nil {
		return err
	}

	contentType := http.DetectContentType(buf)
	bufRd := bytes.NewReader(buf)

	if contentType != requiredType {
		return ErrInvalidType
	}

	limitedRd := io.LimitReader(io.MultiReader(bufRd, rd), MaxUpload)

	n, err := client.PutObject(bucket, name, limitedRd, contentType)
	if err != nil {
		client.RemoveObject(bucket, name)
		client.RemoveIncompleteUpload(bucket, name)
		return err
	}

	if n >= MaxUpload {
		client.RemoveObject(bucket, name)
		client.RemoveIncompleteUpload(bucket, name)
		return ErrTooBig
	}

	return nil
}

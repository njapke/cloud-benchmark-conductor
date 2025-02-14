package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"cloud.google.com/go/storage"
)

func NewObjectWriter(ctx context.Context, bucketName, objectName string) (*storage.Writer, *storage.Client, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	// check if bucket exists
	bucket := client.Bucket(bucketName)
	_, err = bucket.Attrs(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrBucketNotExist) {
			return nil, nil, fmt.Errorf("bucket %s does not exist", bucketName)
		}
		return nil, nil, err
	}

	objectName = strings.TrimPrefix(objectName, "/")
	objectWriter := bucket.Object(objectName).NewWriter(ctx)
	return objectWriter, client, nil
}

func UploadToBucket(ctx context.Context, bucketName, objectName string, r io.Reader) error {
	objectWriter, client, err := NewObjectWriter(ctx, bucketName, objectName)
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = io.Copy(objectWriter, r)
	if err != nil {
		return err
	}
	return objectWriter.Close()
}

func UploadFileToBucket(ctx context.Context, bucketName, objectName, inputFile string) error {
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", inputFile, err)
	}
	defer file.Close()
	return UploadToBucket(ctx, bucketName, objectName, file)
}

func ParseURL(inputURL string) (string, string, error) {
	u, err := url.Parse(inputURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme != "gs" && u.Scheme != "gcs" {
		return "", "", fmt.Errorf("invalid scheme: %s", u.Scheme)
	}
	return u.Host, u.Path, nil
}

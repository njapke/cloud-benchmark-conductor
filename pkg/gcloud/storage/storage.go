package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
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

package output

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/hashicorp/go-multierror"
)

type gcsWriter struct {
	client *storage.Client
	writer *storage.Writer
}

func stripSlashPrefix(s string) string {
	return strings.TrimPrefix(s, "/")
}

func newGCSWriter(config *Output) (io.WriteCloser, error) {
	client, err := storage.NewClient(config.Context)
	if err != nil {
		return nil, err
	}
	bucket := client.Bucket(config.Host)

	// check if bucket exists
	_, err = bucket.Attrs(config.Context)
	if err != nil {
		_ = client.Close()
		if errors.Is(err, storage.ErrBucketNotExist) {
			return nil, fmt.Errorf("bucket %s does not exist", config.Host)
		}
		return nil, err
	}
	object := bucket.Object(stripSlashPrefix(config.GetPath()))
	return &gcsWriter{
		client: client,
		writer: object.NewWriter(config.Context),
	}, nil
}

func (g *gcsWriter) Write(p []byte) (n int, err error) {
	return g.writer.Write(p)
}

func (g *gcsWriter) Close() error {
	var mErr error
	if err := g.writer.Close(); err != nil {
		mErr = multierror.Append(mErr, err)
	}
	if err := g.client.Close(); err != nil {
		mErr = multierror.Append(mErr, err)
	}
	return mErr
}

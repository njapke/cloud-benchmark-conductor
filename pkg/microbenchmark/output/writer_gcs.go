package output

import (
	"io"

	"github.com/christophwitzko/master-thesis/pkg/gcloud/storage"
	"github.com/hashicorp/go-multierror"
)

type gcsWriter struct {
	client io.Closer
	writer io.WriteCloser
}

func newGCSWriter(config *Output) (io.WriteCloser, error) {
	objectWriter, client, err := storage.NewObjectWriter(config.Context, config.Host, config.GetPath())
	if err != nil {
		return nil, err
	}

	return &gcsWriter{
		client: client,
		writer: objectWriter,
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

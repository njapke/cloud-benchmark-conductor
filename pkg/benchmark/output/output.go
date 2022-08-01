package output

import (
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/hashicorp/go-multierror"
)

// adapted MultiWriter from https://github.com/golang/go/blob/master/src/io/multi.go
type multiWriteCloser struct {
	writeClosers []io.WriteCloser
}

func (m *multiWriteCloser) Write(p []byte) (int, error) {
	for _, w := range m.writeClosers {
		n, err := w.Write(p)
		if err != nil {
			return n, err
		}
		if len(p) != n {
			return n, io.ErrShortWrite
		}
	}
	return len(p), nil
}

func (m *multiWriteCloser) Close() error {
	var mErr error
	for _, w := range m.writeClosers {
		if err := w.Close(); err != nil {
			mErr = multierror.Append(mErr, err)
		}
	}
	return mErr
}

func New(outputPath string) (io.WriteCloser, error) {
	if outputPath == "-" {
		return newFileFromOSFile(os.Stdout)
	}
	parsedPath, err := url.Parse(outputPath)
	if err != nil {
		return nil, err
	}
	if parsedPath.Scheme == "file" || parsedPath.Scheme == "" {
		return newFileFromPath(parsedPath.Path)
	}
	return nil, fmt.Errorf("%s not implemented", parsedPath.Scheme)
}

func NewMultiOutput(outputPaths []string) (io.WriteCloser, error) {
	writeClosers := make([]io.WriteCloser, 0, len(outputPaths))
	for _, outputPath := range outputPaths {
		wc, err := New(outputPath)
		if err != nil {
			// close already opened writers
			for _, wc := range writeClosers {
				if cErr := wc.Close(); cErr != nil {
					err = multierror.Append(err, cErr)
				}
			}
			return nil, err
		}
		writeClosers = append(writeClosers, wc)
	}
	return &multiWriteCloser{writeClosers: writeClosers}, nil
}

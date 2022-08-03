package output

import (
	"io"
	"os"
)

type fileWriter struct {
	osFile *os.File
}

func newFileWriter(config *Output) (io.WriteCloser, error) {
	path := config.GetPath()
	if path == "-" {
		return &fileWriter{osFile: os.Stdout}, nil
	}
	outFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &fileWriter{osFile: outFile}, nil
}

func (f *fileWriter) Write(p []byte) (n int, err error) {
	return f.osFile.Write(p)
}

func (f *fileWriter) Close() error {
	_ = f.osFile.Sync()
	return f.osFile.Close()
}

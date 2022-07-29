package output

import (
	"io"
	"os"
)

type File struct {
	osFile *os.File
}

func newFileFromPath(outputFile string) (io.WriteCloser, error) {
	outFile, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return newFileFromOSFile(outFile)
}

func newFileFromOSFile(osFile *os.File) (io.WriteCloser, error) {
	return &File{osFile: osFile}, nil
}

func (f *File) Write(p []byte) (n int, err error) {
	return f.osFile.Write(p)
}

func (f *File) Close() error {
	_ = f.osFile.Sync()
	return f.osFile.Close()
}

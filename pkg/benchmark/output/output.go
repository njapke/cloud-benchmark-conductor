package output

import (
	"fmt"
	"io"
	"net/url"
	"os"
)

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

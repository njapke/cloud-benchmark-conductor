package dir

import (
	"fmt"
	"os"
	"path/filepath"
)

func FromPath(path, pathType string) *os.File {
	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get absolute %s path: %v\n", pathType, err)
		os.Exit(3)
	}

	f, err := os.Open(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s file error: %v\n", pathType, err)
		os.Exit(3)
	}

	stat, err := f.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s file error: %v\n", pathType, err)
		os.Exit(3)
	}

	if !stat.IsDir() {
		fmt.Fprintf(os.Stderr, "%s file is not a directory\n", pathType)
		os.Exit(4)
	}

	return f
}

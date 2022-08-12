package output

import (
	"fmt"
	"io"
)

type WriterFactory func(config *Output) (io.WriteCloser, error)

var writers = map[string]WriterFactory{
	"file": newFileWriter,
	"gcs":  newGCSWriter,
	"gs":   newGCSWriter,
}

func IsValidSchema(schema string) bool {
	_, ok := writers[schema]
	return ok
}

func NewWriter(config *Output) (io.WriteCloser, error) {
	wFactory, ok := writers[config.Schema]
	if !ok {
		return nil, fmt.Errorf("unsupported output schema: %s", config.Type)
	}
	return wFactory(config)
}

package output

import (
	"fmt"
	"io"
	"net/url"
	"path/filepath"

	"github.com/christophwitzko/master-thesis/pkg/benchmark"
)

type Output struct {
	Schema     string
	Path       string
	Type       string
	Parameters url.Values
	writer     io.WriteCloser
	encoder    ResultEncoder
}

func newOutput(outputPath string, defaultType string) (*Output, error) {
	outputType := defaultType
	parsedPath, err := url.Parse(outputPath)
	if err != nil {
		return nil, err
	}
	if ext := filepath.Ext(parsedPath.Path); ext != "" {
		outputType = ext[1:] // strip "."
	}
	schema := parsedPath.Scheme
	if schema == "" {
		schema = "file"
	}
	o := &Output{
		Schema:     schema,
		Path:       parsedPath.Path,
		Type:       outputType,
		Parameters: parsedPath.Query(),
	}

	// setup encoder
	switch o.Type {
	case "json":
		o.encoder, err = NewJSONResultEncoder(o)
	case "csv":
		o.encoder, err = NewCSVResultEncoder(o)
	default:
		err = fmt.Errorf("unsupported output type: %s", o.Type)
	}
	if err != nil {
		return nil, err
	}

	return o, nil
}

func (o *Output) open() error {
	if o.writer != nil {
		return nil
	}
	var err error
	switch o.Schema {
	case "file":
		o.writer, err = newFileWriterFromPath(o.Path)
	default:
		err = fmt.Errorf("unsupported output schema: %s", o.Schema)
	}
	return err
}

func (o *Output) Write(result benchmark.Result) error {
	if err := o.open(); err != nil {
		return err
	}
	env, err := o.encoder.Encode(result)
	if err != nil {
		return err
	}
	_, err = o.writer.Write(env)
	return err
}

func (o *Output) Close() error {
	err := o.writer.Close()
	o.writer = nil
	return err
}

func New(outputPaths []string, defaultType string) (benchmark.ResultWriter, error) {
	resultWriters := make([]benchmark.ResultWriter, 0, len(outputPaths))
	for _, outputPath := range outputPaths {
		out, err := newOutput(outputPath, defaultType)
		if err != nil {
			return nil, err
		}
		resultWriters = append(resultWriters, out)
	}
	return benchmark.NewMultiResultWriter(resultWriters), nil
}

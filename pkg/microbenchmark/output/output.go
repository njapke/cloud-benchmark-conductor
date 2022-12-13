package output

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"sync"

	"github.com/christophwitzko/masters-thesis/pkg/microbenchmark"
)

type NewChunkFunc func(lastResult, newResult *microbenchmark.Result) bool

func NoChunkFn(lastResult, newResult *microbenchmark.Result) bool {
	return false
}

func SuiteChunkFn(lastResult, newResult *microbenchmark.Result) bool {
	if lastResult == nil {
		return true
	}
	return lastResult.S != newResult.S
}

func BenchFnChunkFn(lastResult, newResult *microbenchmark.Result) bool {
	// always chunk a new suite run
	if SuiteChunkFn(lastResult, newResult) {
		return true
	}
	return lastResult.Function.String() != newResult.Function.String()
}

type Output struct {
	Context context.Context

	Schema     string
	Type       string
	Host       string
	path       string
	Parameters url.Values

	writer     io.WriteCloser
	encoder    ResultEncoder
	writeMutex sync.Mutex

	chunked    bool
	newChunkFn func(previous, current *microbenchmark.Result) bool
	chunkIndex uint64
	lastResult *microbenchmark.Result
}

func newOutput(ctx context.Context, outputPath, defaultType string) (*Output, error) {
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
	if !IsValidSchema(schema) {
		return nil, fmt.Errorf("unsupported output schema: %s", schema)
	}
	params := parsedPath.Query()
	o := &Output{
		Context:    ctx,
		Schema:     schema,
		Type:       outputType,
		Host:       parsedPath.Host,
		path:       parsedPath.Path,
		Parameters: params,
		chunked:    params.Get("chunked") == "true",
	}

	if o.chunked && o.path == "-" {
		return nil, fmt.Errorf("cannot chunk to stdout")
	}

	if o.chunked {
		chunkFn := params.Get("new-chunk-fn")
		switch chunkFn {
		case "no-chunk":
			o.newChunkFn = NoChunkFn
		case "suite":
			o.newChunkFn = SuiteChunkFn
		case "benchFn":
			o.newChunkFn = BenchFnChunkFn
		default:
			o.newChunkFn = BenchFnChunkFn
		}
	}

	o.encoder, err = NewEncoder(o)
	if err != nil {
		return nil, err
	}

	return o, nil
}

// opens the output file if it is not already open
// if already open, closes the file and opens it again
func (o *Output) open() error {
	if o.writer != nil {
		if err := o.writer.Close(); err != nil {
			return err
		}
	}
	var err error
	o.writer, err = NewWriter(o)
	return err
}

func (o *Output) GetPath() string {
	if !o.chunked {
		return o.path
	}
	return fmt.Sprintf("%s.%04d", o.path, o.chunkIndex)
}

func (o *Output) Write(result microbenchmark.Result) error {
	o.writeMutex.Lock()
	defer o.writeMutex.Unlock()

	// open new writer if it is not already open or a new chunk is needed
	isNewChunk := o.chunked && o.newChunkFn(o.lastResult, &result)
	if isNewChunk || o.writer == nil {
		if err := o.open(); err != nil {
			return err
		}
		if isNewChunk {
			o.chunkIndex++
		}
	}

	env, err := o.encoder.Encode(result)
	if err != nil {
		return err
	}
	n, err := o.writer.Write(env)
	if err != nil {
		return err
	}
	if n != len(env) {
		return io.ErrShortWrite
	}
	o.lastResult = &result
	return nil
}

func (o *Output) Close() error {
	o.writeMutex.Lock()
	defer o.writeMutex.Unlock()
	if o.writer != nil {
		err := o.writer.Close()
		o.writer = nil
		return err
	}
	return nil
}

func New(ctx context.Context, outputPaths []string, defaultType string) (microbenchmark.ResultWriter, error) {
	resultWriters := make([]microbenchmark.ResultWriter, 0, len(outputPaths))
	for _, outputPath := range outputPaths {
		out, err := newOutput(ctx, outputPath, defaultType)
		if err != nil {
			return nil, err
		}
		resultWriters = append(resultWriters, out)
	}
	return microbenchmark.NewMultiResultWriter(resultWriters), nil
}

package output

import (
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"sync"

	"github.com/christophwitzko/master-thesis/pkg/benchmark"
)

type NewChunkFunc func(lastResult *benchmark.Result, newResult *benchmark.Result) bool

func NoChunkFn(lastResult *benchmark.Result, newResult *benchmark.Result) bool {
	return false
}

func SuiteChunkFn(lastResult *benchmark.Result, newResult *benchmark.Result) bool {
	if lastResult == nil {
		return true
	}
	return lastResult.S != newResult.S
}

func BenchFnChunkFn(lastResult *benchmark.Result, newResult *benchmark.Result) bool {
	// always chunk a new suite run
	if SuiteChunkFn(lastResult, newResult) {
		return true
	}
	return lastResult.Function.String() != newResult.Function.String()
}

type Output struct {
	Schema     string
	Type       string
	path       string
	Parameters url.Values

	writer     io.WriteCloser
	encoder    ResultEncoder
	writeMutex sync.Mutex

	chunked    bool
	newChunkFn func(previous *benchmark.Result, current *benchmark.Result) bool
	chunkIndex uint64
	lastResult *benchmark.Result
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
	if !IsValidSchema(schema) {
		return nil, fmt.Errorf("unsupported output schema: %s", schema)
	}
	params := parsedPath.Query()
	o := &Output{
		Schema:     schema,
		Type:       outputType,
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
		case "suite":
			o.newChunkFn = SuiteChunkFn
		case "benchFn":
			o.newChunkFn = BenchFnChunkFn
		case "no-chunk":
			fallthrough
		default:
			o.newChunkFn = NoChunkFn
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

func (o *Output) Write(result benchmark.Result) error {
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

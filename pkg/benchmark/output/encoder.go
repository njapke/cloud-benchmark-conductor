package output

import (
	"bytes"
	"encoding/csv"
	"encoding/json"

	"github.com/christophwitzko/master-thesis/pkg/benchmark"
)

type ResultEncoder interface {
	Encode(result benchmark.Result) ([]byte, error)
}

type JSONResultEncoder struct {
	buffer  *bytes.Buffer
	encoder *json.Encoder
}

func NewJSONResultEncoder(config *Output) (ResultEncoder, error) {
	buffer := &bytes.Buffer{}
	return &JSONResultEncoder{
		buffer:  buffer,
		encoder: json.NewEncoder(buffer),
	}, nil
}

func (j *JSONResultEncoder) Encode(result benchmark.Result) ([]byte, error) {
	j.buffer.Reset()
	if err := j.encoder.Encode(result); err != nil {
		return nil, err
	}
	return j.buffer.Bytes(), nil
}

type CSVResultEncoder struct {
	buffer           *bytes.Buffer
	csvWriter        *csv.Writer
	hasWrittenHeader bool
}

func NewCSVResultEncoder(config *Output) (ResultEncoder, error) {
	buffer := &bytes.Buffer{}
	csvEncoder := &CSVResultEncoder{
		buffer:    buffer,
		csvWriter: csv.NewWriter(buffer),
	}
	return csvEncoder, nil
}

func (c *CSVResultEncoder) Encode(result benchmark.Result) ([]byte, error) {
	c.buffer.Reset()
	if !c.hasWrittenHeader {
		err := c.csvWriter.Write(benchmark.CSVOutputHeader)
		if err != nil {
			return nil, err
		}
	}
	if err := c.csvWriter.Write(result.Record()); err != nil {
		return nil, err
	}
	c.csvWriter.Flush()
	return c.buffer.Bytes(), nil
}

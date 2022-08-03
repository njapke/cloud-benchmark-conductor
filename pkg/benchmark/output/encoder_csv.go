package output

import (
	"bytes"
	"encoding/csv"

	"github.com/christophwitzko/master-thesis/pkg/benchmark"
)

type csvResultEncoder struct {
	buffer           *bytes.Buffer
	csvWriter        *csv.Writer
	hasWrittenHeader bool
}

func newCsvResultEncoder(config *Output) (ResultEncoder, error) {
	noCSVHeader := false
	if config.Parameters.Get("no-csv-header") == "true" {
		noCSVHeader = true
	}
	buffer := &bytes.Buffer{}
	csvEncoder := &csvResultEncoder{
		buffer:           buffer,
		csvWriter:        csv.NewWriter(buffer),
		hasWrittenHeader: noCSVHeader,
	}
	return csvEncoder, nil
}

func (c *csvResultEncoder) Encode(result benchmark.Result) ([]byte, error) {
	c.buffer.Reset()
	if !c.hasWrittenHeader {
		err := c.csvWriter.Write(benchmark.CSVOutputHeader)
		if err != nil {
			return nil, err
		}
		c.hasWrittenHeader = true
	}
	if err := c.csvWriter.Write(result.Record()); err != nil {
		return nil, err
	}
	c.csvWriter.Flush()
	return c.buffer.Bytes(), nil
}

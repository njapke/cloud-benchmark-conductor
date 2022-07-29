package benchmark

import (
	"encoding/csv"
	"encoding/json"
	"io"
)

type ResultWriter interface {
	Write(result Result) error
}

type CSVResultWriter struct {
	writer *csv.Writer
}

func NewCSVResultWriter(writer io.Writer) *CSVResultWriter {
	return &CSVResultWriter{
		writer: csv.NewWriter(writer),
	}
}

func (c *CSVResultWriter) Write(result Result) error {
	return c.WriteRaw(result.Record())
}

func (c *CSVResultWriter) WriteRaw(record []string) error {
	if err := c.writer.Write(record); err != nil {
		return err
	}
	c.writer.Flush()
	return nil
}

type JSONResultWriter struct {
	encoder *json.Encoder
}

func NewJSONResultWriter(writer io.Writer) *JSONResultWriter {
	return &JSONResultWriter{
		encoder: json.NewEncoder(writer),
	}
}

func (j *JSONResultWriter) Write(result Result) error {
	return j.encoder.Encode(result)
}

package output

import (
	"bytes"
	"encoding/json"

	"github.com/christophwitzko/master-thesis/pkg/benchmark"
)

type jsonResultEncoder struct {
	buffer  *bytes.Buffer
	encoder *json.Encoder
}

func newJSONResultEncoder(config *Output) (ResultEncoder, error) {
	buffer := &bytes.Buffer{}
	return &jsonResultEncoder{
		buffer:  buffer,
		encoder: json.NewEncoder(buffer),
	}, nil
}

func (j *jsonResultEncoder) Encode(result benchmark.Result) ([]byte, error) {
	j.buffer.Reset()
	if err := j.encoder.Encode(result); err != nil {
		return nil, err
	}
	return j.buffer.Bytes(), nil
}

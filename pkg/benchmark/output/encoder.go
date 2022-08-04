package output

import (
	"fmt"

	"github.com/christophwitzko/master-thesis/pkg/benchmark"
)

type ResultEncoder interface {
	Encode(result benchmark.Result) ([]byte, error)
}

type EncoderFactory func(config *Output) (ResultEncoder, error)

var encoders = map[string]EncoderFactory{
	"json": newJSONResultEncoder,
	"csv":  newCsvResultEncoder,
}

func NewEncoder(config *Output) (ResultEncoder, error) {
	encFactory, ok := encoders[config.Type]
	if !ok {
		return nil, fmt.Errorf("unsupported output type: %s", config.Type)
	}
	return encFactory(config)
}

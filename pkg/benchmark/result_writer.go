package benchmark

import "github.com/hashicorp/go-multierror"

type ResultWriter interface {
	Write(result Result) error
	Close() error
}

type multiResultWriter struct {
	writers []ResultWriter
}

func NewMultiResultWriter(writers []ResultWriter) ResultWriter {
	return &multiResultWriter{
		writers: writers,
	}
}

func (m *multiResultWriter) Write(result Result) error {
	for _, writer := range m.writers {
		if err := writer.Write(result); err != nil {
			return err
		}
	}
	return nil
}

func (m *multiResultWriter) Close() error {
	var mErr error
	for _, w := range m.writers {
		if err := w.Close(); err != nil {
			mErr = multierror.Append(mErr, err)
		}
	}
	return mErr
}

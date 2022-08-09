package assets

import (
	"bytes"
	"crypto/sha512"
	_ "embed"
	"encoding/hex"
)

type Binary []byte

func (b Binary) GetHash() string {
	hash := sha512.New()
	hash.Write(b)
	return hex.EncodeToString(hash.Sum(nil))
}

func (b Binary) GetReader() *bytes.Reader {
	return bytes.NewReader(b)
}

//go:embed mb-runner_linux_amd64
var MicrobenchmarkRunner Binary

//go:embed app-runner_linux_amd64
var ApplicationRunner Binary

//go:embed app-bench-runner_linux_amd64
var ApplicationBenchmarkRunner Binary

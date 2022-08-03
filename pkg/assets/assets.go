package assets

import (
	"crypto/sha512"
	_ "embed"
	"encoding/hex"
)

//go:embed mb-runner_linux_amd64
var MicrobenchmarkRunnerBinaryLinuxAmd64 []byte
var MicrobenchmarkRunnerBinaryLinuxAmd64Hash string

func init() {
	hash := sha512.New()
	hash.Write(MicrobenchmarkRunnerBinaryLinuxAmd64)
	MicrobenchmarkRunnerBinaryLinuxAmd64Hash = hex.EncodeToString(hash.Sum(nil))
}

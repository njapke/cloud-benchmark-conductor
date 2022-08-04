package assets

import (
	"crypto/sha512"
	_ "embed"
	"encoding/hex"
)

//go:embed mb-runner_linux_amd64
var MicrobenchmarkRunnerBinaryLinuxAmd64 []byte

func GetMicrobenchmarkRunnerBinaryLinuxAmd64Hash() string {
	hash := sha512.New()
	hash.Write(MicrobenchmarkRunnerBinaryLinuxAmd64)
	return hex.EncodeToString(hash.Sum(nil))
}

package assets

import (
	"crypto/sha512"
	_ "embed"
	"encoding/hex"
)

func getHash(data []byte) string {
	hash := sha512.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

//go:embed mb-runner_linux_amd64
var MicrobenchmarkRunnerBinaryLinuxAmd64 []byte

func GetMicrobenchmarkRunnerBinaryLinuxAmd64Hash() string {
	return getHash(MicrobenchmarkRunnerBinaryLinuxAmd64)
}

//go:embed app-runner_linux_amd64
var ApplicationRunnerBinaryLinuxAmd64 []byte

func GetApplicationRunnerBinaryLinuxAmd64Hash() string {
	return getHash(ApplicationRunnerBinaryLinuxAmd64)
}

//go:embed app-bench-runner_linux_amd64
var ApplicationBenchmarkRunnerBinaryLinuxAmd64 []byte

func GetApplicationBenchmarkRunnerBinaryLinuxAmd64Hash() string {
	return getHash(ApplicationBenchmarkRunnerBinaryLinuxAmd64)
}

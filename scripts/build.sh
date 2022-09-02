#!/usr/bin/env bash

set -euo pipefail

k6Version="0.39.0"
k6DownloadURL="https://github.com/grafana/k6/releases/download/v${k6Version}/k6-v${k6Version}-linux-amd64.tar.gz"
k6AssetPath="./pkg/assets/k6_linux_amd64"

[[ -f "$k6AssetPath" ]] || {
    echo "downloading k6..."
    curl -SL "$k6DownloadURL" | tar xvf - --strip-components=1 "k6-v${k6Version}-linux-amd64/k6"
    mv ./k6 "$k6AssetPath"
}

export CGO_ENABLED=0
function gobuild() {
    go build -trimpath -ldflags="-s -w -extldflags '-static'" "$@"
}

echo "building microbenchmark-runner..."
GOOS=linux GOARCH=amd64 gobuild -o ./pkg/assets/mb-runner_linux_amd64 ./cmd/microbenchmark-runner/

echo "building application-runner..."
GOOS=linux GOARCH=amd64 gobuild -o ./pkg/assets/app-runner_linux_amd64 ./cmd/application-runner/

echo "building application-benchmark-runner..."
GOOS=linux GOARCH=amd64 gobuild -o ./pkg/assets/app-bench-runner_linux_amd64 ./cmd/application-benchmark-runner/

echo "building cloud-benchmark-conductor..."
gobuild -o ./cloud-benchmark-conductor ./cmd/cloud-benchmark-conductor/

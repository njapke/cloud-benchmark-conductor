#!/usr/bin/env bash

set -euo pipefail

export CGO_ENABLED=0

function gobuild() {
    go build -trimpath -ldflags="-s -w -extldflags '-static'" $@
}

echo "building microbenchmark-runner..."
GOOS=linux GOARCH=amd64 gobuild -o ./pkg/assets/mb-runner_linux_amd64 ./cmd/microbenchmark-runner/

echo "building application-runner..."
GOOS=linux GOARCH=amd64 gobuild -o ./pkg/assets/app-runner_linux_amd64 ./cmd/application-runner/

echo "building application-benchmark-runner..."
GOOS=linux GOARCH=amd64 gobuild -o ./pkg/assets/app-bench-runner_linux_amd64 ./cmd/application-benchmark-runner/

echo "building cloud-benchmark-conductor..."
gobuild -o ./cloud-benchmark-conductor ./cmd/cloud-benchmark-conductor/

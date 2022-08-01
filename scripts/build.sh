#!/usr/bin/env bash

set -euo pipefail

export CGO_ENABLED=0

function gobuild() {
    go build -trimpath -ldflags="-s -w -extldflags '-static'" $@
}

GOOS=linux GOARCH=amd64 gobuild -o ./pkg/assets/mb-runner_linux_amd64 ./cmd/microbenchmark-runner/

gobuild -o ./cloud-benchmark-conductor ./cmd/cloud-benchmark-conductor/

#!/usr/bin/env bash

set -euo pipefail

export CGO_ENABLED=0
go build -o ./gocg-overlap ./cmd/overlap/
go build -o ./gocg-minimization ./cmd/minimization/
go build -o ./gocg-recommendation ./cmd/recommendation/

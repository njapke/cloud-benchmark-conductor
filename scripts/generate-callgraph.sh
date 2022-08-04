#!/usr/bin/env bash

set -euo pipefail

# using https://pkg.go.dev/golang.org/x/tools@v0.1.11/cmd/callgraph

callgraph -algo pta -format digraph ./ > callgraph.txt

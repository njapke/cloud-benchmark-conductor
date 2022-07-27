#!/usr/bin/env bash

set -euo pipefail


for i in {1..3} ; do
  ./microbenchmark-runner -i ./bench-fns.json -r "$i" >> results.csv
done

#!/usr/bin/env bash

set -euo pipefail
shopt -s globstar

for filename in $1/*; do
  outfile="${filename}/combined.csv"
  echo "creating $outfile"
  cat ${filename}/run-* > $outfile
done

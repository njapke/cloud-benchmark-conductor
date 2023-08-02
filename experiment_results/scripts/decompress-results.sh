#!/usr/bin/env bash

set -euo pipefail
shopt -s globstar

for filename in $1/**/*.gz; do
  echo "decompress $filename"
  gzip -d $filename
done

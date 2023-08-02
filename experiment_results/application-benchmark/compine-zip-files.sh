#!/usr/bin/env bash

set -euo pipefail
shopt -s globstar

cat application-benchmark-config.zip.?? > application-benchmark-config.zip
cat application-benchmark-data.zip.?? > application-benchmark-data.zip


#!/usr/bin/env bash

set -euo pipefail

rm -f results.csv
for i in {1..3} ; do
  ./microbenchmark-runner \
    --v1 main \
    --v2 perf-issue-request-id \
    --git-repository='https://github.com/christophwitzko/flight-booking-service.git' \
    --exclude-filter="^chi.*$" \
    --include-filter="^service.*$" \
    -o "-?no-csv-header=true" \
    --run "$i" >> results.csv
done

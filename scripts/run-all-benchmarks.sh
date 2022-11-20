#!/usr/bin/env bash

set -euo pipefail

# application benchmark
./cloud-benchmark-conductor config
./cloud-benchmark-conductor ab --application-v2 main
./cloud-benchmark-conductor cleanup
./cloud-benchmark-conductor ab --application-v2 perf-issue-clean-path
./cloud-benchmark-conductor cleanup
./cloud-benchmark-conductor ab --application-v2 perf-issue-request-id
./cloud-benchmark-conductor cleanup
./cloud-benchmark-conductor ab --application-v2 perf-issue-basic-auth
./cloud-benchmark-conductor cleanup


# full microbenchmark suite
full_mb_flags='--microbenchmark-exclude-filter ^chi.*$'
#./cloud-benchmark-conductor config $full_mb_flags
./cloud-benchmark-conductor mb $full_mb_flags --microbenchmark-v2 main
./cloud-benchmark-conductor cleanup
./cloud-benchmark-conductor mb $full_mb_flags --microbenchmark-v2 perf-issue-clean-path
./cloud-benchmark-conductor cleanup
./cloud-benchmark-conductor mb $full_mb_flags --microbenchmark-v2 perf-issue-request-id
 ./cloud-benchmark-conductor cleanup
./cloud-benchmark-conductor mb $full_mb_flags --microbenchmark-v2 perf-issue-basic-auth
./cloud-benchmark-conductor cleanup


# optimized microbenchmark suite
opt_mb_flags='--microbenchmark-output gs://cbc-results/{{.Name}}/mb-opt-{{.V1}}-{{.V2}}-{{.Timestamp}}/run-{{.RunIndex}}.csv?chunked=true&no-csv-header=true --microbenchmark-function service.BenchmarkRequestFlights --microbenchmark-function service.BenchmarkHandlerGetFlightSeats --microbenchmark-function service.BenchmarkRequestDestinations --microbenchmark-function service.BenchmarkRequestCreateBooking'
#./cloud-benchmark-conductor config $opt_mb_flags
./cloud-benchmark-conductor mb $opt_mb_flags --microbenchmark-v2 main
./cloud-benchmark-conductor cleanup
./cloud-benchmark-conductor mb $opt_mb_flags --microbenchmark-v2 perf-issue-clean-path
./cloud-benchmark-conductor cleanup
./cloud-benchmark-conductor mb $opt_mb_flags --microbenchmark-v2 perf-issue-request-id
./cloud-benchmark-conductor cleanup
./cloud-benchmark-conductor mb $opt_mb_flags --microbenchmark-v2 perf-issue-basic-auth
./cloud-benchmark-conductor cleanup

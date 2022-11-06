# master-thesis

## Included Tools

### application-benchmark-runner
Runs artillery/k6 benchmarks.

### application-runner
Runs two different versions of an application simultaneous.

### microbenchmark-runner
Runs microbenchmarks using RMIT (Randomized Multiple Interleaved Trials).

### cloud-benchmark-conductor
Uses the tools form above to run micro and application benchmarks in the cloud.

## Running the benchmarks
```bash
./scripts/build.sh

# application benchmark
./cloud-benchmark-conductor ab --application-v2 main && ./cloud-benchmark-conductor cleanup
./cloud-benchmark-conductor ab --application-v2 perf-issue-clean-path && ./cloud-benchmark-conductor cleanup
./cloud-benchmark-conductor ab --application-v2 perf-issue-request-id && ./cloud-benchmark-conductor cleanup
./cloud-benchmark-conductor ab --application-v2 perf-issue-basic-auth && ./cloud-benchmark-conductor cleanup

# full microbenchmark suite
./cloud-benchmark-conductor mb --microbenchmark-v2 main && ./cloud-benchmark-conductor cleanup
./cloud-benchmark-conductor mb --microbenchmark-v2 perf-issue-clean-path && ./cloud-benchmark-conductor cleanup
./cloud-benchmark-conductor mb --microbenchmark-v2 perf-issue-request-id && ./cloud-benchmark-conductor cleanup
./cloud-benchmark-conductor mb --microbenchmark-v2 perf-issue-basic-auth && ./cloud-benchmark-conductor cleanup

# configure optimized microbenchmark suite
./cloud-benchmark-conductor mb --microbenchmark-v2 main && ./cloud-benchmark-conductor cleanup
./cloud-benchmark-conductor mb --microbenchmark-v2 perf-issue-clean-path && ./cloud-benchmark-conductor cleanup
./cloud-benchmark-conductor mb --microbenchmark-v2 perf-issue-request-id && ./cloud-benchmark-conductor cleanup
./cloud-benchmark-conductor mb --microbenchmark-v2 perf-issue-basic-auth && ./cloud-benchmark-conductor cleanup

# download results
gsutil cp -r gs://cbc-results/fbs ./results/

# combine mb results
./scripts/combine-mb-results.sh ./results
```

## Profiling
```bash
./application-benchmark-runner \
  --reference main \
  --git-repository='https://github.com/christophwitzko/flight-booking-service.git' \
  --target v1=127.0.0.1:3000 \
  --tool k6 \
  --results-output gs://cbc-results/ab-profiles \
  --profiling
```

```bash
./microbenchmark-runner \
  --v1 main \
  --v2 main \
  --git-repository='https://github.com/christophwitzko/flight-booking-service.git' \
  --exclude-filter="^chi.*$" \
  --profiling-gcs-output gs://cbc-results/mb-profiles \
  --profiling
```

## Running gocg
```bash
rm -rf profiling && mkdir profiling
gsutil cp -r gs://cbc-results/mb-profiles ./profiling
gsutil cp -r gs://cbc-results/ab-profiles ./profiling

./scripts/fix-dot-files.sh ./profiling
./scripts/run-gocg.sh
open ./gocg-results/ab-profiles_struct_node_overlap_mins-GreedySystem.csv
```

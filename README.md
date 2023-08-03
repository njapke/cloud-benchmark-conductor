# The Early Microbenchmark Catches the Bug -- Studying Performance Issues Using Micro- and Application Benchmarks

This repository contains the code and results described in our paper on performance issue detection using application benchmarks and microbenchmarks. It includes tools to automate application benchmarks and microbenchmarks of our [testbed application](https://github.com/njapke/flight-booking-service). Experiment configurations and result data can be found in [experiment_results](./experiment_results). Analysis scripts are contained in [analysis_scripts](./analysis_scripts).

## Citation

If you use this software in a publication, please cite it as:

### Text

N. Japke, C. Witzko, M. Grambow and D. Bermbach, **The Early Microbenchmark Catches the Bug -- Studying Performance Issues Using Micro- and Application Benchmarks**, 2023.

### BibTeX

```bibtex
@article{japke2023studyingperformance,
    title = "The Early Microbenchmark Catches the Bug -- Studying Performance Issues Using Micro- and Application Benchmarks",
    author = "Japke, Nils and Witzko, Christoph and Grambow, Martin and Bermbach, David",
    year = 2023
}
```

For a full list of publications, please see [our website](https://www.tu.berlin/en/mcc/research/publications).

## Included Tools

### [application-benchmark-runner](./cmd/application-benchmark-runner/)
Runs artillery/k6 benchmarks.

### [application-runner](./cmd/application-runner/)
Runs two different versions of an application simultaneous.

### [microbenchmark-runner](./cmd/microbenchmark-runner/)
Runs microbenchmarks using RMIT (Randomized Multiple Interleaved Trials).

### [cloud-benchmark-conductor](./cmd/cloud-benchmark-conductor/)
Uses the tools form above to run micro and application benchmarks in the cloud.

### [gocg](./tools/gocg/)
Not used in our paper. A description of this tool can be found in a separate readme file.

## Running the benchmarks
```bash
# build toolchain
./scripts/build.sh

# execute all 12 benchmarks
./scripts/run-all-benchmarks.sh

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

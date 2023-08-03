# gocg
The gocg tool is used to calculate an optimized microbenchmark suite. It is a fork of [the original implementation by Grambow et al.](https://depositonce.tu-berlin.de/items/a2820b75-a5ca-4a75-a37b-ac489a1fd330) to support Go generics.

## Building & Running gocg

**Building gocg**
```bash
cd ./tools/gocg && ./build.sh && cd -
```

**Running gocg**
```bash
rm -rf profiling && mkdir profiling
gsutil cp -r gs://cbc-results/mb-profiles ./profiling
gsutil cp -r gs://cbc-results/ab-profiles ./profiling

./scripts/fix-dot-files.sh ./profiling
./scripts/run-gocg.sh
open ./gocg-results/ab-profiles_struct_node_overlap_mins-GreedySystem.csv
```

#!/usr/bin/env bash

set -euo pipefail

rm -f ./gocg-results/*

./tools/gocg/gocg-overlap github.com/christophwitzko/flight-booking-service \
  ./profiling/ab-profiles/ \
  ./profiling/mb-profiles/ \
  ./gocg-results/

./tools/gocg/gocg-minimization github.com/christophwitzko/flight-booking-service \
  ./profiling/ab-profiles/ \
  ./profiling/mb-profiles/ \
  ./gocg-results/

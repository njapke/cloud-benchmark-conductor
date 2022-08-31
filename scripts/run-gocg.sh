#!/usr/bin/env bash

set -euo pipefail

rm -f ./gocg-results/*
../mt-gocg/gocg-overlap github.com/christophwitzko/flight-booking-service ./profiling/ab-profiles/ ./profiling/mb-profiles/ ./gocg-results/
../mt-gocg/gocg-minimization github.com/christophwitzko/flight-booking-service ./profiling/ab-profiles/ ./profiling/mb-profiles/ ./gocg-results/

# Analysis Scripts
This folder contains all Python scripts we used for statistical analysis. All of these were authored and run in the Spyder IDE, but should also work in an ipython session.

Beware: In many scripts you need to specify input and output folders, before they can be run!

## preprocessing_data_median_app.py
This script preprocesses the application benchmark result data, and generates scatterplots of the data to visualize the results. These should now only contain a slice, where both SUTs were running at the same time (w/ removed warmup).

## perf_analysis_median_ratio_app.py
This script calculates median performance ratios, confidence intervals, and found performance changes for the preprocessed application benchmark data. It also generates plots showing the performance ratio by performance issue severity with the confidence interval.

## perf_analysis_median_ratio_mb.py
This script calculates median performance ratios, confidence intervals, and found performance changes for the microbenchmark data. It also generates plots showing the performance ratio by performance issue severity with the confidence interval.

## variability_calc.py
This script analyzes the relative confidence interval width of all experiment runs, and generates a boxplot to visualize their distribution.

## stat_functions.py
This file contains implementations for bootstrap confidence intervals. These functions are used in the other scripts.

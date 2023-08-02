#!/usr/bin/env python3
# -*- coding: utf-8 -*-

# This file contains some statistical helper functions to generate bootstrap
# confidence intervals.

import numpy as np

# Initialize RNG
rng = np.random.default_rng(42)

# Coefficient of Variation
def cv(data, **kw):
    m = data.mean()
    return np.sqrt(1/(len(data)) * ((data - m)**2).sum() ) / m

# Relative Median Absolute Deviation
def rmad(data, **kw):
    m = np.median(data)
    return np.median(np.absolute(data - m)) / m

# Relative Confidence Interval Width (mean, percentile bootstrap)
def rciw_mean_p(data, it = 10000, cl = 99):
    p = ci_bootstrap_mean_p(data, it, cl)
    return np.absolute(p[1] - p[0]) / data.mean()

# CI bounds of mean with percentile bootstrap
def ci_bootstrap_mean_p(data, it = 10000, cl = 99):
    bs_dist = np.mean(rng.choice(data, (it, len(data))), axis=1)
    
    lower = (100 - cl) / 2
    upper = cl + lower
    return np.percentile(bs_dist, [lower, upper]) # CI bounds

# Relative Confidence Interval Width (median, percentile bootstrap)
def rciw_median_p(data, it = 10000, cl = 99):
    p = ci_bootstrap_mean_p(data, it, cl)
    return np.absolute(p[1] - p[0]) / np.median(data)

# CI bounds of median with percentile bootstrap
def ci_bootstrap_median_p(data, it = 10000, cl = 99):
    bs_dist = np.median(rng.choice(data, (it, len(data))), axis=1)
    
    lower = (100 - cl) / 2
    upper = cl + lower
    return np.percentile(bs_dist, [lower, upper]) # CI bounds

# Bootstrap standard error for the median
def se_bootstrap_median(data, it = 10000):
    bs_dist = np.median(rng.choice(data, (it, len(data))), axis=1)
    
    bs_mean = np.mean(bs_dist)
    
    return np.sqrt(np.sum(np.power((bs_dist - bs_mean), 2)) / (it-1))

# CI bounds of mean ratios of 2 versions with percentile bootstrap
def ci_bootstrap_perf_change_mean(data_v1, data_v2, it = 10000, cl = 99):
    bs_dist_v1 = np.mean(rng.choice(data_v1, (it, len(data_v1))), axis=1)
    bs_dist_v2 = np.mean(rng.choice(data_v2, (it, len(data_v2))), axis=1)
    bs_dist_ratio = bs_dist_v2 / bs_dist_v1
    
    lower = (100 - cl) / 2
    upper = cl + lower
    return np.percentile(bs_dist_ratio, [lower, upper]) # CI bounds

# CI bounds of mean ratios of 2 versions with percentile bootstrap
def ci_bootstrap_perf_change_median(data_v1, data_v2, it = 10000, cl = 99):
    bs_dist_v1 = np.median(rng.choice(data_v1, (it, len(data_v1))), axis=1)
    bs_dist_v2 = np.median(rng.choice(data_v2, (it, len(data_v2))), axis=1)
    bs_dist_ratio = bs_dist_v2 / bs_dist_v1
    
    lower = (100 - cl) / 2
    upper = cl + lower
    return np.percentile(bs_dist_ratio, [lower, upper]) # CI bounds

def in_interval(x,low,high):
    return x >= low and x <= high

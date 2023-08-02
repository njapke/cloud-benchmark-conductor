#!/usr/bin/env python3
# -*- coding: utf-8 -*-

#%%
# This script generates the performance analysis out of the result data
# from the microbenchmark experiment.
#
# Also generates extra plots, showing the median performance ratio + confidence interval.
#
# Can be run using ipython, or the Spyder IDE.

from pathlib import Path
import numpy as np
import pandas as pd
import stat_functions as st
import matplotlib.pyplot as plt

# set this path to the unzipped microbenchmark result data
data_path = "path/to/microbenchmark/experiment/data"

list_of_perf_issues = ["basic-auth-{}","clean-path-{}","request-id-{}",]

list_of_severities = [0,1,2,4,8,16,32,64,128,256,512,1024,2048]

res = []
for perf in list_of_perf_issues:
    for sev in list_of_severities:
        path = data_path + "/" + perf.format(sev) + "/combined.csv"
        df = pd.read_csv(path, names=["R-S-I","Benchmark","Version","FileName","Invocations","sec/op","B/op","allocs/op"])
        
        # get all benchmarks
        benchmarks = np.array(df["Benchmark"].drop_duplicates())
        
        for b in benchmarks:
            df_v1 = df[(df["Benchmark"] == b) & (df["Version"] == 1)]
            df_v2 = df[(df["Benchmark"] == b) & (df["Version"] == 2)]
            
            data_v1 = np.array(df_v1["sec/op"])
            data_v2 = np.array(df_v2["sec/op"])
            
            median_ratio = np.median(data_v2) / np.median(data_v1)
            ci_median_ratio = st.ci_bootstrap_perf_change_median(data_v1, data_v2)
            perf_change_found = not st.in_interval(1, ci_median_ratio[0], ci_median_ratio[1])
            
            res.append([b,perf[:-3],sev,median_ratio,ci_median_ratio[0],ci_median_ratio[1],perf_change_found])

# save results
result_df = pd.DataFrame(res, columns=["Benchmark","Perf Issue","Severity","Median Ratio","CI Lower","CI Upper","Perf Change Found"])
result_df.to_csv("results_mb.csv",index=False)


#%%
# prevent figures from showing (as they are saved)
plt.ioff()
# plt.ion()

# variables
output_folder = "plots_mb/"

# get all benchmarks
benchmarks = np.array(result_df["Benchmark"].drop_duplicates())

for perf in list_of_perf_issues:
    output_folder_perf = output_folder + perf[:-3] + "/"
    Path(output_folder_perf).mkdir(parents=True, exist_ok=True)
    
    for b in benchmarks:
        df_plot = result_df[(result_df["Benchmark"] == b) & (result_df["Perf Issue"] == perf[:-3])]
        
        fig = plt.figure(figsize=(12,8))
        ax = plt.axes(xlim=(0,2048))
        
        ax.set_title("results for " + b)
        ax.set_ylabel("median ratio")
        
        # plot
        ax.plot(df_plot["Severity"], df_plot["Median Ratio"], "b-")
        ax.fill_between(df_plot["Severity"], df_plot["CI Lower"], df_plot["CI Upper"], color="blue", alpha=0.2)
        
        out_name = b.replace("/","-").replace(".","-") + "_" + perf[:-3]
        
        fig.savefig(output_folder_perf + out_name + ".png")
        plt.close(fig)

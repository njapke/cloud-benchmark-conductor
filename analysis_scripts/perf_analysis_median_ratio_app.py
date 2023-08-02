#!/usr/bin/env python3
# -*- coding: utf-8 -*-

#%%
# This script generates the performance analysis out of the preprocessed data
# from an application benchmark. Run it using ipython in the same directory,
# where the output of "preprocessing_data_median_app.py" was generated.
#
# Also generates extra plots, showing the median performance ratio + confidence interval.
#
# Can be run using ipython, or the Spyder IDE.

from pathlib import Path
import numpy as np
import pandas as pd
import stat_functions as st
import matplotlib.pyplot as plt
import re

list_of_perf_issues = ["basic-auth-{}","clean-path-{}","request-id-{}",]

col_dtypes = {"timestamp":"int64","name":"str","count":"int64","median_duration":"float64","version":"int64","perf_issue":"str","severity":"int64"}

df_full = pd.read_csv("df_full.csv",dtype=col_dtypes)

#%%
# get all benchmarks
benchmarks = np.array(df_full["name"].drop_duplicates())

list_of_severities = [0,1,2,4,8,16,32,64,128,256,512,1024,2048]

res = []
for perf in list_of_perf_issues:
    for sev in list_of_severities:
        for b in benchmarks:
            df_v1 = df_full[(df_full["name"] == b)
                              & (df_full["severity"] == sev)
                              & (df_full["perf_issue"] == perf)
                              & (df_full["version"] == 1)]
            df_v2 = df_full[(df_full["name"] == b)
                              & (df_full["severity"] == sev)
                              & (df_full["perf_issue"] == perf)
                              & (df_full["version"] == 2)]
            
            # calculate timestamp, where the first app bench terminates
            a = max(df_v2["timestamp"]) - max(df_v1["timestamp"])
            print("PerfIssue: {}, Sev: {}, Bench: {}, Diff is {}".format(perf,sev,b,a))
            end_exp = min(max(df_v1["timestamp"]), max(df_v2["timestamp"])) - 60
            
            # remove warmup (first 60s)
            df_v1 = df_v1[df_v1["timestamp"] >= 60]
            df_v2 = df_v2[df_v2["timestamp"] >= 60]
            # remove wind-down (last 60s, before one version terminates)
            df_v1 = df_v1[df_v1["timestamp"] <= end_exp]
            df_v2 = df_v2[df_v2["timestamp"] <= end_exp]
            
            
            data_v1 = np.array(df_v1["median_duration"])
            data_v2 = np.array(df_v2["median_duration"])
            
            median_ratio = np.median(data_v2) / np.median(data_v1)
            ci_median_ratio = st.ci_bootstrap_perf_change_median(data_v1, data_v2)
            perf_change_found = not st.in_interval(1, ci_median_ratio[0], ci_median_ratio[1])
            
            res.append([b,perf[:-3],sev,median_ratio,ci_median_ratio[0],ci_median_ratio[1],perf_change_found])

# save results
result_app_df = pd.DataFrame(res, columns=["Name","Perf Issue","Severity","Median Ratio","CI Lower","CI Upper","Perf Change Found"])
result_app_df.to_csv("results_app.csv",index=False)


#%%
# prevent figures from showing (as they are saved)
plt.ioff()
# plt.ion()

# variables
output_folder = "plots_app/"

# get all benchmarks
benchmarks = np.array(result_app_df["Name"].drop_duplicates())

for perf in list_of_perf_issues:
    output_folder_perf = output_folder + perf[:-3] + "/"
    Path(output_folder_perf).mkdir(parents=True, exist_ok=True)
    
    for b in benchmarks:
        df_plot = result_app_df[(result_app_df["Name"] == b) & (result_app_df["Perf Issue"] == perf[:-3])]
        
        fig = plt.figure(figsize=(12,8))
        ax = plt.axes(xlim=(0,2048))
        
        ax.set_title("results for " + b)
        ax.set_ylabel("median ratio")
        
        # plot
        ax.plot(df_plot["Severity"], df_plot["Median Ratio"], "b-")
        ax.fill_between(df_plot["Severity"], df_plot["CI Lower"], df_plot["CI Upper"], color="blue", alpha=0.2)
        
        out_name = re.sub("[\$\{\}/\?\=]+", "-", b) + "_" + perf[:-3]
        
        fig.savefig(output_folder_perf + out_name + ".png")
        plt.close(fig)

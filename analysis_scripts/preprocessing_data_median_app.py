#!/usr/bin/env python3
# -*- coding: utf-8 -*-

#%%
# This script preprocesses the result data from an application benchmark.
# For each timestamp (second), we keep the median response latency.
# This reduces the size of the dataset, but also reduces the noise.
# Overall, this only has negligible impact on the final results, but
# speeds up further computations.
#
# Also generates scatterplots of the preprocessed result data.
#
# Can be run using ipython, or the Spyder IDE.

from pathlib import Path
import numpy as np
import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
import re

# set this path to the unzipped application benchmark result data
data_path = "path/to/application/benchmark/experiment/data"

list_of_perf_issues = ["basic-auth-{}","clean-path-{}","request-id-{}",]

col_dtypes = {"metric_name":"str","timestamp":"int64","metric_value":"float64","check":"str","error":"str","error_code":"str","expected_response":"str","iter":"str","method":"str","name":"str","scenario":"str","status":"str","url":"str","extra_tags":"str"}

list_of_severities = [0,1,2,4,8,16,32,64,128,256,512,1024,2048]

#%%
res = []
df_full = None
for perf in list_of_perf_issues:
    for sev in list_of_severities:
        print("Severity: {}".format(sev))
        path = data_path + "/" + perf.format(sev)
        
        df_v1 = pd.read_csv(path + "/v1.csv",dtype=col_dtypes)
        df_v1 = df_v1[df_v1["metric_name"] == "http_req_duration"]
        df_v1.drop(["metric_name","iter","check","error","error_code","expected_response","scenario","status","method","url","extra_tags"], axis=1, inplace=True)
        df_v1 = df_v1.groupby(["timestamp","name"]).agg(count=("metric_value","count"),median_duration=("metric_value","median")).reset_index()
        df_v1["version"] = 1
        df_v1["perf_issue"] = perf
        df_v1["severity"] = sev
        
        df_v2 = pd.read_csv(path + "/v2.csv",dtype=col_dtypes)
        df_v2 = df_v2[df_v2["metric_name"] == "http_req_duration"]
        df_v2.drop(["metric_name","iter","check","error","error_code","expected_response","scenario","status","method","url","extra_tags"], axis=1, inplace=True)
        df_v2 = df_v2.groupby(["timestamp","name"]).agg(count=("metric_value","count"),median_duration=("metric_value","median")).reset_index()
        df_v2["version"] = 2
        df_v2["perf_issue"] = perf
        df_v2["severity"] = sev
        
        # set timestamps to start at 0
        min_t_v1 = df_v1["timestamp"].iloc[0]
        min_t_v2 = df_v2["timestamp"].iloc[0]
        df_v1["timestamp"] = df_v1["timestamp"] - min(min_t_v1,min_t_v2)
        df_v2["timestamp"] = df_v2["timestamp"] - min(min_t_v1,min_t_v2)
        
        if df_full is None:
            df_full = pd.concat([df_v1,df_v2])
        else:
            df_full = pd.concat([df_full,df_v1,df_v2])

# save preprocessed data
df_full.to_csv("df_full.csv",index=False)


#%%
# prevent figures from showing (as they are saved)
plt.ioff()
# plt.ion()

# variables
output_folder = "plots_app/"

benchmarks = np.array(df_full["name"].drop_duplicates())

for perf in list_of_perf_issues:
    output_folder_perf = output_folder + perf[:-3] + "/"
    Path(output_folder_perf).mkdir(parents=True, exist_ok=True)
    for sev in list_of_severities:
        for b in benchmarks:
            df_dest = df_full[(df_full["name"] == b)
                              & (df_full["severity"] == sev)
                              & (df_full["perf_issue"] == perf)]
            
            fig = plt.figure(figsize=(10,6))
            ax = plt.axes()
            # sns.scatterplot(data=df_dest, x="timestamp", y="count", hue="version", ax=ax)
            sns.scatterplot(data=df_dest, x="timestamp", y="median_duration", hue="version", ax=ax)
            
            out_name = re.sub("[\$\{\}/\?\=]+", "-", b) + "_" + perf[:-3] + "_" + str(sev)
            
            fig.savefig(output_folder_perf + out_name + ".png")
            plt.close(fig)

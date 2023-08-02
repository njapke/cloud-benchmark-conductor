#!/usr/bin/env python3
# -*- coding: utf-8 -*-

#%%
# This script generates an analysis of the relative confidence interval width (RCIW)
# of the application benchmark and microbenchmark results.
#
# Also generates a boxplot which shows the distribution of RCIW across experiment runs.
#
# Can be run using ipython, or the Spyder IDE.

import numpy as np
import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
import stat_functions as st

list_of_perf_issues = ["basic-auth-{}","clean-path-{}","request-id-{}",]

list_of_severities = [0,1,2,4,8,16,32,64,128,256,512,1024,2048]

#%%
# set this path to the unzipped microbenchmark result data
data_path = "path/to/microbenchmark/experiment/data"

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
            
            rciw_v1 = st.rciw_median_p(data_v1)
            rciw_v2 = st.rciw_median_p(data_v2)
            
            res.append([b,perf,sev,1,rciw_v1])
            res.append([b,perf,sev,2,rciw_v2])
rciw_arr_mb = pd.DataFrame(res, columns=["Name","Perf Issue","Severity","Version","RCIW"])

#%%
# set this path to the folder containing the preprocessed application benchmark data (named "df_full.csv")
data_path = "path/to/folder/containing/preprocessed/application/benchmark/data"

col_dtypes = {"timestamp":"int64","name":"str","count":"int64","median_duration":"float64","version":"int64","perf_issue":"str","severity":"int64"}

df_full = pd.read_csv(data_path+"/df_full.csv",dtype=col_dtypes)

# get all benchmarks
benchmarks = np.array(df_full["name"].drop_duplicates())

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
            
            rciw_v1 = st.rciw_median_p(data_v1)
            rciw_v2 = st.rciw_median_p(data_v2)
            
            res.append([b,perf,sev,1,rciw_v1])
            res.append([b,perf,sev,2,rciw_v2])
rciw_arr_app = pd.DataFrame(res, columns=["Name","Perf Issue","Severity","Version","RCIW"])

#%%
from matplotlib import rcParams

# reset kernel, when editing other figures
rcParams["font.size"] = 20.0 # default is 10.0

# set this to a suitable output folder for the boxplot
out_folder = "plots"

fig = plt.figure(layout="constrained", figsize=(8,6))
ax = plt.axes()

keep = {"${}/bookings" : "E$_1$",
        "${}/destinations" : "E$_2$",
        "${}/flights?from=${}" : "E$_3$",
        "${}/flights/${}/seats" : "E$_4$",
        "service.BenchmarkRequestBookings" : "M$_1$",
        "service.BenchmarkRequestCreateBooking" : "M$_2$",
        "service.BenchmarkRequestDestinations" : "M$_3$",
        "service.BenchmarkRequestFlight" : "M$_4$",
        "service.BenchmarkRequestFlights" : "M$_5$",
        "service.BenchmarkRequestFlightsQuery" : "M$_6$",
        "service.BenchmarkRequestSeats" : "M$_7$"}

rciw_keep_mb = rciw_arr_mb[rciw_arr_mb["Name"].isin(keep.keys())]
rciw_keep_mb.replace(keep, inplace=True)
rciw_keep_mb.sort_values("Name", inplace=True)

rciw_app_rep = rciw_arr_app.replace(keep)
rciw_app_rep.sort_values("Name", inplace=True)

rciw_full = pd.concat([rciw_app_rep,rciw_keep_mb])

sns.boxplot(data=rciw_full, x="Name", y="RCIW", hue="Version", palette="Greys", ax=ax)
ax.set(xlabel="Endpoint / Microbenchmark")
fig.savefig(out_folder + "/boxplot.pdf")

plt.close(fig)

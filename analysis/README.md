# analysis

## Setup
```
python3 -m venv ./venv
pip install jupyter numpy pandas seaborn matplotlib
```


# main
```
[GET_${}/destinations] performance change: 2.20% [1.01 - 3.42] (2.41%)
[GET_${}/flights/${}/seats] performance change: 2.20% [-4.88 - 11.59] (16.47%)
[GET_${}/flights?from=${}] performance change: 2.33% [1.19 - 3.55] (2.36%)
[POST_${}/bookings] performance change: -1.68% [-10.83 - 8.83] (19.65%)
```

# perf-issue-clean-path
```
[GET_${}/destinations] performance change: 14.25% [13.03 - 15.43] (2.40%)
[GET_${}/flights/${}/seats] performance change: 41.53% [30.96 - 51.08] (20.12%)
[GET_${}/flights?from=${}] performance change: 16.17% [15.01 - 17.33] (2.31%)
[POST_${}/bookings] performance change: 25.90% [15.28 - 38.68] (23.40%)
```

# perf-issue-request-id
```
[GET_${}/destinations] performance change: 1546.71% [1522.69 - 1569.22] (46.53%)
[GET_${}/flights/${}/seats] performance change: 3842.36% [3483.53 - 4272.50] (788.97%)
[GET_${}/flights?from=${}] performance change: 1243.01% [1226.57 - 1261.47] (34.89%)
[POST_${}/bookings] performance change: 5805.06% [4741.58 - 7044.01] (2302.43%)
```

# perf-issue-basic-auth
```
[GET_${}/destinations] performance change: -0.84% [-1.91 - 0.25] (2.16%)
[GET_${}/flights/${}/seats] performance change: -2.53% [-11.76 - 5.51] (17.27%)
[GET_${}/flights?from=${}] performance change: -1.28% [-2.43 - -0.24] (2.19%)
[POST_${}/bookings] performance change: 8.79% [-4.23 - 23.72] (27.94%)
```

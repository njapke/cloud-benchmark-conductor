# analysis

## Setup
```
python3 -m venv ./venv
pip install jupyter numpy pandas seaborn matplotlib
```


# main
```
[GET /destinations] performance change: 2.23% [1.07 - 3.43] (2.36%) (p=0.000000)
[GET /flights/${}/seats] performance change: 2.42% [-5.02 - 12.14] (17.16%) (p=0.120893)
[GET /flights?from=${}] performance change: 2.32% [1.21 - 3.49] (2.28%) (p=0.000000)
[POST /bookings] performance change: -1.44% [-10.45 - 9.23] (19.69%) (p=0.709460)
```

```
[GET /destinations] performance change: -2.27% [-3.43 - -1.13] (2.30%) (p=0.000000)
[GET /flights/${}/seats] performance change: -4.94% [-10.91 - 2.54] (13.45%) (p=0.185206)
[GET /flights?from=${}] performance change: -2.27% [-3.43 - -1.17] (2.26%) (p=0.000000)
[POST /bookings] performance change: -1.72% [-11.80 - 7.29] (19.10%) (p=0.626582)
```

# perf-issue-clean-path
```
[GET /destinations] performance change: 47.38% [46.12 - 48.78] (2.66%) (p=0.000000)
[GET /flights/${}/seats] performance change: 165.12% [148.75 - 182.37] (33.62%) (p=0.000000)
[GET /flights?from=${}] performance change: 50.17% [48.81 - 51.52] (2.71%) (p=0.000000)
[POST /bookings] performance change: 92.67% [81.03 - 108.23] (27.20%) (p=0.000000)
```

# perf-issue-request-id
```
[GET /destinations] performance change: 1561.01% [1537.95 - 1581.84] (43.89%) (p=0.000000)
[GET /flights/${}/seats] performance change: 3803.12% [3431.02 - 4158.49] (727.47%) (p=0.000000)
[GET /flights?from=${}] performance change: 1249.95% [1233.92 - 1266.43] (32.51%) (p=0.000000)
[POST /bookings] performance change: 5802.34% [4866.03 - 6885.75] (2019.72%) (p=0.000000)
```

# perf-issue-basic-auth
```
[GET /destinations] performance change: 16.96% [15.64 - 18.38] (2.74%) (p=0.000000)
[GET /flights/${}/seats] performance change: 19.35% [10.58 - 29.25] (18.67%) (p=0.000000)
[GET /flights?from=${}] performance change: 16.09% [14.83 - 17.36] (2.53%) (p=0.000000)
[POST /bookings] performance change: 325.88% [296.65 - 360.07] (63.41%) (p=0.000000)
```





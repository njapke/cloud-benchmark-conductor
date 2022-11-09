# analysis

## Setup
```
python3 -m venv ./venv
pip install jupyter numpy pandas seaborn matplotlib
```


# main
```
[searchAndBookFlight - GET /destinations] performance change: 0.13% [-5.80 - 7.36] (13.17%)
[searchAndBookFlight - GET /flights/${}/seats] performance change: 2.20% [-4.73 - 11.80] (16.54%)
[searchAndBookFlight - GET /flights?from=${}] performance change: 6.39% [-0.20 - 13.23] (13.44%)
[searchAndBookFlight - POST /bookings] performance change: -1.68% [-10.88 - 9.27] (20.14%)
[searchFlights - GET /destinations] performance change: 2.16% [1.04 - 3.43] (2.39%)
[searchFlights - GET /flights?from=${}] performance change: 2.16% [1.04 - 3.39] (2.35%)
```

# perf-issue-clean-path
```
[searchAndBookFlight - GET /destinations] performance change: 48.80% [42.35 - 55.51] (13.16%)
[searchAndBookFlight - GET /flights/${}/seats] performance change: 165.32% [148.90 - 182.02] (33.12%)
[searchAndBookFlight - GET /flights?from=${}] performance change: 48.81% [42.30 - 56.67] (14.36%)
[searchAndBookFlight - POST /bookings] performance change: 90.81% [80.30 - 107.32] (27.02%)
[searchFlights - GET /destinations] performance change: 47.40% [46.03 - 48.76] (2.73%)
[searchFlights - GET /flights?from=${}] performance change: 49.85% [48.52 - 51.20] (2.68%)
```

# perf-issue-request-id
```
[searchAndBookFlight - GET /destinations] performance change: 1471.81% [1371.52 - 1599.09] (227.57%)
[searchAndBookFlight - GET /flights/${}/seats] performance change: 3842.36% [3475.19 - 4248.49] (773.30%)
[searchAndBookFlight - GET /flights?from=${}] performance change: 1143.18% [1063.71 - 1230.51] (166.80%)
[searchAndBookFlight - POST /bookings] performance change: 5805.06% [4709.97 - 7056.43] (2346.46%)
[searchFlights - GET /destinations] performance change: 1549.83% [1524.49 - 1573.43] (48.94%)
[searchFlights - GET /flights?from=${}] performance change: 1247.78% [1231.42 - 1265.78] (34.37%)
```

# perf-issue-basic-auth
```
[searchAndBookFlight - GET /destinations] performance change: 21.56% [14.28 - 27.51] (13.23%)
[searchAndBookFlight - GET /flights/${}/seats] performance change: 20.03% [11.03 - 30.79] (19.76%)
[searchAndBookFlight - GET /flights?from=${}] performance change: 10.88% [4.04 - 16.46] (12.42%)
[searchAndBookFlight - POST /bookings] performance change: 333.17% [299.98 - 368.41] (68.43%)
[searchFlights - GET /destinations] performance change: 16.50% [15.13 - 18.04] (2.91%)
[searchFlights - GET /flights?from=${}] performance change: 16.33% [14.99 - 17.63] (2.65%)
```

# analysis

## Setup
```
python3 -m venv ./venv
pip install -r requirements.txt
```


# main
```
[GET /destinations] performance change: -1.51% [-2.67 - -0.40] (2.27%) (p=0.000000)
[GET /flights/${id}/seats] performance change: -2.20% [-8.61 - 4.37] (12.98%) (p=0.725289)
[GET /flights?from=${airport}] performance change: -2.84% [-3.86 - -1.77] (2.10%) (p=0.000000)
[POST /bookings] performance change: -3.32% [-12.02 - 6.32] (18.34%) (p=0.922314)
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


# Full Microbenchmark Suite
```
database.BenchmarkGet
database.BenchmarkGetGenerics
database.BenchmarkPut
database.BenchmarkRawGet
database.BenchmarkRawValues
database.BenchmarkValues
database.BenchmarkValuesGenerics

service.BenchmarkHandlerCreateBooking
service.BenchmarkHandlerGetBookings
service.BenchmarkHandlerGetDestinations
service.BenchmarkHandlerGetFlight
service.BenchmarkHandlerGetFlightSeats
service.BenchmarkHandlerGetFlights
service.BenchmarkHandlerGetFlightsQuery

service.BenchmarkRequestBookings
service.BenchmarkRequestCreateBooking
service.BenchmarkRequestDestinations
service.BenchmarkRequestFlight
service.BenchmarkRequestFlights
service.BenchmarkRequestFlightsQuery
service.BenchmarkRequestSeats
```

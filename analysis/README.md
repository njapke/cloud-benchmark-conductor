# analysis

## Setup
```
python3 -m venv ./venv
pip install -r requirements.txt
```

## Application Benchmarks
The analysis of the application benchmark is done in the [`bootstrapping-app.ipynb`](./bootstrapping-app.ipynb) notebook.

## Microbenchmark Suites
The analysis of the microbenchmark suites is done in the [`bootstrapping-mb.ipynb`](./bootstrapping-mb.ipynb) notebook.

### Full Microbenchmark Suite
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

### Optimized Microbenchmark Suite
```
service.BenchmarkRequestFlights
service.BenchmarkHandlerGetFlightSeats
service.BenchmarkRequestDestinations
service.BenchmarkRequestCreateBooking
```

package benchmark

type ArtilleryResult struct {
	Aggregate    ArtilleryMetrics   `json:"aggregate"`
	Intermediate []ArtilleryMetrics `json:"intermediate"`
}

type ArtilleryMetricsHistogram struct {
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Count  float64 `json:"count"`
	P50    float64 `json:"p50"`
	Median float64 `json:"median"`
	P75    float64 `json:"p75"`
	P90    float64 `json:"p90"`
	P95    float64 `json:"p95"`
	P99    float64 `json:"p99"`
	P999   float64 `json:"p999"`
}

type ArtilleryMetrics struct {
	Counters         map[string]int64                     `json:"counters"`
	Rates            map[string]int64                     `json:"rates"`
	FirstCounterAt   int64                                `json:"firstCounterAt"`
	FirstHistogramAt int64                                `json:"firstHistogramAt"`
	LastCounterAt    int64                                `json:"lastCounterAt"`
	LastHistogramAt  int64                                `json:"lastHistogramAt"`
	FirstMetricAt    int64                                `json:"firstMetricAt"`
	LastMetricAt     int64                                `json:"lastMetricAt"`
	Period           interface{}                          `json:"period"`
	Summaries        map[string]ArtilleryMetricsHistogram `json:"summaries"`
	Histograms       map[string]ArtilleryMetricsHistogram `json:"histograms"`
}

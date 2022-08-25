package benchmark

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type ArtilleryResult struct {
	Aggregate    ArtilleryMetrics   `json:"aggregate"`
	Intermediate []ArtilleryMetrics `json:"intermediate"`
}

func ReadArtilleryResult(outputFileName string) (ArtilleryResult, error) {
	outputFile, err := os.Open(outputFileName)
	if err != nil {
		return ArtilleryResult{}, err
	}
	defer outputFile.Close()
	var artilleryResult ArtilleryResult
	err = json.NewDecoder(outputFile).Decode(&artilleryResult)
	return artilleryResult, err
}

var csvHeader = []string{"version", "index", "period", "width", "scenario", "method", "path", "request_time_median", "request_count"}

func ReadArtilleryResultToCSV(outputFileNames map[string]string) (io.Reader, error) {
	buf := &bytes.Buffer{}
	csvWriter := csv.NewWriter(buf)
	err := csvWriter.Write(csvHeader)
	if err != nil {
		return nil, err
	}
	for outputFileVersion, outputFileName := range outputFileNames {
		artilleryResult, readErr := ReadArtilleryResult(outputFileName)
		if readErr != nil {
			return nil, readErr
		}
		_ = csvWriter.WriteAll(artilleryResult.IntermediateRecords(outputFileVersion))
	}
	csvWriter.Flush()
	err = csvWriter.Error()
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (r ArtilleryResult) IntermediateRecords(version string) [][]string {
	res := make([][]string, 0)
	for i, metrics := range r.Intermediate {
		prefix := []string{
			version,
			fmt.Sprintf("%d", i),
			fmt.Sprintf("%v", metrics.Period),
			fmt.Sprintf("%d", metrics.LastMetricAt-metrics.FirstMetricAt),
		}
		res = append(res, metrics.Histograms.Records(prefix)...)
	}
	return res
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

type ArtilleryMetricsHistograms map[string]ArtilleryMetricsHistogram

func (h ArtilleryMetricsHistograms) Records(prefixRows []string) [][]string {
	res := make([][]string, 0)
	for key, val := range h {
		// example key: scenario.searchAndBookFlight.GET./flights/$flightID/seats.total
		if !strings.HasPrefix(key, "scenario") {
			continue
		}
		fields := strings.Split(key, ".")
		if len(fields) < 5 || fields[4] != "total" {
			continue
		}
		row := append([]string{}, prefixRows...)
		row = append(row,
			fields[1], // scenario name
			fields[2], // method
			fields[3], // path
			strconv.FormatFloat(val.Median, 'f', -1, 64),
			strconv.FormatFloat(val.Count, 'f', -1, 64),
		)
		res = append(res, row)
	}
	return res
}

type ArtilleryMetrics struct {
	Counters         map[string]int64           `json:"counters"`
	Rates            map[string]int64           `json:"rates"`
	FirstCounterAt   int64                      `json:"firstCounterAt"`
	FirstHistogramAt int64                      `json:"firstHistogramAt"`
	LastCounterAt    int64                      `json:"lastCounterAt"`
	LastHistogramAt  int64                      `json:"lastHistogramAt"`
	FirstMetricAt    int64                      `json:"firstMetricAt"`
	LastMetricAt     int64                      `json:"lastMetricAt"`
	Period           interface{}                `json:"period"`
	Summaries        ArtilleryMetricsHistograms `json:"summaries"`
	Histograms       ArtilleryMetricsHistograms `json:"histograms"`
}

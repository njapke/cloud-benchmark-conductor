package min

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func WriteAll(projects []string, res []*Result, out io.Writer, projectOnly, writeHeader bool) error {
	csv := csv.NewWriter(out)
	csv.Comma = ';'

	// write header
	if writeHeader {
		err := csv.Write([]string{
			"project",
			"system",
			"config_node_count",
			"config_node_fraction",
			"project_only",
			"total",
			"rank",
			"name",
			"additional_nodes",
			"appBenchTime(ms)",
		})
		if err != nil {
			return fmt.Errorf("could not write header: %w", err)
		}
		csv.Flush()
	}

	for _, r := range res {
		Write(projects, r, csv, projectOnly)
	}

	csv.Flush()

	return nil
}

func Write(projects []string, r *Result, csv *csv.Writer, projectOnly bool) {
	sels := r.Selected
	nrSels := len(sels)

	for i, selected := range sels {
		csv.Write([]string{
			strings.Join(projects, ","),
			r.CG.System,
			strconv.Itoa(r.CG.Config.NodeCount),
			fmt.Sprintf("%.5f", r.CG.Config.NodeFraction),
			strconv.FormatBool(projectOnly),
			strconv.Itoa(nrSels),
			strconv.Itoa(i + 1),
			selected.Benchmark,
			strconv.Itoa(selected.AdditionalNodes),
			strconv.FormatInt(selected.appBenchTime.Milliseconds(), 10),
		})
	}
}

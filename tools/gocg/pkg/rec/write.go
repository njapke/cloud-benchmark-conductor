package rec

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func WriteAll(projects []string, res []*Result, nrBenchs int, out io.Writer) error {
	csv := csv.NewWriter(out)
	csv.Comma = ';'

	// write header
	err := csv.Write([]string{
		"project",
		"system",
		"config_node_count",
		"config_node_fraction",
		"requested_recs",
		"actual_recs",
		"func_name",
		"func_time",
		"func_total_time",
		"additional_nodes",
	})
	if err != nil {
		return fmt.Errorf("could not write header: %w", err)
	}
	csv.Flush()

	for _, r := range res {
		Write(projects, r, nrBenchs, csv)
	}

	csv.Flush()

	return nil
}

func Write(projects []string, r *Result, nrBenchs int, csv *csv.Writer) {
	for _, rf := range r.RecommendedFunctions {
		f := rf.Function
		csv.Write([]string{
			strings.Join(projects, ","),
			r.CG.System,
			strconv.Itoa(r.CG.Config.NodeCount),
			fmt.Sprintf("%.5f", r.CG.Config.NodeFraction),
			strconv.Itoa(nrBenchs),
			strconv.Itoa(len(r.RecommendedFunctions)),
			f.Name,
			strconv.FormatInt(f.FunctionTime.Nanoseconds(), 10),
			strconv.FormatInt(f.TotalTime.Nanoseconds(), 10),
			strconv.Itoa(rf.AdditionalNodes),
		})
	}
}

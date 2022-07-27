package main

import (
	"encoding/csv"
	"encoding/json"
	"os"

	"github.com/christophwitzko/master-thesis/pkg/benchmark"
	"github.com/christophwitzko/master-thesis/pkg/cli"
	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/spf13/cobra"
)

var log = logger.Default()

var rootCmd = &cobra.Command{
	Use:   "microbenchmark-runner",
	Short: "microbenchmark runner tool",
	Long:  "This tool is used to run microbenchmarks using RMIT (Randomized Multiple Interleaved Trials).",
	Run:   cli.WrapRunE(rootRun),
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func main() {
	rootCmd.Flags().StringP("input-file", "i", "", "input file")
	_ = rootCmd.MarkFlagRequired("input-file")
	rootCmd.Flags().IntP("run", "r", 1, "run index")
	_ = rootCmd.MarkFlagRequired("run")

	rootCmd.Flags().Bool("csv-header", false, "add csv header")
	rootCmd.Flags().Int("suite-runs", 3, "suite runs")
	rootCmd.Flags().StringP("output", "o", "-", "output file (default stdout)")

	rootCmd.Flags().SortFlags = true

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func readInputFunctions(filename string) ([]benchmark.VersionedFunction, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var functions []benchmark.VersionedFunction
	err = json.NewDecoder(f).Decode(&functions)
	if err != nil {
		return nil, err
	}
	return functions, nil
}

func rootRun(cmd *cobra.Command, args []string) error {
	inputFile, _ := cmd.Flags().GetString("input-file")
	runIndex, _ := cmd.Flags().GetInt("run")
	csvHeader, _ := cmd.Flags().GetBool("csv-header")
	suiteRuns, _ := cmd.Flags().GetInt("suite-runs")
	outputFile, _ := cmd.Flags().GetString("output")

	functions, err := readInputFunctions(inputFile)
	if err != nil {
		return err
	}

	var csvWriter *csv.Writer
	if outputFile == "-" {
		csvWriter = csv.NewWriter(os.Stdout)
	} else {
		log.Printf("writing output to %s", outputFile)
		outFile, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		csvWriter = csv.NewWriter(outFile)
	}

	if csvHeader {
		if err := csvWriter.Write(benchmark.CSVOutputHeader); err != nil {
			return err
		}
		csvWriter.Flush()
	}

	log.Printf("run index: %d\n", runIndex)

	for s := 1; s <= suiteRuns; s++ {
		log.Printf("suite run: %d/%d\n", s, suiteRuns)
		err := benchmark.RunSuite(csvWriter, functions, runIndex, s)
		if err != nil {
			return err
		}
	}

	if err := csvWriter.Error(); err != nil {
		return err
	}
	log.Println("done")
	return nil
}

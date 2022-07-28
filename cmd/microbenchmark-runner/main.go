package main

import (
	"encoding/csv"
	"os"

	"github.com/christophwitzko/master-thesis/pkg/benchmark"
	"github.com/christophwitzko/master-thesis/pkg/cli"
	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/spf13/cobra"
)

func main() {
	log := logger.New()
	rootCmd := &cobra.Command{
		Use:   "microbenchmark-runner",
		Short: "microbenchmark runner tool",
		Long:  "This tool is used to run microbenchmarks using RMIT (Randomized Multiple Interleaved Trials).",
		Run:   cli.WrapRunE(log, rootRun),
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	rootCmd.PersistentFlags().StringP("source-path-v1", "1", "", "source path for version 1")
	_ = rootCmd.MarkPersistentFlagRequired("source-path-v1")
	rootCmd.PersistentFlags().StringP("source-path-v2", "2", "", "source path for version 2")
	_ = rootCmd.MarkPersistentFlagRequired("source-path-v2")

	rootCmd.AddCommand(listCmd(log))

	rootCmd.Flags().IntP("run", "r", 1, "run index")
	rootCmd.Flags().Bool("csv-header", false, "add csv header")
	rootCmd.Flags().Int("suite-runs", 3, "suite runs")
	rootCmd.Flags().StringP("output", "o", "-", "output file (default stdout)")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func rootRun(log *logger.Logger, cmd *cobra.Command, args []string) error {
	sourcePathV1 := cli.MustGetString(cmd, "source-path-v1")
	sourcePathV2 := cli.MustGetString(cmd, "source-path-v2")
	runIndex := cli.MustGetInt(cmd, "run")
	csvHeader := cli.MustGetBool(cmd, "csv-header")
	suiteRuns := cli.MustGetInt(cmd, "suite-runs")
	outputFile := cli.MustGetString(cmd, "output")

	functions, err := benchmark.CombinedFunctionsFromPaths(sourcePathV1, sourcePathV2)
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

	log.Printf("run index: %d", runIndex)

	for s := 1; s <= suiteRuns; s++ {
		log.Printf("suite run: %d/%d", s, suiteRuns)
		err := benchmark.RunSuite(log, csvWriter, functions, runIndex, s)
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

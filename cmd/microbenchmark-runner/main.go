package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	"github.com/christophwitzko/master-thesis/pkg/benchmark"
	"github.com/christophwitzko/master-thesis/pkg/benchmark/setup"
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
	rootCmd.Flags().String("v1", "", "source path or git reference for version 1")
	rootCmd.Flags().String("v2", "", "source path or git reference for version 2")
	rootCmd.Flags().String("git-repository", "", "git repository to use for benchmarking")
	rootCmd.Flags().String("benchmark-directory", ".bench", "directory to use for benchmarking")

	rootCmd.Flags().Bool("list", false, "list all overlapping benchmark functions of the given source paths")
	rootCmd.Flags().Bool("json", false, "output in json format")

	rootCmd.Flags().Int("run", 1, "current run index")
	rootCmd.Flags().Int("suite-runs", 3, "amount of suite runs")

	rootCmd.Flags().Bool("csv-header", false, "add csv header")
	rootCmd.Flags().StringP("output", "o", "-", "output file (default stdout)")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func rootRun(log *logger.Logger, cmd *cobra.Command, args []string) error {
	sourcePathOrRefV1 := cli.MustGetString(cmd, "v1")
	sourcePathOrRefV2 := cli.MustGetString(cmd, "v2")
	runIndex := cli.MustGetInt(cmd, "run")
	suiteRuns := cli.MustGetInt(cmd, "suite-runs")
	csvHeader := cli.MustGetBool(cmd, "csv-header")
	outputFile := cli.MustGetString(cmd, "output")
	gitRepository := cli.MustGetString(cmd, "git-repository")
	benchmarkDirectory := cli.MustGetString(cmd, "benchmark-directory")
	listFunctions := cli.MustGetBool(cmd, "list")
	outputJSON := cli.MustGetBool(cmd, "json")

	if sourcePathOrRefV1 == "" || sourcePathOrRefV2 == "" {
		return fmt.Errorf("source path or git reference for version 1 & 2 are required")
	}

	var sourcePathV1, sourcePathV2 string
	if gitRepository != "" {
		var err error
		sourcePathV1, sourcePathV2, err = setup.SourcePathsFromGitRepository(log, benchmarkDirectory, gitRepository, sourcePathOrRefV1, sourcePathOrRefV2)
		if err != nil {
			return err
		}
	} else {
		sourcePathV1, sourcePathV2 = sourcePathOrRefV1, sourcePathOrRefV2
	}

	versionedFunctions, err := benchmark.CombinedFunctionsFromPaths(sourcePathV1, sourcePathV2)
	if err != nil {
		return err
	}
	// TODO: add versionedFunctions filter

	if listFunctions {
		if outputJSON {
			return json.NewEncoder(os.Stdout).Encode(versionedFunctions)
		}
		for _, fn := range versionedFunctions {
			log.Infof("%s (%s)", fn.V1.Name, fn.V1.PackageName)
			log.Infof("--> %s", fn.V1.FileName)
			log.Infof("--> %s", fn.V2.FileName)
		}
		return nil
	}
	var csvWriter *csv.Writer
	if outputFile == "-" {
		csvWriter = csv.NewWriter(os.Stdout)
	} else {
		log.Infof("writing output to %s", outputFile)
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

	log.Infof("run index: %d", runIndex)

	for s := 1; s <= suiteRuns; s++ {
		log.Infof("suite run: %d/%d", s, suiteRuns)
		err := benchmark.RunSuite(log, csvWriter, versionedFunctions, runIndex, s)
		if err != nil {
			return err
		}
	}

	if err := csvWriter.Error(); err != nil {
		return err
	}
	log.Info("done")
	return nil
}

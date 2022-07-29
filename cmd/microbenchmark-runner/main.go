package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

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

	rootCmd.Flags().Int("run", 1, "current run index")
	rootCmd.Flags().Int("suite-runs", 3, "amount of suite runs")

	rootCmd.Flags().Bool("csv-header", false, "add csv header")
	rootCmd.Flags().StringP("output", "o", "-", "output file (default stdout)")
	rootCmd.Flags().Bool("json", false, "output in json format")
	rootCmd.Flags().Bool("csv", true, "output in csv format")
	rootCmd.MarkFlagsMutuallyExclusive("json", "csv")
	rootCmd.MarkFlagsMutuallyExclusive("json", "csv-header")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func rootRun(log *logger.Logger, cmd *cobra.Command, args []string) error {
	sourcePathOrRefV1 := cli.MustGetString(cmd, "v1")
	sourcePathOrRefV2 := cli.MustGetString(cmd, "v2")
	gitRepository := cli.MustGetString(cmd, "git-repository")
	benchmarkDirectory := cli.MustGetString(cmd, "benchmark-directory")

	listFunctions := cli.MustGetBool(cmd, "list")

	runIndex := cli.MustGetInt(cmd, "run")
	suiteRuns := cli.MustGetInt(cmd, "suite-runs")

	csvHeader := cli.MustGetBool(cmd, "csv-header")
	outputFile := cli.MustGetString(cmd, "output")
	outputJSON := cli.MustGetBool(cmd, "json")
	outputCSV := cli.MustGetBool(cmd, "csv")

	// if --csv is not set and --json is set, output format should be json
	if cmd.Flags().Lookup("json").Changed && !cmd.Flags().Lookup("csv").Changed {
		outputCSV = false
	}

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
		for _, fn := range versionedFunctions {
			log.Infof("%s", fn.V1.String())
			log.Infof("--> %s", filepath.Join(fn.V1.RootDirectory, fn.V1.FileName))
			log.Infof("--> %s", filepath.Join(fn.V2.RootDirectory, fn.V2.FileName))
		}
		return nil
	}

	var outputWriter io.Writer
	if outputFile == "-" {
		outputWriter = os.Stdout
	} else {
		log.Infof("writing output to %s", outputFile)
		outFile, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer func() {
			outFile.Sync()
			outFile.Close()
		}()
		outputWriter = outFile
	}

	var resultWriter benchmark.ResultWriter
	if outputCSV {
		resultWriter = benchmark.NewCSVResultWriter(outputWriter)
	} else if outputJSON {
		resultWriter = benchmark.NewJSONResultWriter(outputWriter)
	} else {
		return fmt.Errorf("no output format specified")
	}

	if csvWriter, ok := resultWriter.(*benchmark.CSVResultWriter); ok && csvHeader {
		if err := csvWriter.WriteRaw(benchmark.CSVOutputHeader); err != nil {
			return err
		}
	}

	log.Infof("run index: %d", runIndex)

	for s := 1; s <= suiteRuns; s++ {
		log.Infof("suite run: %d/%d", s, suiteRuns)
		err := benchmark.RunSuite(log, resultWriter, versionedFunctions, runIndex, s)
		if err != nil {
			return err
		}
	}

	log.Info("done")
	return nil
}

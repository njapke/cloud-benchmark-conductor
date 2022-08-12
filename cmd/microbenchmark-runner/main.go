package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/christophwitzko/master-thesis/pkg/benchmark"
	"github.com/christophwitzko/master-thesis/pkg/benchmark/output"
	"github.com/christophwitzko/master-thesis/pkg/cli"
	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/christophwitzko/master-thesis/pkg/setup"
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
	rootCmd.Flags().String("benchmark-directory", "/tmp/.bench", "directory to use for benchmarking")

	rootCmd.Flags().Int("run", 1, "current run index")
	rootCmd.Flags().Int("suite-runs", 3, "amount of suite runs")

	rootCmd.Flags().StringArrayP("output", "o", []string{"-"}, "output files (default stdout)")
	rootCmd.Flags().Bool("json", false, "output in json format")
	rootCmd.Flags().Bool("csv", true, "output in csv format")
	rootCmd.MarkFlagsMutuallyExclusive("json", "csv")

	rootCmd.Flags().String("include-filter", ".*", "regular expression to filter packages or functions")
	rootCmd.Flags().String("exclude-filter", "^$", "regular expression to exclude packages or functions")
	rootCmd.Flags().StringArray("function", []string{}, "specific functions to benchmark")
	rootCmd.MarkFlagsMutuallyExclusive("function", "include-filter")
	rootCmd.MarkFlagsMutuallyExclusive("function", "exclude-filter")
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func getVersionedFunctions(sourcePathV1, sourcePathV2, includeRegexp, excludeRegexp string, functions []string) (benchmark.VersionedFunctions, error) {
	versionedFunctions, err := benchmark.CombinedFunctionsFromPaths(sourcePathV1, sourcePathV2)
	if err != nil {
		return nil, err
	}

	if len(functions) > 0 {
		return versionedFunctions.Filter(func(vf benchmark.VersionedFunction) bool {
			fnName := vf.String()
			for _, fn := range functions {
				if fn == fnName {
					return true
				}
			}
			return false
		}), nil
	}

	// using include/exclude filters
	includeFilter, err := regexp.Compile(includeRegexp)
	if err != nil {
		return nil, fmt.Errorf("invalid include filter expression %s: %w", includeRegexp, err)
	}
	excludeFilter, err := regexp.Compile(excludeRegexp)
	if err != nil {
		return nil, fmt.Errorf("invalid exclude filter expression %s: %w", excludeRegexp, err)
	}
	return versionedFunctions.Filter(func(vf benchmark.VersionedFunction) bool {
		fnName := vf.String()
		return includeFilter.MatchString(fnName) && !excludeFilter.MatchString(fnName)
	}), nil
}

func rootRun(log *logger.Logger, cmd *cobra.Command, args []string) error {
	sourcePathOrRefV1 := cli.MustGetString(cmd, "v1")
	sourcePathOrRefV2 := cli.MustGetString(cmd, "v2")
	gitRepository := cli.MustGetString(cmd, "git-repository")
	benchmarkDirectory := cli.MustGetString(cmd, "benchmark-directory")
	includeRegexp := cli.MustGetString(cmd, "include-filter")
	excludeRegexp := cli.MustGetString(cmd, "exclude-filter")
	runIndex := cli.MustGetInt(cmd, "run")
	suiteRuns := cli.MustGetInt(cmd, "suite-runs")
	outputPaths := cli.MustGetStringArray(cmd, "output")
	outputFormatJSON := cli.MustGetBool(cmd, "json")
	outputFormatCSV := cli.MustGetBool(cmd, "csv")
	functions := cli.MustGetStringArray(cmd, "function")

	if !outputFormatCSV && !outputFormatJSON {
		return fmt.Errorf("either --json or --csv must be set to true")
	}
	if sourcePathOrRefV1 == "" || sourcePathOrRefV2 == "" {
		return fmt.Errorf("source path or git reference for version 1 & 2 are required")
	}

	log.Info(cli.GetBuildInfo())

	defaultOutputFormat := "csv"
	if outputFormatJSON {
		defaultOutputFormat = "json"
	}
	log.Infof("default output format: %s", defaultOutputFormat)

	sourcePathV1, sourcePathV2, err := setup.SourcePaths(log, benchmarkDirectory, gitRepository, sourcePathOrRefV1, sourcePathOrRefV2)
	if err != nil {
		return err
	}
	versionedFunctions, err := getVersionedFunctions(sourcePathV1, sourcePathV2, includeRegexp, excludeRegexp, functions)
	if err != nil {
		return err
	}

	log.Infof("found %d functions:", len(versionedFunctions))
	for _, fn := range versionedFunctions {
		log.Infof("%s", fn.V1.String())
	}

	// maximum runtime: 30 minutes
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*30)
	defer cancel()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	resultWriter, err := output.New(ctx, outputPaths, defaultOutputFormat)
	if err != nil {
		return fmt.Errorf("failed to open output: %w", err)
	}
	defer resultWriter.Close()

	log.Infof("run index: %d", runIndex)
	for s := 1; s <= suiteRuns; s++ {
		log.Infof("suite run: %d/%d", s, suiteRuns)
		err := benchmark.RunSuite(ctx, log, resultWriter, versionedFunctions, runIndex, s)
		if err != nil {
			return err
		}
	}

	log.Info("done.")
	return nil
}

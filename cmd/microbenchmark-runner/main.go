package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"syscall"
	"time"

	"github.com/christophwitzko/master-thesis/pkg/benchmark"
	"github.com/christophwitzko/master-thesis/pkg/benchmark/output"
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
	rootCmd.Flags().String("benchmark-directory", "/tmp/.bench", "directory to use for benchmarking")

	rootCmd.Flags().Bool("list", false, "list all overlapping benchmark functions of the given source paths")

	rootCmd.Flags().Int("run", 1, "current run index")
	rootCmd.Flags().Int("suite-runs", 3, "amount of suite runs")

	rootCmd.Flags().StringArrayP("output", "o", []string{"-"}, "output files (default stdout)")
	rootCmd.Flags().Bool("json", false, "output in json format")
	rootCmd.Flags().Bool("csv", true, "output in csv format")
	rootCmd.MarkFlagsMutuallyExclusive("json", "csv")

	rootCmd.Flags().String("include-filter", ".*", "regular expression to filter packages or functions")
	rootCmd.Flags().String("exclude-filter", "^$", "regular expression to exclude packages or functions")

	rootCmd.Flags().BoolP("version", "v", false, "print build information")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func rootRun(log *logger.Logger, cmd *cobra.Command, args []string) error {
	if cli.MustGetBool(cmd, "version") {
		fmt.Println(cli.GetBuildInfo())
		return nil
	}
	sourcePathOrRefV1 := cli.MustGetString(cmd, "v1")
	sourcePathOrRefV2 := cli.MustGetString(cmd, "v2")
	gitRepository := cli.MustGetString(cmd, "git-repository")
	benchmarkDirectory := cli.MustGetString(cmd, "benchmark-directory")

	listFunctions := cli.MustGetBool(cmd, "list")

	runIndex := cli.MustGetInt(cmd, "run")
	suiteRuns := cli.MustGetInt(cmd, "suite-runs")

	outputPaths := cli.MustGetStringArray(cmd, "output")

	outputFormatJSON := cli.MustGetBool(cmd, "json")
	outputFormatCSV := cli.MustGetBool(cmd, "csv")

	if !outputFormatCSV && !outputFormatJSON {
		return fmt.Errorf("either --json or --csv must be set to true")
	}

	defaultOutputFormat := "csv"
	if outputFormatJSON {
		defaultOutputFormat = "json"
	}

	log.Info(cli.GetBuildInfo())
	log.Infof("default output format: %s", defaultOutputFormat)

	includeRegexp := cli.MustGetString(cmd, "include-filter")
	excludeRegexp := cli.MustGetString(cmd, "exclude-filter")
	includeFilter, err := regexp.Compile(includeRegexp)
	if err != nil {
		return fmt.Errorf("invalid include filter expression %s: %w", includeRegexp, err)
	}
	excludeFilter, err := regexp.Compile(excludeRegexp)
	if err != nil {
		return fmt.Errorf("invalid exclude filter expression %s: %w", excludeRegexp, err)
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

	versionedFunctions = versionedFunctions.Filter(func(vf benchmark.VersionedFunction) bool {
		fnName := vf.String()
		return includeFilter.MatchString(fnName) && !excludeFilter.MatchString(fnName)
	})

	if listFunctions {
		for _, fn := range versionedFunctions {
			log.Infof("%s", fn.V1.String())
			log.Infof("--> %s", filepath.Join(fn.V1.RootDirectory, fn.V1.FileName))
			log.Infof("--> %s", filepath.Join(fn.V2.RootDirectory, fn.V2.FileName))
		}
		return nil
	}

	if len(outputPaths) != 1 || outputPaths[0] != "-" {
		for _, outputPath := range outputPaths {
			log.Infof("writing output to %s", outputPath)
		}
	}

	// maximum runtime: 30 minutes
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*30)
	defer cancel()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var resultWriter benchmark.ResultWriter
	resultWriter, err = output.New(ctx, outputPaths, defaultOutputFormat)
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

	log.Info("done")
	return nil
}

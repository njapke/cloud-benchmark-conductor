package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/christophwitzko/masters-thesis/pkg/cli"
	"github.com/christophwitzko/masters-thesis/pkg/gcloud/storage"
	"github.com/christophwitzko/masters-thesis/pkg/logger"
	"github.com/christophwitzko/masters-thesis/pkg/microbenchmark"
	"github.com/christophwitzko/masters-thesis/pkg/microbenchmark/output"
	"github.com/christophwitzko/masters-thesis/pkg/profile"
	"github.com/christophwitzko/masters-thesis/pkg/retry"
	"github.com/christophwitzko/masters-thesis/pkg/setup"
	"github.com/spf13/cobra"
)

func main() {
	log := logger.New()
	rootCmd := &cobra.Command{
		Use:   "microbenchmark-runner",
		Short: "microbenchmark runner tool",
		Long:  "This tool is used to run microbenchmarks using RMIT (Randomized Multiple Interleaved Trials).",
		Args:  cobra.NoArgs,
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

	rootCmd.Flags().Bool("profiling", false, "create a profile for each function")
	rootCmd.Flags().String("profiling-local-output", "./profiles", "output directory for profiling")
	rootCmd.Flags().String("profiling-gcs-output", "", "if set uploads the profiles to google cloud storage")
	rootCmd.Flags().Duration("timeout", 60*time.Minute, "timeout for the benchmark execution")
	rootCmd.Flags().StringArray("env", []string{}, "additional environment variables to set")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func getVersionedFunctions(sourcePathV1, sourcePathV2, includeRegexp, excludeRegexp string, functions []string) (microbenchmark.VersionedFunctions, error) {
	versionedFunctions, err := microbenchmark.CombinedFunctionsFromPaths(sourcePathV1, sourcePathV2)
	if err != nil {
		return nil, err
	}

	if len(functions) > 0 {
		return versionedFunctions.Filter(func(vf microbenchmark.VersionedFunction) bool {
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
	return versionedFunctions.Filter(func(vf microbenchmark.VersionedFunction) bool {
		fnName := vf.String()
		return includeFilter.MatchString(fnName) && !excludeFilter.MatchString(fnName)
	}), nil
}

func runMicrobenchmarks(ctx context.Context, log *logger.Logger, versionedFunctions microbenchmark.VersionedFunctions, outputPaths []string, defaultOutputFormat string, suiteRuns, runIndex int, env []string) error {
	resultWriter, err := output.New(ctx, outputPaths, defaultOutputFormat)
	if err != nil {
		return fmt.Errorf("failed to open output: %w", err)
	}
	defer resultWriter.Close()

	log.Infof("run index: %d", runIndex)
	for s := 1; s <= suiteRuns; s++ {
		log.Infof("suite run: %d/%d", s, suiteRuns)
		err := microbenchmark.RunSuite(ctx, log, resultWriter, versionedFunctions, runIndex, s, env)
		if err != nil {
			return err
		}
	}
	log.Info("done.")
	return nil
}

func runProfiling(ctx context.Context, log *logger.Logger, versionedFunctions microbenchmark.VersionedFunctions, profilingLocalOutput, profilingGCSOutput string, env []string) error {
	log.Warn("profiling only functions from version 1")
	err := setup.CreateDirectory(profilingLocalOutput)
	if err != nil {
		return fmt.Errorf("failed to create local profile output directory: %w", err)
	}
	var profilingGCSOutputHost, profilingGCSOutputPath string
	if profilingGCSOutput != "" {
		profilingGCSOutputHost, profilingGCSOutputPath, err = storage.ParseURL(profilingGCSOutput)
		if err != nil {
			return fmt.Errorf("failed to parse gcs output url: %w", err)
		}
	}
	logPrefix := "|pprof|"
	for _, vf := range versionedFunctions {
		profileFile, err := microbenchmark.RunProfile(ctx, log, vf.V1, profilingLocalOutput, env)
		profileFileName := filepath.Base(profileFile)
		if profilingGCSOutput != "" {
			log.Infof("%s uploading pprof profile to gcs: %s", logPrefix, profileFileName)
			err = retry.OnError(ctx, log, logPrefix, func() error {
				return storage.UploadFileToBucket(ctx, profilingGCSOutputHost, filepath.Join(profilingGCSOutputPath, profileFileName), profileFile)
			})
			if err != nil {
				return err
			}
		}
		if err != nil {
			return err
		}
		callGraphFile := profileFile + ".dot"
		callGraphFileName := filepath.Base(callGraphFile)
		err = profile.ToCallGraph(log, logPrefix, profileFile, callGraphFile)
		if err != nil {
			return err
		}
		if profilingGCSOutput != "" {
			log.Infof("%s uploading call graph to gcs: %s", logPrefix, callGraphFileName)
			err = retry.OnError(ctx, log, logPrefix, func() error {
				return storage.UploadFileToBucket(ctx, profilingGCSOutputHost, filepath.Join(profilingGCSOutputPath, callGraphFileName), callGraphFile)
			})
			if err != nil {
				return err
			}
		}
	}
	log.Info("done.")
	return nil
}

func rootRun(log *logger.Logger, cmd *cobra.Command, _ []string) error {
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
	shouldRunProfiling := cli.MustGetBool(cmd, "profiling")
	profilingLocalOutput := cli.MustGetString(cmd, "profiling-local-output")
	profilingGCSOutput := cli.MustGetString(cmd, "profiling-gcs-output")
	timeout := cli.MustGetDuration(cmd, "timeout")
	envVars := cli.MustGetStringArray(cmd, "env")

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

	log.Infof("timeout: %s", timeout)
	ctx, cancel := cli.NewContext(timeout)
	defer cancel()

	if shouldRunProfiling {
		return runProfiling(ctx, log, versionedFunctions, filepath.Join(sourcePathV1, profilingLocalOutput), profilingGCSOutput, envVars)
	}

	return runMicrobenchmarks(ctx, log, versionedFunctions, outputPaths, defaultOutputFormat, suiteRuns, runIndex, envVars)
}

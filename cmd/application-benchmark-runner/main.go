package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/christophwitzko/master-thesis/pkg/application/benchmark"
	"github.com/christophwitzko/master-thesis/pkg/cli"
	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/christophwitzko/master-thesis/pkg/netutil"
	"github.com/christophwitzko/master-thesis/pkg/profile"
	"github.com/christophwitzko/master-thesis/pkg/retry"
	"github.com/christophwitzko/master-thesis/pkg/setup"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var (
	defaultArtilleryConfig = "./artillery/config.yaml"
	defaultK6Config        = "./k6/script.js"
)

func main() {
	log := logger.New()
	rootCmd := &cobra.Command{
		Use:   "application-benchmark-runner",
		Short: "application benchmark runner tool",
		Long:  "This tool is used to run the application benchmarks using artillery or k6.",
		Args:  cobra.NoArgs,
		Run:   cli.WrapRunE(log, rootRun),
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	rootCmd.Flags().String("reference", "", "git reference or source path of the desired application benchmark config")
	rootCmd.Flags().String("git-repository", "", "git repository to use for installing the applications")
	rootCmd.Flags().String("benchmark-directory", "/tmp/.appbench", "directory to use for running the application benchmarks")
	rootCmd.Flags().String("config", defaultArtilleryConfig, "location of the application benchmark config or script relative to the repository root or provided source path")
	rootCmd.Flags().StringArray("target", []string{"v1=127.0.0.1:3000"}, "target to run the application benchmark on")
	rootCmd.Flags().String("results-output", "", "path where the results should be stored [e.g. gs://ab-results/app]")
	rootCmd.Flags().Duration("timeout", 60*time.Minute, "timeout for the benchmark execution")
	rootCmd.Flags().Bool("profiling", false, "enable continuous profiling")
	rootCmd.Flags().String("profiling-endpoint", "/debug/pprof/profile", "pprof endpoint to use for profiling")
	rootCmd.Flags().Duration("profiling-interval", 5*time.Minute, "profiling interval")
	rootCmd.Flags().Duration("profiling-duration", 30*time.Second, "profiling duration")
	rootCmd.Flags().String("tool", "artillery", "tool to run the benchmarks [artillery or k6]")
	rootCmd.Flags().StringArray("env", []string{}, "additional environment variables to set")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func parseInputTargets(config *benchmark.Config, inputTargets []string) []*benchmark.TargetInfo {
	fileExtenstion := "json"
	if config.Tool == "k6" {
		fileExtenstion = "csv.gz"
	}
	targets := make([]*benchmark.TargetInfo, 0, len(inputTargets))
	for _, target := range inputTargets {
		targetName := target
		targetEndpoint := target
		if strings.Contains(target, "=") {
			targetName, targetEndpoint, _ = strings.Cut(target, "=")
		}
		newTarget := &benchmark.TargetInfo{
			Name:     targetName,
			Endpoint: targetEndpoint,
			OutputFile: filepath.Join(
				config.ConfigDir,
				fmt.Sprintf("%s.%s", targetName, fileExtenstion),
			),
		}
		targets = append(targets, newTarget)
	}
	return targets
}

type profilingConfig struct {
	Endpoint  string
	Interval  time.Duration
	Duration  time.Duration
	OutputDir string
}

func (c profilingConfig) EndpointFromTarget(target string) string {
	return fmt.Sprintf("http://%s%s?seconds=%.0f", target, c.Endpoint, c.Duration.Seconds())
}

func (c profilingConfig) ProfileFileName(targetName string, index int) string {
	return filepath.Join(c.OutputDir, fmt.Sprintf("pprof-%s-%d.out", targetName, index))
}

func runProfiler(ctx context.Context, log *logger.Logger, logPrefix string, benchConf *benchmark.Config, profileEndpoint, profileFile string) error {
	err := profile.Fetch(ctx, profileEndpoint, profileFile)
	if err != nil {
		return err
	}
	profileFileName := filepath.Base(profileFile)
	log.Infof("%s profiling finished. uploading: %s", logPrefix, profileFileName)
	err = retry.OnError(ctx, log, logPrefix, func() error {
		return benchConf.UploadToBucketFromFile(ctx, profileFileName, profileFile)
	})
	if err != nil {
		return err
	}
	log.Infof("%s converting to call graph...", logPrefix)
	callGraphFile := profileFile + ".dot"
	callGraphFileName := filepath.Base(callGraphFile)
	err = profile.ToCallGraph(log, logPrefix, profileFile, callGraphFile)
	if err != nil {
		return err
	}
	log.Infof("%s uploading: %s", logPrefix, callGraphFileName)
	return retry.OnError(ctx, log, logPrefix, func() error {
		return benchConf.UploadToBucketFromFile(ctx, callGraphFileName, callGraphFile)
	})
}

func runContinuousProfiler(ctx context.Context, log *logger.Logger, benchConf *benchmark.Config, profilingConf profilingConfig, targetInfo *benchmark.TargetInfo) error {
	profileEndpoint := profilingConf.EndpointFromTarget(targetInfo.Endpoint)
	logPrefix := fmt.Sprintf("[pprof/%s]", targetInfo.Name)
	log.Infof("%s starting continuous profiling on %s", logPrefix, profileEndpoint)
	ticker := time.NewTicker(profilingConf.Interval)
	defer ticker.Stop()
	i := 0
	for {
		select {
		case <-ctx.Done():
			log.Infof("%s stopping...", logPrefix)
			return ctx.Err()
		case <-ticker.C:
			profileFile := profilingConf.ProfileFileName(targetInfo.Name, i)
			profileFileName := filepath.Base(profileFile)
			i++
			log.Infof("%s profiling to %s", logPrefix, profileFileName)
			err := runProfiler(ctx, log, logPrefix, benchConf, profileEndpoint, profileFile)
			if err != nil {
				log.Warningf("%s error while profiling: %v", logPrefix, err)
				continue
			}
		}
	}
}

func waitForTargets(ctx context.Context, log *logger.Logger, targets []*benchmark.TargetInfo) error {
	errGroup, groupCtx := errgroup.WithContext(ctx)
	for _, targetInfo := range targets {
		targetInfo := targetInfo
		errGroup.Go(func() error {
			log.Infof("waiting for target %s (%s)", targetInfo.Name, targetInfo.Endpoint)
			return netutil.WaitForPortOpen(groupCtx, targetInfo.Endpoint)
		})
	}
	return errGroup.Wait()
}

//gocyclo:ignore
func rootRun(log *logger.Logger, cmd *cobra.Command, args []string) error {
	referenceOrPath := cli.MustGetString(cmd, "reference")
	gitRepository := cli.MustGetString(cmd, "git-repository")
	benchmarkDirectory := cli.MustGetString(cmd, "benchmark-directory")
	relConfigFile := cli.MustGetString(cmd, "config")
	inputTargets := cli.MustGetStringArray(cmd, "target")
	resultsOutputPath := cli.MustGetString(cmd, "results-output")
	timeout := cli.MustGetDuration(cmd, "timeout")
	shouldProfile := cli.MustGetBool(cmd, "profiling")
	profilingConf := profilingConfig{
		Endpoint: cli.MustGetString(cmd, "profiling-endpoint"),
		Interval: cli.MustGetDuration(cmd, "profiling-interval"),
		Duration: cli.MustGetDuration(cmd, "profiling-duration"),
	}
	appBenchTool := strings.ToLower(cli.MustGetString(cmd, "tool"))
	envConfig := cli.MustGetStringArray(cmd, "env")

	if appBenchTool != "artillery" && appBenchTool != "k6" {
		return fmt.Errorf("invalid benchmark tool: %s", appBenchTool)
	}
	log.Infof("application benchmarking tool: %s", appBenchTool)

	// if config is not set but the tool is set, use the default config for the tool
	if appBenchTool == "k6" && !cmd.Flags().Changed("config") {
		relConfigFile = defaultK6Config
	}

	if referenceOrPath == "" {
		return fmt.Errorf("source path or git reference is required")
	}

	log.Info("setting up environment...")
	applicationBenchmarkPath, err := setup.ApplicationBenchmarkPath(log, benchmarkDirectory, gitRepository, referenceOrPath)
	if err != nil {
		return err
	}

	appBenchConfigFile := cli.GetAbsolutePath(filepath.Join(applicationBenchmarkPath, relConfigFile))
	log.Infof("using application benchmark config file: %s", appBenchConfigFile)

	log.Infof("timeout: %s", timeout)
	ctx, cancel := cli.NewContext(timeout)
	defer cancel()

	appBenchConfigDir := filepath.Dir(appBenchConfigFile)
	appBenchConfig := &benchmark.Config{
		Tool:       appBenchTool,
		ConfigDir:  appBenchConfigDir,
		ConfigFile: appBenchConfigFile,
		OutputPath: resultsOutputPath,
		Env:        envConfig,
	}
	err = appBenchConfig.Validate()
	if err != nil {
		return err
	}

	if shouldProfile {
		profilingConf.OutputDir = filepath.Join(appBenchConfigDir, "profile")
		err = setup.CreateDirectory(profilingConf.OutputDir)
		if err != nil {
			return err
		}
	}
	targets := parseInputTargets(appBenchConfig, inputTargets)

	log.Info("waiting for targets to be ready....")
	err = waitForTargets(ctx, log, targets)
	if err != nil {
		return err
	}

	log.Infof("starting %s...", appBenchTool)
	errGroup, groupCtx := errgroup.WithContext(ctx)
	// the profile context is linked to the group error group
	// with their own cancel function
	profileCtx, cancelProfile := context.WithCancel(groupCtx)
	profileGroup, profileCtx := errgroup.WithContext(profileCtx)
	for _, targetInfo := range targets {
		targetInfo := targetInfo
		errGroup.Go(func() error {
			return benchmark.Run(groupCtx, log, appBenchConfig, targetInfo)
		})
		if shouldProfile {
			profileGroup.Go(func() error {
				return runContinuousProfiler(profileCtx, log, appBenchConfig, profilingConf, targetInfo)
			})
		}
	}
	err = errGroup.Wait()
	cancelProfile()
	if err != nil {
		return err
	}
	if shouldProfile {
		log.Infof("waiting for continuous profiling to finish...")
		err = profileGroup.Wait()
		if err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
	}

	if appBenchTool == "artillery" {
		log.Infof("creating combined results csv...")
		resultCSV, err := benchmark.ReadArtilleryResultToCSV(targets)
		if err != nil {
			return err
		}
		log.Infof("uploading combined results...")
		err = appBenchConfig.UploadToBucketFromReader(ctx, "combined-results.csv", resultCSV)
		if err != nil {
			return err
		}
		log.Infof("uploaded to %s", appBenchConfig.GetOutputObjectName("combined-results.csv"))
	}

	log.Infof("done.")
	return nil
}

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/christophwitzko/master-thesis/pkg/application/benchmark"
	"github.com/christophwitzko/master-thesis/pkg/cli"
	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/christophwitzko/master-thesis/pkg/netutil"
	"github.com/christophwitzko/master-thesis/pkg/setup"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func main() {
	log := logger.New()
	rootCmd := &cobra.Command{
		Use:   "application-benchmark-runner",
		Short: "application benchmark runner tool",
		Long:  "This tool is used to run the application benchmarks using artillery.",
		Args:  cobra.NoArgs,
		Run:   cli.WrapRunE(log, rootRun),
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	rootCmd.Flags().String("reference", "", "git reference or source path of the desired application benchmark config")
	rootCmd.Flags().String("git-repository", "", "git repository to use for installing the applications")
	rootCmd.Flags().String("benchmark-directory", "/tmp/.appbench", "directory to use for running the application benchmarks")
	rootCmd.Flags().String("config", "./artillery/config.yaml", "location of the application benchmark config relative to the repository root or provided source path")
	rootCmd.Flags().StringArray("target", []string{"v1=127.0.0.1:3000"}, "target to run the application benchmark on")
	rootCmd.Flags().String("results-output", "", "path where the results should be stored [e.g. gs://ab-results/app]")
	rootCmd.Flags().Duration("timeout", 60*time.Minute, "timeout for the benchmark execution")
	rootCmd.Flags().Bool("profile", false, "enable continuous profiling")
	rootCmd.Flags().String("profile-endpoint", "/debug/pprof/profile", "pprof endpoint to use for profiling")
	rootCmd.Flags().Duration("profile-interval", 5*time.Minute, "profile interval")
	rootCmd.Flags().Duration("profile-duration", 30*time.Second, "profile duration")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func parseInputTargets(configDir string, inputTargets []string) []*benchmark.TargetInfo {
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
				configDir,
				fmt.Sprintf("%s.json", targetName),
			),
		}
		targets = append(targets, newTarget)
	}
	return targets
}

type profileConfig struct {
	Endpoint string
	Interval time.Duration
	Duration time.Duration
}

func (c profileConfig) EndpointFromTarget(target string) string {
	return fmt.Sprintf("http://%s%s?seconds=%.0f", target, c.Endpoint, c.Duration.Seconds())
}

func runContinuousProfiler(ctx context.Context, log *logger.Logger, profileConf profileConfig, targetInfo *benchmark.TargetInfo) error {
	profileEndpoint := profileConf.EndpointFromTarget(targetInfo.Endpoint)
	logPrefix := fmt.Sprintf("[pprof/%s]", targetInfo.Name)
	log.Infof("%s starting continuous profiling on %s", logPrefix, profileEndpoint)
	ticker := time.NewTicker(profileConf.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Infof("%s stopping...", logPrefix)
			return ctx.Err()
		case <-ticker.C:
			log.Infof("%s profiling...", logPrefix)
		}
	}
}

func rootRun(log *logger.Logger, cmd *cobra.Command, args []string) error {
	referenceOrPath := cli.MustGetString(cmd, "reference")
	gitRepository := cli.MustGetString(cmd, "git-repository")
	benchmarkDirectory := cli.MustGetString(cmd, "benchmark-directory")
	relConfigFile := cli.MustGetString(cmd, "config")
	inputTargets := cli.MustGetStringArray(cmd, "target")
	resultsOutputPath := cli.MustGetString(cmd, "results-output")
	timeout := cli.MustGetDuration(cmd, "timeout")
	profile := cli.MustGetBool(cmd, "profile")
	profileConf := profileConfig{
		Endpoint: cli.MustGetString(cmd, "profile-endpoint"),
		Interval: cli.MustGetDuration(cmd, "profile-interval"),
		Duration: cli.MustGetDuration(cmd, "profile-duration"),
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
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	appBenchConfig := &benchmark.Config{
		ConfigFile: appBenchConfigFile,
		OutputPath: resultsOutputPath,
	}
	err = appBenchConfig.Validate()
	if err != nil {
		return err
	}

	appBenchConfigDir := filepath.Dir(appBenchConfigFile)
	targets := parseInputTargets(appBenchConfigDir, inputTargets)

	log.Info("waiting for targets to be ready....")
	errGroup, groupCtx := errgroup.WithContext(ctx)
	for _, targetInfo := range targets {
		targetInfo := targetInfo
		errGroup.Go(func() error {
			log.Infof("waiting for target %s (%s)", targetInfo.Name, targetInfo.Endpoint)
			return netutil.WaitForPortOpen(groupCtx, targetInfo.Endpoint)
		})
	}
	err = errGroup.Wait()
	if err != nil {
		return err
	}

	log.Info("starting artillery...")
	errGroup, groupCtx = errgroup.WithContext(ctx)
	// the profile context is linked to the group error group
	// with their own cancel function
	profileCtx, cancelProfile := context.WithCancel(groupCtx)
	profileGroup, profileCtx := errgroup.WithContext(profileCtx)
	for _, targetInfo := range targets {
		targetInfo := targetInfo
		errGroup.Go(func() error {
			return benchmark.RunArtillery(groupCtx, log, appBenchConfig, targetInfo)
		})
		if profile {
			profileGroup.Go(func() error {
				return runContinuousProfiler(profileCtx, log, profileConf, targetInfo)
			})
		}
	}
	err = errGroup.Wait()
	cancelProfile()
	if err != nil {
		return err
	}
	if profile {
		log.Infof("waiting for continuous profiling to finish...")
		err = profileGroup.Wait()
		if err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
	}

	log.Infof("creating combined results csv...")
	resultCSV, err := benchmark.ReadArtilleryResultToCSV(targets)
	if err != nil {
		return err
	}
	log.Infof("uploading combined results...")
	err = appBenchConfig.UploadToBucket(ctx, "combined-results.csv", resultCSV)
	if err != nil {
		return err
	}
	log.Infof("uploaded to %s", appBenchConfig.GetOutputObjectName("combined-results.csv"))
	log.Infof("done.")
	return nil
}

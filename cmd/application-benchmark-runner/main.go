package main

import (
	"context"
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

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
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

	targets := make(map[string]string)
	outputPaths := make(map[string]string)
	for _, target := range inputTargets {
		targetName := target
		targetEndpoint := target
		if strings.Contains(target, "=") {
			targetName, targetEndpoint, _ = strings.Cut(target, "=")
		}
		targets[targetName] = targetEndpoint
		outputPaths[targetName] = filepath.Join(
			filepath.Dir(appBenchConfig.ConfigFile),
			fmt.Sprintf("%s.json", targetName),
		)
		log.Infof("target: %s (%s)", targetEndpoint, targetName)
	}

	log.Info("waiting for targets to be ready....")
	errGroup, groupCtx := errgroup.WithContext(ctx)
	for _, targetEndpoint := range targets {
		targetEndpoint := targetEndpoint
		errGroup.Go(func() error {
			return netutil.WaitForPortOpen(groupCtx, targetEndpoint)
		})
	}
	err = errGroup.Wait()
	if err != nil {
		return err
	}

	log.Info("starting artillery...")
	errGroup, groupCtx = errgroup.WithContext(ctx)
	for targetName, targetEndpoint := range targets {
		targetName := targetName
		targetEndpoint := targetEndpoint
		errGroup.Go(func() error {
			return benchmark.RunArtillery(groupCtx, log, appBenchConfig, targetName, targetEndpoint)
		})
	}
	err = errGroup.Wait()
	if err != nil {
		return err
	}

	log.Infof("creating combined results csv...")
	resultCSV, err := benchmark.ReadArtilleryResultToCSV(outputPaths)
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

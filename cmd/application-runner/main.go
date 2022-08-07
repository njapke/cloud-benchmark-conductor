package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/christophwitzko/master-thesis/pkg/application"
	"github.com/christophwitzko/master-thesis/pkg/cli"
	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/christophwitzko/master-thesis/pkg/setup"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
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
	rootCmd.Flags().String("git-repository", "", "git repository to use for installing the applications")
	rootCmd.Flags().String("application-directory", "/tmp/.application", "directory to use for running the application")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func rootRun(log *logger.Logger, cmd *cobra.Command, args []string) error {
	sourcePathOrRefV1 := cli.MustGetString(cmd, "v1")
	sourcePathOrRefV2 := cli.MustGetString(cmd, "v2")
	gitRepository := cli.MustGetString(cmd, "git-repository")
	applicationDirectory := cli.MustGetString(cmd, "application-directory")

	sourcePathV1, sourcePathV2, err := setup.SourcePaths(log, applicationDirectory, gitRepository, sourcePathOrRefV1, sourcePathOrRefV2)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*30)
	defer cancel()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	execFileV1 := filepath.Join(sourcePathV1, "v1")
	execFileV2 := filepath.Join(sourcePathV2, "v2")
	buildGroup, buildCtx := errgroup.WithContext(ctx)
	buildGroup.Go(func() error {
		return application.Build(buildCtx, log, sourcePathV1, execFileV1)
	})
	buildGroup.Go(func() error {
		return application.Build(buildCtx, log, sourcePathV2, execFileV2)
	})
	err = buildGroup.Wait()
	if err != nil {
		return err
	}
	log.Info("-> all builds finished successfully")

	runGroup, runCtx := errgroup.WithContext(ctx)
	runGroup.Go(func() error {
		return application.Run(runCtx, log, execFileV1, "127.0.0.1:3030")
	})
	runGroup.Go(func() error {
		return application.Run(runCtx, log, execFileV2, "127.0.0.1:3031")
	})

	return runGroup.Wait()
}

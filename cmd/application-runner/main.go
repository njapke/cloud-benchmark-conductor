package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/christophwitzko/master-thesis/pkg/application"
	"github.com/christophwitzko/master-thesis/pkg/cli"
	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/christophwitzko/master-thesis/pkg/setup"
	"github.com/hashicorp/go-multierror"
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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
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

	execFiles := []string{execFileV1, execFileV2}
	errChan := make(chan error, len(execFiles))
	wg := sync.WaitGroup{}
	wg.Add(len(execFiles))
	startPort := 3000
	for i, execFile := range execFiles {
		go func(i int, execFile string) {
			defer wg.Done()
			appErr := application.Run(ctx, log, execFile, fmt.Sprintf("127.0.0.1:%d", startPort+i))
			if err != nil {
				log.Warnf("-> application %s exited with error: %v", execFile, appErr)
			}
			errChan <- appErr
		}(i, execFile)
	}
	wg.Wait()

	// combine all errors
	err = multierror.Append(<-errChan, <-errChan)
	if errors.Is(err, context.Canceled) {
		log.Warnf("-> applications stopped")
		return nil
	}
	return err
}

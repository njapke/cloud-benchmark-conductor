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
		Use:   "application-runner",
		Short: "application runner tool",
		Long:  "This tool builds and runs two versions of an application concurrently.",
		Args:  cobra.NoArgs,
		Run:   cli.WrapRunE(log, rootRun),
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	rootCmd.Flags().String("v1", "", "source path or git reference for version 1")
	rootCmd.Flags().String("v2", "", "source path or git reference for version 2")
	rootCmd.Flags().String("git-repository", "", "git repository to use for installing the applications")
	rootCmd.Flags().String("application-directory", "/tmp/.application", "directory to use for running the application")
	rootCmd.Flags().String("application-package", "./", "package that should be build and run")
	rootCmd.Flags().String("bind", "127.0.0.1", "bind address")
	rootCmd.Flags().StringArray("env", []string{}, "environment variable to set")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func rootRun(log *logger.Logger, cmd *cobra.Command, args []string) error {
	sourcePathOrRefV1 := cli.MustGetString(cmd, "v1")
	sourcePathOrRefV2 := cli.MustGetString(cmd, "v2")
	gitRepository := cli.MustGetString(cmd, "git-repository")
	applicationDirectory := cli.MustGetString(cmd, "application-directory")
	applicationPackage := cli.MustGetString(cmd, "application-package")
	bindAddress := cli.MustGetString(cmd, "bind")
	envVars := cli.MustGetStringArray(cmd, "env")

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
		return application.Build(buildCtx, log, sourcePathV1, applicationPackage, execFileV1)
	})
	buildGroup.Go(func() error {
		return application.Build(buildCtx, log, sourcePathV2, applicationPackage, execFileV2)
	})
	err = buildGroup.Wait()
	if err != nil {
		return err
	}
	log.Info("-> all builds finished successfully")

	var mErrMutex sync.Mutex
	var mErr error
	wg := sync.WaitGroup{}
	startPort := 3000
	for i, execFile := range []string{execFileV1, execFileV2} {
		wg.Add(1)
		go func(i int, execFile string) {
			defer wg.Done()
			runEnv := append([]string{
				fmt.Sprintf("BIND_ADDRESS=%s:%d", bindAddress, startPort+i),
			}, envVars...)
			appErr := application.Run(ctx, log, execFile, runEnv)
			if appErr != nil {
				log.Warnf("-> application %s exited with error: %v", execFile, appErr)
				mErrMutex.Lock()
				mErr = multierror.Append(mErr, appErr)
				mErrMutex.Unlock()
			}
		}(i, execFile)
	}
	wg.Wait()

	if errors.Is(mErr, context.Canceled) {
		log.Warnf("-> applications stopped")
		return nil
	}
	return err
}

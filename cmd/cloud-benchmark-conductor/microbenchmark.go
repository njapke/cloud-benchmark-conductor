package main

import (
	"github.com/christophwitzko/master-thesis/pkg/cli"
	"github.com/christophwitzko/master-thesis/pkg/config"
	"github.com/christophwitzko/master-thesis/pkg/gcloud"
	"github.com/christophwitzko/master-thesis/pkg/gcloud/run"
	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func microbenchmarkCmd(log *logger.Logger) *cobra.Command {
	return &cobra.Command{
		Use:     "microbenchmark",
		Aliases: []string{"mb", "micro"},
		Short:   "Run microbenchmarks in the cloud",
		Args:    cobra.NoArgs,
		Run:     cli.WrapRunE(log, microbenchmarkRun),
	}
}

func microbenchmarkRun(log *logger.Logger, cmd *cobra.Command, args []string) error {
	conf, err := config.NewConductorConfig(cmd)
	if err != nil {
		return err
	}
	service, err := gcloud.NewService(conf)
	if err != nil {
		return err
	}
	defer service.Close()

	ctx, cancel := cli.NewContext(conf.Timeout)
	defer cancel()

	log.Info("setting up firewall rules...")
	if err := service.EnsureFirewallRules(ctx); err != nil {
		return err
	}

	log.Infof("setting up %d instances...", conf.Microbenchmark.Runs)
	errGroup, ctx := errgroup.WithContext(ctx)
	for runIndex := 1; runIndex <= conf.Microbenchmark.Runs; runIndex++ {
		runIndex := runIndex
		errGroup.Go(func() error {
			return run.Microbenchmark(ctx, log, service, runIndex)
		})
	}

	if err := errGroup.Wait(); err != nil {
		return err
	}
	log.Info("done")
	return nil
}

package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/christophwitzko/master-thesis/pkg/cli"
	"github.com/christophwitzko/master-thesis/pkg/config"
	"github.com/christophwitzko/master-thesis/pkg/gcloud"
	"github.com/christophwitzko/master-thesis/pkg/gcloud/run"
	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func mbCmd(log *logger.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "mb",
		Short: "Run benchmarks in the cloud",
		Run:   cli.WrapRunE(log, mbRun),
	}
}

func mbRun(log *logger.Logger, cmd *cobra.Command, args []string) error {
	conf, err := config.NewConductorConfig(cmd)
	if err != nil {
		return err
	}
	service, err := gcloud.NewService(conf)
	if err != nil {
		return err
	}
	defer service.Close()

	// maximum runtime: 30 minutes
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*30)
	defer cancel()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("setting up firewall rules...")
	if err := service.EnsureFirewallRules(ctx); err != nil {
		return err
	}

	errGroup, ctx := errgroup.WithContext(ctx)
	errGroup.Go(func() error {
		return run.Microbenchmark(ctx, log, service, run.MicrobenchmarkConfig{
			Repository:    conf.Microbenchmark.Repository,
			V1:            conf.Microbenchmark.V1,
			V2:            conf.Microbenchmark.V2,
			ExcludeFilter: conf.Microbenchmark.ExcludeFilter,
			RunIndex:      1,
		})
	})

	if err := errGroup.Wait(); err != nil {
		return err
	}
	log.Info("done")
	return nil
}

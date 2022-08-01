package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/christophwitzko/master-thesis/pkg/cli"
	"github.com/christophwitzko/master-thesis/pkg/config"
	"github.com/christophwitzko/master-thesis/pkg/gcloud"
	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func mbCmd(log *logger.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "mb",
		Short: "Run microbenchmarks in the cloud",
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

	// maximum runtime: 10 minutes
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*30)
	defer cancel()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("setting up firewall rules...")
	if err := service.EnsureFirewallRules(ctx); err != nil {
		return err
	}

	log.Info("setting up instance...")
	instance, err := service.CreateInstance(ctx, "test")
	if err != nil {
		return err
	}
	// close open ssh connection
	defer instance.Close()

	log.Infof("[%s] instance up (%s)", instance.Name(), instance.ExternalIP())
	errGroup, ctx := errgroup.WithContext(ctx)

	errGroup.Go(func() error {
		for {
			<-time.After(time.Second * 5)
			if err := instance.RunWithLog(ctx, log, "uptime"); err != nil {
				log.Errorf("[%s] error running uptime: %v", instance.Name(), err)
			}
		}
	})

	errGroup.Go(func() error {
		err := instance.ExecuteActions(ctx, gcloud.NewActionInstallGo(log), gcloud.NewActionInstallMicrobenchmarkRunner(log))
		if err != nil {
			return err
		}
		err = instance.RunWithLog(ctx, log, "microbenchmark-runner --v1 main --v2 main --git-repository https://github.com/christophwitzko/flight-booking-service.git --exclude-filter=\"^chi.*$\"")
		return err
	})

	if err := errGroup.Wait(); err != nil {
		return err
	}
	log.Info("done")
	return nil
}

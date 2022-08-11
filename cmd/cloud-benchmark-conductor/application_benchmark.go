package main

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/christophwitzko/master-thesis/pkg/cli"
	"github.com/christophwitzko/master-thesis/pkg/config"
	"github.com/christophwitzko/master-thesis/pkg/gcloud"
	"github.com/christophwitzko/master-thesis/pkg/gcloud/run"
	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/spf13/cobra"
)

func applicationBenchmarkCmd(log *logger.Logger) *cobra.Command {
	return &cobra.Command{
		Use:     "application-benchmark",
		Aliases: []string{"ab", "app"},
		Short:   "Run application benchmarks in the cloud",
		Run:     cli.WrapRunE(log, applicationBenchmarkRun),
	}
}

func applicationBenchmarkRun(log *logger.Logger, cmd *cobra.Command, args []string) error {
	conf, err := config.NewConductorConfig(cmd)
	if err != nil {
		return err
	}
	service, err := gcloud.NewService(conf)
	if err != nil {
		return err
	}
	defer service.Close()

	ctx, cancel := context.WithTimeout(context.Background(), conf.Timeout)
	defer cancel()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("setting up firewall rules...")
	err = service.EnsureFirewallRules(ctx)
	if err != nil {
		return err
	}
	log.Info("running application benchmarks...")

	internalIPCh := make(chan string)
	appErrCh := make(chan error)
	go func() {
		defer close(appErrCh)
		appErr := run.Application(ctx, log, service, internalIPCh)
		if appErr != nil && !errors.Is(appErr, context.Canceled) {
			log.Errorf("error running application: %s", appErr)
		}
		appErrCh <- appErr
	}()

	internalIP := <-internalIPCh
	if internalIP == "" {
		// some error happened during application setup
		return <-appErrCh
	}

	targets := []string{
		fmt.Sprintf("%s:3000", internalIP),
		fmt.Sprintf("%s:3001", internalIP),
	}
	log.Infof("starting benchmarks on internal IP: %s", internalIP)
	err = run.ApplicationBenchmark(ctx, log, service, targets)
	if err != nil {
		log.Errorf("error running application benchmark: %s", err)
	}
	log.Infof("stopping applications...")
	cancel()
	err = <-appErrCh
	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	log.Info("done")
	return nil
}

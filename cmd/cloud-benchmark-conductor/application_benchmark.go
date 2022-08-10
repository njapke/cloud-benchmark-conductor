package main

import (
	"context"
	"fmt"
	"os/signal"
	"strings"
	"syscall"

	"github.com/christophwitzko/master-thesis/pkg/assets"
	"github.com/christophwitzko/master-thesis/pkg/cli"
	"github.com/christophwitzko/master-thesis/pkg/config"
	"github.com/christophwitzko/master-thesis/pkg/gcloud"
	"github.com/christophwitzko/master-thesis/pkg/gcloud/actions"
	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func applicationBenchmarkCmd(log *logger.Logger) *cobra.Command {
	return &cobra.Command{
		Use:     "application-benchmark",
		Aliases: []string{"ab", "app"},
		Short:   "Run application benchmarks in the cloud",
		Run:     cli.WrapRunE(log, applicationBenchmarkRun),
	}
}

func runWrk(ctx context.Context, log *logger.Logger, service gcloud.Service, id, endpoint string) error {
	instance, err := service.GetOrCreateInstance(ctx, "wrk-"+id)
	if err != nil {
		return err
	}
	defer instance.Close()
	logFn := func(stdout, stderr string) {
		log.Infof("|wrk:%s| %s%s", endpoint, stdout, stderr)
	}
	err = instance.RunWithLogger(ctx, logFn, "sudo apt-get update")
	if err != nil {
		return err
	}
	err = instance.RunWithLogger(ctx, logFn, "sudo apt-get -y install wrk")
	if err != nil {
		return err
	}
	stdout, stderr, err := instance.Run(ctx, "wrk -t5 -c10 -d10s http://"+endpoint+"/")
	if err != nil {
		return fmt.Errorf("error running wrk: %w STDOUT: %s\nSTDERR: %s", err, stdout, stderr)
	}
	logFn(stdout, stderr)
	return nil
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

	appInstance, err := service.GetOrCreateInstance(ctx, "apps")
	if err != nil {
		return err
	}
	defer appInstance.Close()

	err = appInstance.ExecuteActions(ctx,
		actions.NewActionInstallGo(log),
		actions.NewActionInstallBinary(log, "application-runner", assets.ApplicationRunner),
	)
	if err != nil {
		return err
	}

	errGroup, ctx := errgroup.WithContext(ctx)
	errGroup.Go(func() error {
		return appInstance.RunWithLogger(ctx, func(stdout, stderr string) {
			if strings.Contains(stderr, "GET") {
				return
			}
			log.Infof("|app-runner| %s%s", stdout, stderr)
		}, "application-runner --v1 main --v2 main --git-repository='https://github.com/christophwitzko/go-benchmark-tests.git' --bind 0.0.0.0")
	})
	errGroup.Go(func() error {
		return runWrk(ctx, log, service, "1", appInstance.InternalIP()+":3000")
	})
	errGroup.Go(func() error {
		return runWrk(ctx, log, service, "2", appInstance.InternalIP()+":3001")
	})

	err = errGroup.Wait()
	if err != nil {
		return err
	}
	log.Info("done")
	return nil
}

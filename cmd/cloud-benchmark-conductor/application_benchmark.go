package main

import (
	"context"
	"errors"
	"os/signal"
	"syscall"
	"time"

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

// func runWrk(ctx context.Context, log *logger.Logger, service gcloud.Service, id, endpoint string) error {
// 	instance, err := service.GetOrCreateInstance(ctx, "wrk-"+id)
// 	if err != nil {
// 		return err
// 	}
// 	defer instance.Close()
// 	logFn := func(stdout, stderr string) {
// 		log.Infof("|wrk:%s| %s%s", endpoint, stdout, stderr)
// 	}
// 	for i := 0; i < 100; i++ {
// 		select {
// 		case <-ctx.Done():
// 			return ctx.Err()
// 		default:
// 		}
// 		stdout, stderr, err := instance.Run(ctx, "curl http://"+endpoint+"/")
// 		if err != nil {
// 			log.Error(err)
// 			<-time.After(time.Second)
// 		}
// 		logFn(stdout, stderr)
// 	}
// 	return nil
// }

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

	log.Infof("starting benchmarks on internal IP: %s", internalIP)
	// errGroup, groupCtx := errgroup.WithContext(ctx)
	// errGroup.Go(func() error {
	// 	return runWrk(groupCtx, log, service, "1", internalIP+":3000")
	// })
	// errGroup.Go(func() error {
	// 	return runWrk(groupCtx, log, service, "2", internalIP+":3001")
	// })
	//
	// err = errGroup.Wait()
	// if err != nil {
	// 	return err
	// }
	select {
	case <-ctx.Done():
	case <-time.After(10 * time.Minute):
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

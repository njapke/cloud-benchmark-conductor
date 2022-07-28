package main

import (
	"context"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/christophwitzko/master-thesis/pkg/cli"
	"github.com/christophwitzko/master-thesis/pkg/config"
	"github.com/christophwitzko/master-thesis/pkg/gcloud"
	"github.com/spf13/cobra"
)

var mbCmd = &cobra.Command{
	Use:   "mb",
	Short: "Run microbenchmarks in the cloud",
	Run:   cli.WrapRunE(mbRun),
}

func init() {
	rootCmd.AddCommand(mbCmd)
}

func mbRun(cmd *cobra.Command, args []string) error {
	conf, err := config.NewConductorConfig(cmd)
	if err != nil {
		return err
	}
	service, err := gcloud.NewService(conf)
	if err != nil {
		return err
	}

	// maximum runtime: 10 minutes
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*30)
	defer cancel()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Println("setting up firewall rules")
	if err := service.EnsureFirewallRules(ctx); err != nil {
		return err
	}
	log.Println("setting up instance")
	//instance, err := service.CreateInstance(ctx, "test")
	instance, err := service.GetInstance(ctx, "test")
	if err != nil {
		return err
	}
	// close open ssh connection
	defer instance.Close()
	log.Printf("[%s]: instance up (%s)\n", instance.Name(), instance.ExternalIP())
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		if err := instance.Run(ctx, log, "ping 1.1.1.1"); err != nil {
			log.Println(err)
		}
		wg.Done()
	}()
	go func() {
		if err := instance.Run(ctx, log, "ping 8.8.8.8"); err != nil {
			log.Println(err)
		}
		wg.Done()
	}()
	wg.Wait()
	log.Println("done")
	return nil
}

package main

import (
	"context"

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

	ctx := context.Background()

	log.Println("setting up firewall rules")
	if err := service.EnsureFirewallRules(ctx); err != nil {
		return err
	}
	log.Println("setting up instance")
	instance, err := service.CreateInstance(ctx, "test")
	if err != nil {
		return err
	}
	log.Printf("[%s]: instance up (%s)\n", instance.Name(), instance.ExternalIP())
	if err := instance.WaitForSSHPortReady(ctx); err != nil {
		return err
	}
	log.Println("instance ready")
	return nil
}

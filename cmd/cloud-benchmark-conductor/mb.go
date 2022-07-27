package main

import (
	"context"
	"log"
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
	ctx, cancelCtx := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelCtx()

	//instance, err := service.CreateInstance(ctx, "test-1")
	//if err != nil {
	//	return err
	//}
	//log.Println(instance.Name(), instance.ExternalIP())
	//return service.DeleteInstances(ctx)
	log.Println(service.EnsureFirewallRules(ctx))
	log.Println(service.Cleanup(ctx))
	return nil
}

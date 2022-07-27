package main

import (
	"context"
	"log"
	"time"

	"github.com/christophwitzko/master-thesis/pkg/cli"
	"github.com/christophwitzko/master-thesis/pkg/gcloud"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(mbCmd())
}

func mbCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mb",
		Short: "microbenchmark runner",
		Run:   cli.WrapRunE(mbRun),
	}

	return cmd
}

func mbRun(cmd *cobra.Command, args []string) error {
	project, _ := cmd.Flags().GetString("project")
	region, _ := cmd.Flags().GetString("region")
	zone, _ := cmd.Flags().GetString("zone")
	service, err := gcloud.NewService(project, region, zone)
	if err != nil {
		return err
	}
	ctx, cancelCtx := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelCtx()
	log.Println(service.GetLatestUbuntuImage(ctx))
	//log.Println(service.CreateInstance(context.Background(), "test-1"))
	return nil
}

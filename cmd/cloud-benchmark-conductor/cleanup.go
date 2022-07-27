package main

import (
	"context"

	"github.com/christophwitzko/master-thesis/pkg/cli"
	"github.com/christophwitzko/master-thesis/pkg/config"
	"github.com/christophwitzko/master-thesis/pkg/gcloud"
	"github.com/spf13/cobra"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Delete all cloud resources created by the cloud benchmark conductor",
	Run: cli.WrapRunE(func(cmd *cobra.Command, args []string) error {
		conf, err := config.NewConductorConfig(cmd)
		if err != nil {
			return err
		}

		service, err := gcloud.NewService(conf)
		if err != nil {
			return err
		}
		log.Println("cleanup started...")
		deletedResources, err := service.Cleanup(context.Background())
		if err != nil {
			return err
		}
		for _, resource := range deletedResources {
			log.Printf("deleted %s", resource)
		}
		log.Println("cleanup finished")
		return nil
	}),
}

func init() {
	rootCmd.AddCommand(cleanupCmd)
}

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
	Short: "Delete cloud resources created by the cloud benchmark conductor",
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
		cleanupAll := cli.MustGetBool(cmd, "all")
		var deletedResources []string
		if cleanupAll {
			deletedResources, err = service.Cleanup(context.Background())
		} else {
			log.Println("deleting instances only...")
			deletedResources, err = service.DeleteInstances(context.Background())
		}
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
	cleanupCmd.Flags().Bool("all", false, "delete all resources created by the cloud benchmark conductor")
	rootCmd.AddCommand(cleanupCmd)
}

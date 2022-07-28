package main

import (
	"context"

	"github.com/christophwitzko/master-thesis/pkg/cli"
	"github.com/christophwitzko/master-thesis/pkg/config"
	"github.com/christophwitzko/master-thesis/pkg/gcloud"
	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/spf13/cobra"
)

func cleanupCmd(log *logger.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Delete cloud resources created by the cloud benchmark conductor",
		Run:   cli.WrapRunE(log, cleanupRun),
	}
	cmd.Flags().Bool("all", false, "delete all resources created by the cloud benchmark conductor")
	return cmd
}

func cleanupRun(log *logger.Logger, cmd *cobra.Command, args []string) error {
	conf, err := config.NewConductorConfig(cmd)
	if err != nil {
		return err
	}

	service, err := gcloud.NewService(conf)
	if err != nil {
		return err
	}

	log.Info("cleanup started...")
	cleanupAll := cli.MustGetBool(cmd, "all")
	var deletedResources []string
	if cleanupAll {
		deletedResources, err = service.Cleanup(context.Background())
	} else {
		log.Info("deleting instances only...")
		deletedResources, err = service.CleanupInstances(context.Background())
	}
	if err != nil {
		return err
	}
	for _, resource := range deletedResources {
		log.Warnf("deleted: %s", resource)
	}
	log.Info("cleanup finished")
	return nil
}

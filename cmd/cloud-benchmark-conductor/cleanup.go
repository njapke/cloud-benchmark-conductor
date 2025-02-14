package main

import (
	"github.com/christophwitzko/masters-thesis/pkg/cli"
	"github.com/christophwitzko/masters-thesis/pkg/config"
	"github.com/christophwitzko/masters-thesis/pkg/gcloud"
	"github.com/christophwitzko/masters-thesis/pkg/logger"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func cleanupCmd(log *logger.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Delete cloud resources created by the cloud benchmark conductor",
		Args:  cobra.NoArgs,
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
	defer service.Close()

	log.Info("cleanup started...")
	cleanupAll := cli.MustGetBool(cmd, "all")

	ctx, cancel := cli.NewContext(cli.DefaultTimeout)
	defer cancel()

	instanceNames, err := service.ListInstances(ctx)
	if err != nil {
		return err
	}

	log.Info("deleting instances...")
	errGroup, groupCtx := errgroup.WithContext(ctx)
	for _, name := range instanceNames {
		name := name
		errGroup.Go(func() error {
			log.Warnf("deleting instance %s", name)
			return service.DeleteInstance(groupCtx, name)
		})
	}
	err = errGroup.Wait()
	if err != nil {
		return err
	}
	if cleanupAll {
		log.Info("deleting firewall rules...")
		var firewallDeleted bool
		firewallDeleted, err = service.DeleteFirewallRules(ctx)
		if err != nil {
			return err
		}
		if firewallDeleted {
			log.Warn("firewall rules deleted")
		}
	}
	log.Info("cleanup finished")
	return nil
}

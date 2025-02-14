package main

import (
	"os"

	"github.com/christophwitzko/masters-thesis/pkg/cli"
	"github.com/christophwitzko/masters-thesis/pkg/config"
	"github.com/christophwitzko/masters-thesis/pkg/logger"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func configCmd(log *logger.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Print the current benchmark config",
		Args:  cobra.NoArgs,
		Run:   cli.WrapRunE(log, configRun),
	}
}

func configRun(log *logger.Logger, cmd *cobra.Command, args []string) error {
	c, err := config.NewConductorConfig(cmd)
	if err != nil {
		return err
	}

	return yaml.NewEncoder(os.Stdout).Encode(c)
}

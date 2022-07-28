package main

import (
	"fmt"

	"github.com/christophwitzko/master-thesis/pkg/cli"
	"github.com/christophwitzko/master-thesis/pkg/config"
	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func configCmd(log *logger.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Print the current benchmark config",
		Run:   cli.WrapRunE(log, configRun),
	}
}

func configRun(log *logger.Logger, cmd *cobra.Command, args []string) error {
	c, err := config.NewConductorConfig(cmd)
	if err != nil {
		return err
	}

	cfgStr, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	fmt.Printf("# cloud benchmark conductor config\n%s", cfgStr)
	return nil
}

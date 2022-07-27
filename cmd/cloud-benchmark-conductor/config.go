package main

import (
	"fmt"

	"github.com/christophwitzko/master-thesis/pkg/cli"
	"github.com/christophwitzko/master-thesis/pkg/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Print the current benchmark config",
	Run: cli.WrapRunE(func(cmd *cobra.Command, args []string) error {
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
	}),
}

func init() {
	rootCmd.AddCommand(configCmd)
}

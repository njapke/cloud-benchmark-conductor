package main

import (
	"os"

	"github.com/christophwitzko/master-thesis/pkg/config"
	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	log := logger.New()
	rootCmd := &cobra.Command{
		Use:   "cloud-benchmark-conductor",
		Short: "cloud benchmark conductor",
		Long: `The cloud benchmark conductor takes care of running benchmarks in the cloud.
Therefore compute instances are provisioned and used to execute the benchmarks.`,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	cobra.OnInitialize(func() {
		if err := config.InitConfig(rootCmd, "cbc.yaml"); err != nil {
			log.Errorf("Config error: %v", err)
			os.Exit(1)
		}
		usedConfigFile := viper.ConfigFileUsed()
		if usedConfigFile != "" {
			log.Infof("using config: %s", usedConfigFile)
		}
	})
	config.ConductorSetupFlagsAndViper(rootCmd)
	rootCmd.AddCommand(configCmd(log), cleanupCmd(log), mbCmd(log))
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

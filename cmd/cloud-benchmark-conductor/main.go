package main

import (
	"os"

	"github.com/christophwitzko/master-thesis/pkg/config"
	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var log = logger.Default()

var rootCmd = &cobra.Command{
	Use:   "cloud-benchmark-conductor",
	Short: "cloud benchmark conductor",
	Long: `The cloud benchmark conductor takes care of running benchmarks in the cloud.
Therefore compute instances are provisioned and used to execute the benchmarks.`,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func main() {
	cobra.OnInitialize(func() {
		if err := config.InitConfig(rootCmd, "cbc.yaml"); err != nil {
			log.Printf("Config error: %v\n", err)
			os.Exit(1)
		}
		usedConfigFile := viper.ConfigFileUsed()
		if usedConfigFile != "" {
			log.Printf("using config: %s\n", usedConfigFile)
		}
	})
	config.ConductorSetupFlagsAndViper(rootCmd)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

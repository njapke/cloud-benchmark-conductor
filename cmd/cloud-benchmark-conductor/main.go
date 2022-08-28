package main

import (
	"os"

	"github.com/christophwitzko/master-thesis/pkg/cli"
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
		Args: cobra.NoArgs,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.Info(cli.GetBuildInfo())
		},
	}
	cobra.OnInitialize(func() {
		if err := config.InitConfig(rootCmd, "cbc.yaml"); err != nil {
			log.Errorf("Config error: %v", err)
			os.Exit(1)
		}
		usedConfigFile := viper.ConfigFileUsed()
		if usedConfigFile != "" {
			log.Infof("using config: %s", cli.GetRelativePath(usedConfigFile))
		}
	})
	config.ConductorSetupFlagsAndViper(rootCmd)
	rootCmd.AddCommand(
		configCmd(log),
		cleanupCmd(log),
		microbenchmarkCmd(log),
		applicationBenchmarkCmd(log),
	)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

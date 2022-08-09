package main

import (
	"os"

	"github.com/christophwitzko/master-thesis/pkg/cli"
	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/spf13/cobra"
)

func main() {
	log := logger.New()
	rootCmd := &cobra.Command{
		Use:   "application-benchmark-runner",
		Short: "application benchmark runner tool",
		Long:  "This tool is used to run the application benchmarks using artillery.",
		Run:   cli.WrapRunE(log, rootRun),
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	rootCmd.Flags().String("v1", "", "source path or git reference for version 1")
	rootCmd.Flags().String("v2", "", "source path or git reference for version 2")
	rootCmd.Flags().String("git-repository", "", "git repository to use for installing the applications")
	rootCmd.Flags().String("application-directory", "/tmp/.application", "directory to use for running the application")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func rootRun(log *logger.Logger, cmd *cobra.Command, args []string) error {
	// sourcePathOrRefV1 := cli.MustGetString(cmd, "v1")
	// sourcePathOrRefV2 := cli.MustGetString(cmd, "v2")
	// gitRepository := cli.MustGetString(cmd, "git-repository")
	// applicationDirectory := cli.MustGetString(cmd, "application-directory")
	return nil
}

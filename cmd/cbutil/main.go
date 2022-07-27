package main

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cbutil",
	Short: "cloud benchmark util",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func main() {
	rootCmd.PersistentFlags().String("region", os.Getenv("CLOUDSDK_COMPUTE_REGION"), "compute region")
	rootCmd.PersistentFlags().String("zone", os.Getenv("CLOUDSDK_COMPUTE_ZONE"), "compute zone")
	rootCmd.PersistentFlags().String("project", os.Getenv("CLOUDSDK_CORE_PROJECT"), "google cloud project")
	rootCmd.Flags().SortFlags = true

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

package cli

import (
	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/spf13/cobra"
)

func WrapRunE(log *logger.Logger, fn func(log *logger.Logger, cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if err := fn(log, cmd, args); err != nil {
			log.Fatalf("ERROR: %v", err)
		}
	}
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func MustGetString(cmd *cobra.Command, name string) string {
	val, err := cmd.Flags().GetString(name)
	Must(err)
	return val
}

func MustGetBool(cmd *cobra.Command, name string) bool {
	val, err := cmd.Flags().GetBool(name)
	Must(err)
	return val
}

func MustGetInt(cmd *cobra.Command, name string) int {
	val, err := cmd.Flags().GetInt(name)
	Must(err)
	return val
}

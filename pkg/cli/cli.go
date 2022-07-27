package cli

import (
	"log"

	"github.com/spf13/cobra"
)

func WrapRunE(fn func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if err := fn(cmd, args); err != nil {
			log.Fatalf("ERROR: %v", err)
		}
	}
}

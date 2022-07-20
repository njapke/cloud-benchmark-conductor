package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

type ctxKeySourcePath int

const SourcePathKey ctxKeySourcePath = 0

func main() {
	cmd := &cobra.Command{
		Use:     "microbenchmark-runner",
		Short:   "microbenchmark runner tool",
		PreRunE: preRun,
		Run:     wrapRun(rootRun),
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	cmd.PersistentFlags().StringP("source-path", "s", "", "source path")
	cmd.PersistentFlags().SortFlags = true

	cmd.AddCommand(listCmd())

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func rootRun(cmd *cobra.Command, args []string) error {
	sourcePath := cmd.Context().Value(SourcePathKey)
	log.Printf("using %s as source path", sourcePath)

	return fmt.Errorf("not implemented")
}

func preRun(cmd *cobra.Command, args []string) error {
	sourcePath, err := cmd.Flags().GetString("source-path")
	if err != nil {
		return err
	}
	if sourcePath == "" {
		return fmt.Errorf("source path not configured")
	}
	cmd.SetContext(context.WithValue(cmd.Context(), SourcePathKey, sourcePath))
	return nil
}

func wrapRun(fn func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if err := fn(cmd, args); err != nil {
			log.Fatal(err)
		}
	}
}

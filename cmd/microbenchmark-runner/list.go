package main

import (
	"log"
	"master-thesis/pkg/benchmark"

	"github.com/spf13/cobra"
)

func listCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "list all benchmark functions of the given source path",
		PreRunE: preRun,
		Run:     wrapRun(listRun),
	}
	return cmd
}

func listRun(cmd *cobra.Command, args []string) error {
	sourcePath := cmd.Context().Value(SourcePathKey).(string)
	log.Printf("listing benchmarks for %s", sourcePath)
	functions, err := benchmark.GetFunctions(sourcePath)
	if err != nil {
		return err
	}
	for _, fn := range functions {
		log.Printf("[%s]: %s (%s)", fn.Directory, fn.Name, fn.PackageName)
	}
	return nil
}

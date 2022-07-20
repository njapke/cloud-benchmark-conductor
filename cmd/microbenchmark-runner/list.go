package main

import (
	"encoding/json"
	"log"
	"os"

	"master-thesis/pkg/benchmark"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd())
}

func listCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all benchmark functions of the given source path",
		Run:   wrapRunE(listRun),
	}

	cmd.Flags().StringP("source-path", "s", "", "source path")
	_ = cmd.MarkFlagRequired("source-path")
	cmd.Flags().Bool("json", false, "output in json format")
	cmd.Flags().SortFlags = true

	return cmd
}

func listRun(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()
	sourcePath, _ := flags.GetString("source-path")
	outputJSON, _ := flags.GetBool("json")
	log.Printf("listing benchmarks for %s", sourcePath)
	functions, err := benchmark.GetFunctions(sourcePath)
	if err != nil {
		return err
	}

	if outputJSON {
		return json.NewEncoder(os.Stdout).Encode(functions)
	}
	for _, fn := range functions {
		log.Printf("[%s]: %s (%s)", fn.Directory, fn.Name, fn.PackageName)
	}
	return nil
}

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

	cmd.Flags().StringP("source-path-v1", "1", "", "source path for version 1")
	_ = cmd.MarkFlagRequired("source-path-v1")
	cmd.Flags().StringP("source-path-v2", "2", "", "source path for version 2")
	_ = cmd.MarkFlagRequired("source-path-v2")
	cmd.Flags().Bool("json", false, "output in json format")
	cmd.Flags().SortFlags = true

	return cmd
}

func listRun(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()
	sourcePathV1, _ := flags.GetString("source-path-v1")
	sourcePathV2, _ := flags.GetString("source-path-v2")
	outputJSON, _ := flags.GetBool("json")
	log.Printf("listing benchmarks for %s and %s", sourcePathV1, sourcePathV2)
	functionsV1, err := benchmark.GetFunctions(sourcePathV1)
	if err != nil {
		return err
	}

	functionsV2, err := benchmark.GetFunctions(sourcePathV2)
	if err != nil {
		return err
	}

	combinedFunctions := benchmark.CombineFunctions(functionsV1, functionsV2)

	if outputJSON {
		return json.NewEncoder(os.Stdout).Encode(combinedFunctions)
	}
	for _, fn := range combinedFunctions {
		log.Printf("%s (%s)", fn.V1.Name, fn.V1.PackageName)
		log.Printf("--> %s", fn.V1.Directory)
		log.Printf("--> %s", fn.V2.Directory)
	}
	return nil
}

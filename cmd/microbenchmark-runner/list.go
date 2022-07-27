package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/christophwitzko/master-thesis/pkg/benchmark"
	"github.com/christophwitzko/master-thesis/pkg/cli"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all overlapping benchmark functions of the given source paths",
	Run:   cli.WrapRunE(listRun),
}

func init() {
	listCmd.Flags().StringP("source-path-v1", "1", "", "source path for version 1")
	_ = listCmd.MarkFlagRequired("source-path-v1")
	listCmd.Flags().StringP("source-path-v2", "2", "", "source path for version 2")
	_ = listCmd.MarkFlagRequired("source-path-v2")
	listCmd.Flags().Bool("json", false, "output in json format")
	listCmd.Flags().SortFlags = true
	rootCmd.AddCommand(listCmd)
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

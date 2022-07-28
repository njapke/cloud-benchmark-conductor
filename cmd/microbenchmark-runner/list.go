package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/christophwitzko/master-thesis/pkg/benchmark"
	"github.com/christophwitzko/master-thesis/pkg/cli"
	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/spf13/cobra"
)

func listCmd(log *logger.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all overlapping benchmark functions of the given source paths",
		Run:   cli.WrapRunE(log, listRun),
	}
	cmd.Flags().Bool("json", false, "output in json format")
	cmd.Flags().SortFlags = true
	return cmd
}

func listRun(log *logger.Logger, cmd *cobra.Command, args []string) error {
	sourcePathV1 := cli.MustGetString(cmd, "v1")
	sourcePathV2 := cli.MustGetString(cmd, "v2")
	outputJSON := cli.MustGetBool(cmd, "json")

	if sourcePathV1 == "" || sourcePathV2 == "" {
		return fmt.Errorf("source-path-v1 and source-path-v2 are required")
	}

	log.Infof("listing benchmarks for %s and %s", sourcePathV1, sourcePathV2)

	combinedFunctions, err := benchmark.CombinedFunctionsFromPaths(sourcePathV1, sourcePathV2)
	if err != nil {
		return err
	}
	if outputJSON {
		return json.NewEncoder(os.Stdout).Encode(combinedFunctions)
	}
	for _, fn := range combinedFunctions {
		log.Infof("%s (%s)", fn.V1.Name, fn.V1.PackageName)
		log.Infof("--> %s", fn.V1.Directory)
		log.Infof("--> %s", fn.V2.Directory)
	}
	return nil
}

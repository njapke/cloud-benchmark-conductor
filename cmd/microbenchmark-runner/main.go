package main

import (
	"encoding/csv"
	"encoding/json"
	"log"
	"master-thesis/pkg/benchmark"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "microbenchmark-runner",
	Short: "microbenchmark runner tool",
	Run:   wrapRunE(rootRun),
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func main() {
	rootCmd.Flags().StringP("input-file", "i", "", "input file")
	_ = rootCmd.MarkFlagRequired("input-file")
	rootCmd.Flags().IntP("run", "r", 1, "run index")
	_ = rootCmd.MarkFlagRequired("run")

	rootCmd.Flags().SortFlags = true

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func readInput(filename string) ([]benchmark.VersionedFunction, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var functions []benchmark.VersionedFunction
	err = json.NewDecoder(f).Decode(&functions)
	if err != nil {
		return nil, err
	}
	return functions, nil
}

func rootRun(cmd *cobra.Command, args []string) error {
	inputFile, _ := cmd.Flags().GetString("input-file")
	runIndex, _ := cmd.Flags().GetInt("run")

	functions, err := readInput(inputFile)
	if err != nil {
		return err
	}

	resultsV1 := make(benchmark.Results, 0)
	resultsV2 := make(benchmark.Results, 0)
	for s := 1; s <= 3; s++ {
		log.Printf("suite run: %d\n", s)
		rV1, rV2, err := benchmark.RunSuite(functions, runIndex, s)
		if err != nil {
			return err
		}
		resultsV1 = append(resultsV1, rV1...)
		resultsV2 = append(resultsV2, rV2...)
	}

	w := csv.NewWriter(os.Stdout)
	_ = w.WriteAll(resultsV1.Records())
	_ = w.WriteAll(resultsV2.Records())
	if err := w.Error(); err != nil {
		return err
	}
	log.Println("done")
	return nil
}

func wrapRunE(fn func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if err := fn(cmd, args); err != nil {
			log.Fatalf("ERROR: %v", err)
		}
	}
}

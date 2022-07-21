package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
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

	rootCmd.Flags().SortFlags = true

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func readInput(filenname string) ([]benchmark.VersionedFunction, error) {
	f, err := os.Open(filenname)
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

func randomBool() bool {
	b := [1]byte{}
	_, _ = rand.Read(b[:])
	return b[0] > 127
}

func v1OrV2(b bool) string {
	if b {
		return "v1"
	}
	return "v2"
}

func rootRun(cmd *cobra.Command, args []string) error {
	inputFile, _ := cmd.Flags().GetString("input-file")

	functions, err := readInput(inputFile)
	if err != nil {
		return err
	}

	resultsV1 := make([]benchmark.Result, 0)
	resultsV2 := make([]benchmark.Result, 0)

	for s := 1; s <= 3; s++ {
		log.Printf("suite run: %d\n", s)
		rV1, rV2, err := benchmark.RunSuite(functions, s)
		if err != nil {
			return err
		}
		resultsV1 = append(resultsV1, rV1...)
		resultsV2 = append(resultsV2, rV2...)
	}

	for _, r := range resultsV1 {
		fmt.Printf("%#v\n", r)
	}
	fmt.Println("-----------------------------------------------------")
	for _, r := range resultsV2 {
		fmt.Printf("%#v\n", r)
	}
	fmt.Println("done")
	return nil
}

func wrapRunE(fn func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if err := fn(cmd, args); err != nil {
			log.Fatalf("ERROR: %v", err)
		}
	}
}

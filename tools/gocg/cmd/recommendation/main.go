package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/sealuzh/gocg/pkg/cg"
	"bitbucket.org/sealuzh/gocg/pkg/dir"
	"bitbucket.org/sealuzh/gocg/pkg/overlap"
	"bitbucket.org/sealuzh/gocg/pkg/rec"
)

const (
	expArgs = 5
)

// recommendation command command
// arguments:
//   1) project name(s) (comma ',' seperated)
//   2) system benchmark folder
//   3) microbenchmark folder
//   4) out folder
//   5) number of new microbenchmarks
func main() {
	projects, system, micro, out, nrBenchs := parseArgs()
	defer system.Close()
	defer micro.Close()
	defer out.Close()

	fmt.Printf("# Recommend microbenchmarks for project '%s'\n", projects)

	start := time.Now()
	scenario := filepath.Base(system.Name())

	cgRes, err := cg.FromDotsSystemMicro(system.Name(), micro.Name())
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not transform dot files to CGs: %v\n", err)
		os.Exit(1)
	}

	overlaps, err := overlap.Structurals(projects, cgRes, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get overlaps: %v\n", err)
		os.Exit(1)
	}

	recFunc := rec.StrategyGreedyAdditional

	for _, nrbs := range nrBenchs {
		runForNrBenchs(projects, scenario, cgRes, overlaps, out.Name(), nrbs, recFunc)
	}

	fmt.Printf("# Finished recommendation in %v\n", time.Since(start))
}

func runForNrBenchs(projects []string, scenario string, cgRes []*cg.Result, overlaps []*overlap.System, outDir string, nrbs int, recFunc rec.StrategyFunc) {
	rs, err := rec.ApplyAll(projects, cgRes, overlaps, recFunc, nrbs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "nrBenchs=%d: could not apply recommendations: %v\n", nrbs, err)
		return
	}

	recFilePath := filepath.Join(outDir, fmt.Sprintf("%s_recFile_recs-%d.csv", scenario, nrbs))
	recFile, err := os.Create(recFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "nrBenchs=%d: could not create recommendation file '%s': %v\n", nrbs, recFilePath, err)
		return
	}
	defer recFile.Close()

	err = rec.WriteAll(projects, rs, nrbs, recFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "nrBenchs=%d: could not write recommendations: %v\n", nrbs, err)
		return
	}

	// write new overlaps
	structNodeOverlapPath := filepath.Join(outDir, fmt.Sprintf("%s_struct_node_overlap_recs-%d.csv", scenario, nrbs))
	structNodeOverlapFile, err := os.Create(structNodeOverlapPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "nrBenchs=%d: could not create structural node overlap file '%s': %v\n", nrbs, structNodeOverlapPath, err)
		return
	}
	defer structNodeOverlapFile.Close()

	newCGRes := make([]*cg.Result, len(rs))
	for i, r := range rs {
		newCGRes[i] = r.CG
	}

	err = overlap.StructuralsWrite(projects, newCGRes, true, structNodeOverlapFile, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "nrBenchs=%d: could not get structural overlap for all nodes: %v\n", nrbs, err)
		return
	}
}

func parseArgs() (projects []string, system, micro, out *os.File, nrBenchs []int) {
	args := os.Args
	if largs := len(args) - 1; largs != expArgs {
		fmt.Fprintf(os.Stderr, "invalid argument number: expected %d was %d\n", expArgs, largs)
		os.Exit(2)
	}

	projectsSlice := strings.Split(os.Args[1], ",")

	system = dir.FromPath(args[2], "system")
	micro = dir.FromPath(args[3], "micro")
	out = dir.FromPath(args[4], "out")

	nrBenchs = parseNrBenchs(args[5])

	return projectsSlice, system, micro, out, nrBenchs
}

func parseNrBenchs(arg string) []int {
	arg = strings.TrimSpace(arg)
	if len(arg) == 0 {
		return []int{}
	}
	argSlice := strings.Split(arg, ",")
	var ret []int

	for _, benchArg := range argSlice {
		benchArg = strings.TrimSpace(benchArg)
		if len(benchArg) == 0 {
			continue
		}

		nrBenchs, err := strconv.Atoi(benchArg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not parse nrBenchs into int: %v\n", err)
			os.Exit(3)
		}
		if nrBenchs < 0 {
			fmt.Fprintf(os.Stderr, "invalid nrBenchs value: must be positive but was %d\n", nrBenchs)
			os.Exit(4)
		}

		ret = append(ret, nrBenchs)
	}

	return ret
}

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	"bitbucket.org/sealuzh/gocg/pkg/cg"
	"bitbucket.org/sealuzh/gocg/pkg/dir"
	"bitbucket.org/sealuzh/gocg/pkg/min"
	"bitbucket.org/sealuzh/gocg/pkg/overlap"
)

const (
	expArgs = 4
)

// minimization command command
// arguments:
//   1) project name(s) (comma ',' seperated)
//   2) system benchmark folder
//   3) microbenchmark folder
//   4) out folder
func main() {
	projects, system, micro, out := parseArgs()
	defer system.Close()
	defer micro.Close()
	defer out.Close()

	fmt.Printf("# Minimize microbenchmarks for project '%s'\n", projects)

	start := time.Now()
	scenario := filepath.Base(system.Name())

	cgRes, err := cg.FromDotsSystemMicro(system.Name(), micro.Name())
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not transform dot files to CGs: %v\n", err)
		os.Exit(1)
	}

	minStrategies := []min.StrategyFunc{min.GreedyMicro, min.GreedySystem}

	for _, strat := range minStrategies {
		runWithStrategy(projects, scenario, cgRes, out.Name(), strat, true, true)
		runWithStrategy(projects, scenario, cgRes, out.Name(), strat, false, false)
	}

	fmt.Printf("# Finished minimization in %v\n", time.Since(start))
}

func funcName(f interface{}) string {
	fqn := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	dotPos := strings.LastIndex(fqn, ".")
	fn := fqn[dotPos+1:]
	return fn
}

func runWithStrategy(projects []string, scenario string, cgRes []*cg.Result, outDir string, strat min.StrategyFunc, projectOnly, writeHeader bool) {
	fn := funcName(strat)

	overlaps, err := overlap.Structurals(projects, cgRes, projectOnly)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get overlaps: %v\n", err)
		os.Exit(1)
	}

	excl := cg.ExclusionNot(cg.IsProjects(projects))

	rs, err := min.ApplyAll(projects, cgRes, overlaps, strat, excl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "strat=%s: could not apply minimization: %v\n", fn, err)
		return
	}

	minFilePath := filepath.Join(outDir, fmt.Sprintf("%s_minFile_%s.csv", scenario, fn))
	minFile, err := os.OpenFile(minFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "strat=%s: could not create minimization file '%s': %v\n", fn, minFilePath, err)
		return
	}
	defer minFile.Close()

	err = min.WriteAll(projects, rs, minFile, projectOnly, writeHeader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "strat=%s: could not write minimizations: %v\n", fn, err)
		return
	}

	// write new overlaps
	structNodeOverlapPath := filepath.Join(outDir, fmt.Sprintf("%s_struct_node_overlap_mins-%s.csv", scenario, fn))
	structNodeOverlapFile, err := os.OpenFile(structNodeOverlapPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "strat=%s: could not create structural node overlap file '%s': %v\n", fn, structNodeOverlapPath, err)
		return
	}
	defer structNodeOverlapFile.Close()

	newCGs := make([]*cg.Result, len(rs))
	for i, r := range rs {
		newCGs[i] = r.CG
	}

	err = overlap.StructuralsWrite(projects, newCGs, projectOnly, structNodeOverlapFile, writeHeader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "strat=%s: could not get structural overlap for all nodes: %v\n", fn, err)
		return
	}
}

func parseArgs() (projects []string, system, micro, out *os.File) {
	args := os.Args
	if largs := len(args) - 1; largs != expArgs {
		fmt.Fprintf(os.Stderr, "invalid argument number: expected %d was %d\n", expArgs, largs)
		os.Exit(2)
	}

	projectsSlice := strings.Split(os.Args[1], ",")

	system = dir.FromPath(args[2], "system")
	micro = dir.FromPath(args[3], "micro")
	out = dir.FromPath(args[4], "out")

	return projectsSlice, system, micro, out
}

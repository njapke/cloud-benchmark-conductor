package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bitbucket.org/sealuzh/gocg/pkg/cg"
	"bitbucket.org/sealuzh/gocg/pkg/dir"
	"bitbucket.org/sealuzh/gocg/pkg/overlap"
)

const (
	expArgs = 4
)

// overlap command
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

	fmt.Printf("# Transform dot files into callgraphs of project(s) %s\n", projects)

	startTrans := time.Now()
	cgRes, err := cg.FromDotsSystemMicro(system.Name(), micro.Name())
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not transform dot files to CGs: %v\n", err)
		os.Exit(1)
	}

	structNodeOverlapPath := filepath.Join(out.Name(), "struct_node_overlap.csv")
	structNodeOverlapFile, err := os.Create(structNodeOverlapPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not create structural node overlap file '%s': %v\n", structNodeOverlapPath, err)
		os.Exit(1)
	}
	defer structNodeOverlapFile.Close()

	err = overlap.StructuralsWrite(projects, cgRes, false, structNodeOverlapFile, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "all: could not get structural overlap for all nodes: %v\n", err)
		os.Exit(1)
	}
	err = overlap.StructuralsWrite(projects, cgRes, true, structNodeOverlapFile, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "project-only: could not get structural overlap for project nodes: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("# Finished overlaps in %v\n", time.Since(startTrans))
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

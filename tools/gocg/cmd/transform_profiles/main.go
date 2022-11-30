package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"bitbucket.org/sealuzh/gocg/pkg/dir"
	"bitbucket.org/sealuzh/gocg/pkg/profile"
)

const (
	expMinArgs = 3
)

// transform profiles command
// arguments:
//   1) input folder
//   2) output folder
//   3-n) configs (^(type:node_count:node_fraction:edge_fraction)$)
func main() {
	in, out, configs := parseArgs()
	defer in.Close()
	defer out.Close()

	err := profile.Tos(in, in.Name(), out.Name(), configs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing profile transformation: %v", err)
		os.Exit(1)
	}
}

func parseArgs() (in, out *os.File, outConfigs []*profile.OutConfig) {
	args := os.Args
	largs := len(args)
	if actualLargs := largs - 1; actualLargs < expMinArgs {
		fmt.Fprintf(os.Stderr, "invalid argument number: expected min %d was %d\n", expMinArgs, actualLargs)
		os.Exit(2)
	}

	in = dir.FromPath(args[1], "input")
	out = dir.FromPath(args[2], "out")

	// configs
	outConfigs = make([]*profile.OutConfig, 0, largs-3)
	for i := 3; i < largs; i++ {
		oc, err := parseOutConfig(args[i])
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not parse profile.OutConfig '%s': %v\n", args[i], err)
			os.Exit(3)
		}
		outConfigs = append(outConfigs, oc)
	}

	return in, out, outConfigs
}

func parseOutConfig(s string) (*profile.OutConfig, error) {
	splitted := strings.Split(s, ":")
	if l := len(splitted); l != 4 {
		return nil, fmt.Errorf("out config does not have 4 elements, has %d", l)
	}

	t, err := profile.OutTypeFrom(splitted[0])
	if err != nil {
		return nil, fmt.Errorf("invalid out type: %v", err)
	}

	nc, err := strconv.Atoi(splitted[1])
	if err != nil {
		return nil, fmt.Errorf("invalid node count: %v", err)
	}

	nf, err := strconv.ParseFloat(splitted[2], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid node fraction: %v", err)
	}

	ef, err := strconv.ParseFloat(splitted[3], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid edge fraction: %v", err)
	}

	return &profile.OutConfig{
		Type:         t,
		NodeCount:    nc,
		NodeFraction: nf,
		EdgeFraction: ef,
	}, nil
}

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/christophwitzko/master-thesis/pkg/benchmark"
)

func split(input string, c byte) []string {
	res := make([]string, 0)
	for {
		pos := strings.IndexByte(input, c)
		if pos == -1 {
			return append(res, input)
		}
		res = append(res, input[:pos])
		input = input[pos+1:]
	}
}

func main() {
	fmt.Printf("%#v\n", split("A,B,C,D", ','))
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	fns, err := benchmark.GetFunctions("./research")
	if err != nil {
		return err
	}
	fmt.Println(fns)
	return nil
}

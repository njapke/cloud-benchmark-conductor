package benchmark

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"os/exec"
	"time"

	"golang.org/x/perf/benchfmt"
)

const Timeout = 60
const ExecutionCount = 5

var timeoutArg = fmt.Sprintf("-timeout=%ds", Timeout)
var countArg = fmt.Sprintf("-count=%d", ExecutionCount)

type Result struct {
	Ops      float64
	Bytes    float64
	Allocs   float64
	I        int
	S        int
	Function Function
}

func init() {
	rand.Seed(time.Now().Unix())
}

func NewResult(b *benchfmt.Result, i, s int, fn Function) Result {
	ops, _ := b.Value("sec/op")
	bytes, _ := b.Value("B/op")
	allocs, _ := b.Value("allocs/op")
	return Result{
		Ops:      ops,
		Bytes:    bytes,
		Allocs:   allocs,
		I:        i,
		S:        s,
		Function: fn,
	}
}

func RunFunction(f Function, suite int) ([]Result, error) {
	args := []string{
		"test",
		"-run=^$",
		"-benchmem",
		"-benchtime=1s",
		timeoutArg,
		countArg,
		fmt.Sprintf("-bench=^%s$", f.Name),
		f.Directory,
	}
	cmd := exec.Command("go", args...)
	pipeRead, pipeWrite := io.Pipe()
	cmd.Stdout = pipeWrite
	cmd.Stderr = pipeWrite
	errCh := make(chan error, 1)
	go func() {
		defer pipeWrite.Close()
		if err := cmd.Run(); err != nil {
			errCh <- err
		}
		close(errCh)
	}()

	res := make([]Result, ExecutionCount)
	i := 0
	bReader := benchfmt.NewReader(pipeRead, "bench.txt")
	for bReader.Scan() {
		switch rec := bReader.Result(); rec := rec.(type) {
		case *benchfmt.SyntaxError:
			log.Printf("syntax error: %s", rec.Error())
			continue
		case *benchfmt.Result:
			res[i] = NewResult(rec, i+1, suite, f)
			i++
		default:
			log.Printf("unknown record type: %T", rec)
			continue
		}
	}
	if err := bReader.Err(); err != nil {
		return nil, err
	}
	if err := <-errCh; err != nil {
		return nil, err
	}
	return res, nil
}

func RunVersionedFunction(vFunction VersionedFunction, suite int) ([]Result, []Result, error) {
	var resultsV1 []Result
	var resultsV2 []Result

	a, b := vFunction.V1, vFunction.V2
	aVersion, bVersion := 1, 2

	// randomly change execution order
	if rand.Intn(2) == 0 {
		a, b = vFunction.V2, vFunction.V1
		aVersion, bVersion = 2, 1
	}

	log.Printf("  |--> running[%d]: %s\n", aVersion, a.Directory)
	res, err := RunFunction(a, suite)
	if err != nil {
		return nil, nil, err
	}
	if aVersion == 1 {
		resultsV1 = res
	} else {
		resultsV2 = res
	}

	log.Printf("  |--> running[%d]: %s", bVersion, b.Directory)
	res, err = RunFunction(b, suite)
	if err != nil {
		return nil, nil, err
	}
	if bVersion == 1 {
		resultsV1 = res
	} else {
		resultsV2 = res
	}

	return resultsV1, resultsV2, nil
}

func RunSuite(fns []VersionedFunction, suite int) ([]Result, []Result, error) {
	newFns := make([]VersionedFunction, len(fns))
	copy(newFns, fns)

	// shuffle execution order
	rand.Shuffle(len(newFns), func(i, j int) {
		newFns[i], newFns[j] = newFns[j], newFns[i]
	})
	resultsV1 := make([]Result, 0)
	resultsV2 := make([]Result, 0)
	for _, function := range newFns {
		log.Printf("--| benchmarking: %s\n", function.String())
		rV1, rV2, err := RunVersionedFunction(function, suite)
		if err != nil {
			return nil, nil, err
		}
		resultsV1 = append(resultsV1, rV1...)
		resultsV2 = append(resultsV2, rV2...)
	}
	return resultsV1, resultsV2, nil
}

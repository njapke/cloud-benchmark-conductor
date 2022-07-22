package benchmark

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"time"

	"golang.org/x/perf/benchfmt"
)

const Timeout = 60
const ExecutionCount = 5

var timeoutArg = fmt.Sprintf("-timeout=%ds", Timeout)
var countArg = fmt.Sprintf("-count=%d", ExecutionCount)

type Result struct {
	Function   Function
	Iterations int
	Ops        float64 // sec/op
	Bytes      float64 // B/op
	Allocs     float64 // allocs/op
	R          int     // run index
	S          int     // suite execution
	I          int     // benchmark function index
	Version    int
}

var CSVOutputHeader = []string{
	"R-S-I",
	"package.BenchmarkFunction",
	"Version",
	"Directory",
	"Iterations",
	"sec/op",
	"B/op",
	"allocs/op",
}

func NewResult(fn Function, version, r, s, i int, b *benchfmt.Result) Result {
	ops, _ := b.Value("sec/op")
	bytes, _ := b.Value("B/op")
	allocs, _ := b.Value("allocs/op")
	return Result{
		Function:   fn,
		Iterations: b.Iters,
		Ops:        ops,
		Bytes:      bytes,
		Allocs:     allocs,
		R:          r,
		S:          s,
		I:          i,
		Version:    version,
	}
}

func (r Result) RSI() string {
	return fmt.Sprintf("%d-%d-%d", r.R, r.S, r.I)
}

func (r Result) Record() []string {
	return []string{
		r.RSI(),
		fmt.Sprintf("%s.%s", r.Function.PackageName, r.Function.Name),
		strconv.FormatInt(int64(r.Version), 10),
		r.Function.Directory,
		strconv.FormatInt(int64(r.Iterations), 10),
		strconv.FormatFloat(r.Ops, 'f', -1, 32),
		strconv.FormatFloat(r.Bytes, 'f', -1, 32),
		strconv.FormatFloat(r.Allocs, 'f', -1, 32),
	}
}

type Results []Result

func (r Results) Records() [][]string {
	res := make([][]string, len(r))
	for i, result := range r {
		res[i] = result.Record()
	}
	return res
}

func init() {
	rand.Seed(time.Now().Unix())
}

func RunFunction(csvWriter *csv.Writer, f Function, version, run, suite int) error {
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
	cmd.Stderr = os.Stderr

	errCh := make(chan error, 1)
	go func() {
		if err := cmd.Run(); err != nil {
			errCh <- err
		}
		if err := pipeWrite.Close(); err != nil {
			errCh <- err
		}
		close(errCh)
	}()

	i := 0
	bReader := benchfmt.NewReader(pipeRead, "bench.txt")
	for bReader.Scan() {
		switch rec := bReader.Result(); rec := rec.(type) {
		case *benchfmt.SyntaxError:
			log.Printf("syntax error: %s", rec.Error())
			continue
		case *benchfmt.Result:
			res := NewResult(f, version, run, suite, i+1, rec)
			if err := csvWriter.Write(res.Record()); err != nil {
				return err
			}
			csvWriter.Flush()
			i++
		default:
			log.Printf("unknown record type: %T", rec)
			continue
		}
	}
	if err := bReader.Err(); err != nil {
		return err
	}
	if err := <-errCh; err != nil {
		return err
	}
	return nil
}

func RunVersionedFunction(csvWriter *csv.Writer, vFunction VersionedFunction, run, suite int) error {
	a, b := vFunction.V1, vFunction.V2
	aVersion, bVersion := 1, 2

	// randomly change execution order
	if rand.Intn(2) == 0 {
		a, b = vFunction.V2, vFunction.V1
		aVersion, bVersion = 2, 1
	}

	log.Printf("  |--> running[%d]: %s\n", aVersion, a.Directory)
	if err := RunFunction(csvWriter, a, aVersion, run, suite); err != nil {
		return err
	}

	log.Printf("  |--> running[%d]: %s", bVersion, b.Directory)
	if err := RunFunction(csvWriter, b, bVersion, run, suite); err != nil {
		return err
	}

	return nil
}

func RunSuite(csvWriter *csv.Writer, fns []VersionedFunction, run, suite int) error {
	newFns := make([]VersionedFunction, len(fns))
	copy(newFns, fns)

	// shuffle execution order
	rand.Shuffle(len(newFns), func(i, j int) {
		newFns[i], newFns[j] = newFns[j], newFns[i]
	})

	for _, function := range newFns {
		log.Printf("--| benchmarking: %s\n", function.String())
		err := RunVersionedFunction(csvWriter, function, run, suite)
		if err != nil {
			return err
		}
	}
	return nil
}

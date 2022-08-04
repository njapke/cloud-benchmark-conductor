package benchmark

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/christophwitzko/master-thesis/pkg/logger"
	"golang.org/x/perf/benchfmt"
)

const (
	Timeout        = 60
	ExecutionCount = 5
)

var (
	timeoutArg = fmt.Sprintf("-timeout=%ds", Timeout)
	countArg   = fmt.Sprintf("-count=%d", ExecutionCount)
)

func init() {
	rand.Seed(time.Now().Unix())
}

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
	"FileName",
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
		r.Function.FileName,
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

func RunFunction(ctx context.Context, log *logger.Logger, resultWriter ResultWriter, f Function, version, run, suite int) error {
	args := []string{
		"test",
		"-run=^$",
		"-benchmem",
		"-benchtime=1s",
		timeoutArg,
		countArg,
		fmt.Sprintf("-bench=^%s$", f.Name),
		// package path relative to the root directory
		"./" + filepath.Dir(f.FileName),
	}

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = f.RootDirectory
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	pipeRead, pipeWrite := io.Pipe()
	logPipeRead, logPipeWrite := io.Pipe()
	cmd.Stdout = pipeWrite
	cmd.Stderr = logPipeWrite

	benchFmtReader := io.TeeReader(pipeRead, logPipeWrite)
	go func() {
		logLineScanner := bufio.NewScanner(logPipeRead)
		for logLineScanner.Scan() {
			log.Infof("       | %s", logLineScanner.Text())
		}
	}()

	errCh := make(chan error, 1)
	go func() {
		if err := cmd.Run(); err != nil {
			errCh <- err
		}
		_ = pipeWrite.Close()
		_ = logPipeWrite.Close()
		close(errCh)
	}()

	i := 0
	bReader := benchfmt.NewReader(benchFmtReader, "bench.txt")
	for bReader.Scan() {
		switch rec := bReader.Result(); rec := rec.(type) {
		case *benchfmt.SyntaxError:
			log.Warnf("syntax error: %s", rec.Error())
			continue
		case *benchfmt.Result:
			res := NewResult(f, version, run, suite, i+1, rec)
			if err := resultWriter.Write(res); err != nil {
				return err
			}
			i++
		default:
			log.Warnf("unknown record type: %T", rec)
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

func RunVersionedFunction(ctx context.Context, log *logger.Logger, resultWriter ResultWriter, vFunction VersionedFunction, run, suite int) error {
	a, b := vFunction.V1, vFunction.V2
	aVersion, bVersion := 1, 2

	// randomly change execution order
	if rand.Intn(2) == 0 {
		a, b = vFunction.V2, vFunction.V1
		aVersion, bVersion = 2, 1
	}

	log.Infof("  |--> running[v%d]: %s", aVersion, a.FileName)
	if err := RunFunction(ctx, log, resultWriter, a, aVersion, run, suite); err != nil {
		return err
	}

	log.Infof("  |--> running[v%d]: %s", bVersion, b.FileName)
	if err := RunFunction(ctx, log, resultWriter, b, bVersion, run, suite); err != nil {
		return err
	}

	return nil
}

func RunSuite(ctx context.Context, log *logger.Logger, resultWriter ResultWriter, fns VersionedFunctions, run, suite int) error {
	newFns := make(VersionedFunctions, len(fns))
	copy(newFns, fns)

	// shuffle execution order
	rand.Shuffle(len(newFns), func(i, j int) {
		newFns[i], newFns[j] = newFns[j], newFns[i]
	})

	for _, function := range newFns {
		log.Infof("--| benchmarking: %s", function.String())
		err := RunVersionedFunction(ctx, log, resultWriter, function, run, suite)
		if err != nil {
			return err
		}
	}
	return nil
}

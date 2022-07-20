package benchmark

import (
	"fmt"
	"io"
	"log"
	"os/exec"

	"golang.org/x/perf/benchfmt"
)

const Timeout = 60

var timeoutArg = fmt.Sprintf("-timeout=%ds", Timeout)

type Result struct {
	Ops    float64
	Bytes  float64
	Allocs float64
}

func NewResult(b *benchfmt.Result) *Result {
	ops, _ := b.Value("sec/op")
	bytes, _ := b.Value("B/op")
	allocs, _ := b.Value("allocs/op")
	return &Result{
		Ops:    ops,
		Bytes:  bytes,
		Allocs: allocs,
	}
}

func RunFunction(f Function) (*Result, error) {
	args := []string{
		"test",
		"-run=^$",
		"-benchmem",
		"-benchtime=1s",
		timeoutArg,
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

	var res *Result
	bReader := benchfmt.NewReader(pipeRead, "bench.txt")
	for bReader.Scan() {
		switch rec := bReader.Result(); rec := rec.(type) {
		case *benchfmt.SyntaxError:
			log.Printf("syntax error: %s", rec.Error())
			continue
		case *benchfmt.Result:
			res = NewResult(rec)
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

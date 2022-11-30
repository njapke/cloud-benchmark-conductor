package min

import (
	"fmt"

	"bitbucket.org/sealuzh/gocg/pkg/cg"
	"bitbucket.org/sealuzh/gocg/pkg/overlap"
)

type Result struct {
	CG       *cg.Result
	OldCG    *cg.Result
	Selected []Selected // slice of selected benchmarks (keys and additional nodes) that correspond to CG.MicroCGs keys
}

func ApplyAll(projects []string, cgRes []*cg.Result, ovls []*overlap.System, strategy StrategyFunc, excl cg.ExclusionFunc) ([]*Result, error) {
	// check pre-conditions
	lcg := len(cgRes)
	lols := len(ovls)
	if lcg != lols {
		return nil, fmt.Errorf("cgRes and overlaps lengths unequal: %d != %d", lcg, lols)
	}

	var ret []*Result

	for i := 0; i < lcg; i++ {
		res, err := Apply(projects, cgRes[i], ovls[i], strategy, excl)
		if err != nil {
			return nil, fmt.Errorf("could not apply minimization for index %d: %w", i, err)
		}
		ret = append(ret, res)
	}

	return ret, nil
}

func Apply(projects []string, cgRes *cg.Result, ovls *overlap.System, strategy StrategyFunc, excl cg.ExclusionFunc) (*Result, error) {
	selected := strategy(cgRes, ovls, excl)

	copied := cgRes.Copy()
	copied.MicroCGs = make(cg.Map)

	for _, selected := range selected {
		bench := selected.Benchmark
		benchCG, ok := cgRes.MicroCGs[bench]
		if !ok {
			return nil, fmt.Errorf("could not get CG for bench '%s'", bench)
		}
		copied.MicroCGs[bench] = benchCG
	}

	return &Result{
		CG:       copied,
		OldCG:    cgRes,
		Selected: selected,
	}, nil
}

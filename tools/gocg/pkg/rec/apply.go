package rec

import (
	"fmt"

	"bitbucket.org/sealuzh/gocg/pkg/cg"
	"bitbucket.org/sealuzh/gocg/pkg/overlap"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/traverse"
)

type Result struct {
	OldCG                *cg.Result
	CG                   *cg.Result
	RecommendedFunctions []Function
	NrBenchs             int
}

func ApplyAll(projects []string, cgRes []*cg.Result, ovls []*overlap.System, strategy StrategyFunc, nrBenchs int) ([]*Result, error) {
	// check pre-conditions
	lcg := len(cgRes)
	lols := len(ovls)
	if lcg != lols {
		return nil, fmt.Errorf("cgRes and overlaps lengths unequal: %d != %d", lcg, lols)
	}

	var ret []*Result

	for i := 0; i < lcg; i++ {
		res, err := Apply(projects, cgRes[i], ovls[i], strategy, nrBenchs)
		if err != nil {
			return nil, fmt.Errorf("could not apply recoomendation for index %d: %w", i, err)
		}
		ret = append(ret, res)
	}

	return ret, nil
}

func Apply(projects []string, cgRes *cg.Result, ovls *overlap.System, strategy StrategyFunc, nrBenchs int) (*Result, error) {
	rfs := strategy(projects, cgRes, ovls, nrBenchs)
	newCG := cgRes.Copy()

	// add recommended functions
	addRecFuncs(newCG, rfs)

	return &Result{
		OldCG:                cgRes,
		CG:                   newCG,
		RecommendedFunctions: rfs,
		NrBenchs:             nrBenchs,
	}, nil
}

func addRecFuncs(cgRes *cg.Result, recFuncs []Function) {
	scg := cgRes.SystemCG

	for _, recFunc := range recFuncs {
		rf := recFunc.Function

		rfg := simple.NewWeightedDirectedGraph(-1, -2)

		// add recommended node
		rfg.AddNode(rf)

		// v := func(n graph.Node) {
		// 	// fmt.Println(rfg)
		// 	fmt.Printf("%v\n", n)
		// 	f := n.(*cg.Function)
		// 	rfg.AddNode(f)
		// }

		t := func(e graph.Edge) bool {
			c := e.(*cg.Call)
			rfg.SetWeightedEdge(c)
			return true
		}

		bfs := traverse.BreadthFirst{
			// Visit:    v,
			Traverse: t,
		}

		bfs.Walk(scg, rf, nil)

		microCGID := fmt.Sprintf("rec-bench_%d_%s__%d__%.5f__%.5f.dot", rf.ID(), rf.Name, cgRes.Config.NodeCount, cgRes.Config.NodeFraction, cgRes.Config.EdgeFraction)
		cgRes.MicroCGs[microCGID] = rfg
	}

}

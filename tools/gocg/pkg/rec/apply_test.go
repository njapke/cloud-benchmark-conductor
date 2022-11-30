package rec_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"bitbucket.org/sealuzh/gocg/pkg/cg"
	"bitbucket.org/sealuzh/gocg/pkg/overlap"
	"bitbucket.org/sealuzh/gocg/pkg/rec"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

func TestApplyNoRec(t *testing.T) {
	g, ider := newGraph(true, false, true)
	ovls := newOverlap(ider, map[int]map[int]struct{}{})

	cgRes := &cg.Result{
		IDer:     ider,
		System:   "TestSystem",
		MicroCGs: make(cg.Map),
		SystemCG: g,
	}

	nrBenchs := 2
	recBenchs := 0

	ar, err := rec.Apply(projectPrefixSlice, cgRes, ovls, recFunc(recBenchs, ider, g), nrBenchs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ar.CG == ar.OldCG {
		t.Fatalf("old and new cg.Result are the same")
	}

	if ar.NrBenchs != nrBenchs {
		t.Fatalf("expected %d nrBenchs but was %d", nrBenchs, ar.NrBenchs)
	}

	if l := len(ar.RecommendedFunctions); l != recBenchs {
		t.Fatalf("expected %d recommended benchs but was %d", recBenchs, l)
	}
}

func TestApplyThreeRecs(t *testing.T) {
	g, ider := newGraph(true, false, true)
	ovls := newOverlap(ider, map[int]map[int]struct{}{
		1: map[int]struct{}{
			3: struct{}{},
			6: struct{}{},
		},
	})

	cgRes := &cg.Result{
		IDer:     ider,
		System:   "TestSystem",
		MicroCGs: make(cg.Map),
		SystemCG: g,
	}

	nrBenchs := 3
	recBenchs := 3

	rff := recFunc(recBenchs, ider, g)
	recFuncSlice := rff([]string{""}, nil, nil, -1)

	ar, err := rec.Apply(projectPrefixSlice, cgRes, ovls, rff, nrBenchs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ar.CG == ar.OldCG {
		t.Fatalf("old and new cg.Result are the same")
	}

	if ar.NrBenchs != nrBenchs {
		t.Fatalf("expected %d nrBenchs but was %d", nrBenchs, ar.NrBenchs)
	}

	if l := len(ar.RecommendedFunctions); l != recBenchs {
		t.Fatalf("expected %d recommended benchs but was %d", recBenchs, l)
	}

	for i, rf := range ar.RecommendedFunctions {
		exp := *recFuncSlice[i].Function
		was := *rf.Function

		if was != exp {
			t.Fatalf("enexpected recommended function: expected %+v was %+v", exp, was)
		}
	}

	if lmcgs := len(ar.CG.MicroCGs); lmcgs != recBenchs {
		t.Fatalf("new microbenchmark CG count wrong: expected %d, was %d", recBenchs, lmcgs)
	}

	expCGs := recFuncCGs(g, ider, recBenchs)

	for key, mbcg := range ar.CG.MicroCGs {
		excg := cgForKey(expCGs, key)
		cgEqual(t, projectPrefix, excg, mbcg)
	}
}

// helper

func recFunc(acRecFuncs int, ider *cg.IDer, g graph.WeightedDirected) func(projects []string, callgraphs *cg.Result, overlaps *overlap.System, count int) []rec.Function {
	if acRecFuncs > 3 {
		panic(fmt.Sprintf("invalid number of recommended functions: between 0 and 3 required"))
	}

	n1 := g.Node(ider.ID(nodeName(1)(2, true))).(*cg.Function)
	n2 := g.Node(ider.ID(nodeName(2)(2, true))).(*cg.Function)
	n3 := g.Node(ider.ID(nodeName(2)(3, true))).(*cg.Function)

	rfs := []rec.Function{
		{
			Function:        n1,
			AdditionalNodes: 10, // not in line with CG
		},
		{
			Function:        n2,
			AdditionalNodes: 5, // not in line with CG
		},
		{
			Function:        n3,
			AdditionalNodes: 2, // not in line with CG
		},
	}

	return func(projects []string, callgraphs *cg.Result, overlaps *overlap.System, count int) []rec.Function {
		return rfs[0:acRecFuncs]
	}
}

func recFuncCGs(g graph.WeightedDirected, ider *cg.IDer, count int) []graph.WeightedDirected {
	gs := make([]graph.WeightedDirected, 0, count)
	et := time.Microsecond

	// f1_2
	g1 := simple.NewWeightedDirectedGraph(-1, -2)
	n12 := g.Node(ider.ID(nodeName(1)(2, true))).(*cg.Function)
	n14 := g.Node(ider.ID(nodeName(1)(4, true))).(*cg.Function)
	n15 := g.Node(ider.ID(nodeName(1)(5, true))).(*cg.Function)

	g1.AddNode(n12)
	g1.AddNode(n14)
	g1.AddNode(n15)

	g1.SetWeightedEdge(cg.NewCall(n12, n14, et))
	g1.SetWeightedEdge(cg.NewCall(n12, n15, et))
	g1.SetWeightedEdge(cg.NewCall(n14, n15, et))
	g1.SetWeightedEdge(cg.NewCall(n15, n12, et))

	gs = append(gs, g1)

	// f2_2
	g2 := simple.NewWeightedDirectedGraph(-1, -2)
	n22 := g.Node(ider.ID(nodeName(2)(2, true))).(*cg.Function)
	n24 := g.Node(ider.ID(nodeName(2)(4, false))).(*cg.Function)
	n25 := g.Node(ider.ID(nodeName(2)(5, true))).(*cg.Function)
	n26 := g.Node(ider.ID(nodeName(2)(6, true))).(*cg.Function)
	n28 := g.Node(ider.ID(nodeName(2)(8, false))).(*cg.Function)
	n29 := g.Node(ider.ID(nodeName(2)(9, false))).(*cg.Function)

	g2.AddNode(n22)
	g2.AddNode(n24)
	g2.AddNode(n25)
	g2.AddNode(n26)
	g2.AddNode(n28)
	g2.AddNode(n29)

	g2.SetWeightedEdge(cg.NewCall(n22, n24, et))
	g2.SetWeightedEdge(cg.NewCall(n22, n25, et))
	g2.SetWeightedEdge(cg.NewCall(n22, n26, et))
	g2.SetWeightedEdge(cg.NewCall(n24, n28, et))
	g2.SetWeightedEdge(cg.NewCall(n25, n29, et))
	g2.SetWeightedEdge(cg.NewCall(n26, n29, et))
	g2.SetWeightedEdge(cg.NewCall(n28, n24, et))
	g2.SetWeightedEdge(cg.NewCall(n29, n26, et))

	gs = append(gs, g2)

	// f2_3
	g3 := simple.NewWeightedDirectedGraph(-1, -2)
	n23 := g.Node(ider.ID(nodeName(2)(3, true))).(*cg.Function)
	n27 := g.Node(ider.ID(nodeName(2)(7, true))).(*cg.Function)
	n210 := g.Node(ider.ID(nodeName(2)(10, false))).(*cg.Function)
	n211 := g.Node(ider.ID(nodeName(2)(11, false))).(*cg.Function)

	g3.AddNode(n23)
	g3.AddNode(n27)
	g3.AddNode(n210)
	g3.AddNode(n211)

	g3.SetWeightedEdge(cg.NewCall(n23, n27, et))
	g3.SetWeightedEdge(cg.NewCall(n27, n210, et))
	g3.SetWeightedEdge(cg.NewCall(n27, n211, et))

	gs = append(gs, g3)

	return gs[:count]
}

func recOverlaps(g graph.WeightedDirected, ider *cg.IDer, recCount int) map[int64]struct{} {
	ret := make(map[int64]struct{})
	if recCount >= 1 {
		ret[ider.ID(nodeName(1)(2, true))] = struct{}{}
		ret[ider.ID(nodeName(1)(4, true))] = struct{}{}
		ret[ider.ID(nodeName(1)(5, true))] = struct{}{}
	}

	if recCount >= 2 {
		ret[ider.ID(nodeName(2)(2, true))] = struct{}{}
		ret[ider.ID(nodeName(2)(5, true))] = struct{}{}
		ret[ider.ID(nodeName(2)(6, true))] = struct{}{}
	}

	if recCount >= 3 {
		ret[ider.ID(nodeName(2)(3, true))] = struct{}{}
		ret[ider.ID(nodeName(2)(7, true))] = struct{}{}
	}

	return ret
}

func cgForKey(cgs []graph.WeightedDirected, key string) graph.WeightedDirected {
	var g graph.WeightedDirected

	switch {
	case strings.Contains(key, nodeName(1)(2, true)):
		g = cgs[0]
	case strings.Contains(key, nodeName(2)(2, true)):
		g = cgs[1]
	case strings.Contains(key, nodeName(2)(3, true)):
		g = cgs[2]
	default:
		panic(fmt.Sprintf("unknown key '%s'", key))
	}

	return g
}

func cgEqual(t *testing.T, project string, exp, was graph.WeightedDirected) {
	wasNodes := make(map[int64]struct{})

	ns := was.Nodes()
	for ns.Next() {
		f := ns.Node().(*cg.Function)
		if strings.HasPrefix(f.Name, project) {
			wasNodes[f.ID()] = struct{}{}
		}
	}

	ns = exp.Nodes()
	for ns.Next() {
		f := ns.Node().(*cg.Function)

		if !strings.HasPrefix(f.Name, project) {
			continue
		}

		if _, ok := wasNodes[f.ID()]; !ok {
			t.Fatalf(fmt.Sprintf("expected %v but did not get it", f))
		}
	}
}

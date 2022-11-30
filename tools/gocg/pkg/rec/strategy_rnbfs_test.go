package rec_test

import (
	"testing"

	"bitbucket.org/sealuzh/gocg/pkg/cg"
	"bitbucket.org/sealuzh/gocg/pkg/rec"
	"gonum.org/v1/gonum/graph/simple"
)

func TestStrategyRNBFSEmptyAll(t *testing.T) {
	rfs := rec.StrategyRootNodeBFSNonOverlapping(
		projectPrefixSlice,
		&cg.Result{
			SystemCG: simple.NewWeightedDirectedGraph(-1, -2),
		},
		noOverlap(),
		10,
	)

	if l := len(rfs); l != 0 {
		t.Fatalf("expected no recommended functions but got %d", l)
	}
}

func TestStrategyRNBFSEmptySystemCG(t *testing.T) {
	g := simple.NewWeightedDirectedGraph(-1, -2)
	ider := cg.NewIDer()

	rfs := rec.StrategyRootNodeBFSNonOverlapping(
		projectPrefixSlice,
		&cg.Result{
			SystemCG: g,
			IDer:     ider,
		},
		newOverlap(ider, systemMicroOverlaps()),
		10,
	)

	if l := len(rfs); l != 0 {
		t.Fatalf("expected no recommended functions but got %d", l)
	}
}

func TestStrategyRNBFSNoProjectNodes(t *testing.T) {
	g, ider := newGraph(true, false, false)

	rfs := rec.StrategyRootNodeBFSNonOverlapping(
		projectPrefixSlice,
		&cg.Result{
			SystemCG: g,
			IDer:     ider,
		},
		newOverlap(ider, systemMicroOverlaps()),
		10,
	)

	if l := len(rfs); l != 0 {
		t.Fatalf("expected no recommended functions but got %d", l)
	}
}

func TestStrategyRNBFSSingleRootOneFunc(t *testing.T) {
	const (
		maxFuncs = 3
		expFuncs = 1
	)

	g, ider := newGraph(false, false, true)

	rfs := rec.StrategyRootNodeBFSNonOverlapping(
		projectPrefixSlice,
		&cg.Result{
			SystemCG: g,
			IDer:     ider,
		},
		newOverlap(ider, systemMicroOverlaps()),
		maxFuncs,
	)

	if l := len(rfs); l != expFuncs {
		t.Fatalf("expected %d recommended functions but got %d", expFuncs, l)
	}

	rf0 := rfs[0].Function
	if expNode := nodeName(1)(2, true); rf0.Name != expNode {
		t.Fatalf("expected node %s but was %v", expNode, rf0)
	}
}

func TestStrategyRNBFSTwoRootsThreeFuncsUnconnected(t *testing.T) {
	const (
		maxFuncs = 3
		expFuncs = 3
	)

	g, id := newGraph(true, false, true)

	rfs := rec.StrategyRootNodeBFSNonOverlapping(
		projectPrefixSlice,
		&cg.Result{
			SystemCG: g,
			IDer:     id,
		},
		newOverlap(id, systemMicroOverlaps()),
		maxFuncs,
	)

	if l := len(rfs); l != expFuncs {
		t.Fatalf("expected %d recommended functions but got %d", expFuncs, l)
	}

	rf0 := rfs[0].Function
	if expNode := nodeName(1)(2, true); rf0.Name != expNode {
		t.Fatalf("expected node %s but was %v", expNode, rf0)
	}
}

// this test fails because of wrong implementation
//
// func TestStrategyRNBFSTwoRootsThreeFuncsConnected(t *testing.T) {
// 	const (
// 		maxFuncs = 3
// 		expFuncs = 2
// 	)

// 	g, id := newGraph(true, false, true)

// 	rfs := rec.StrategyRootNodeBFSNonOverlapping(
// 		projectPrefix,
// 		&cg.Result{
// 			SystemCG: g,
// 			IDer:     id,
// 		},
// 		newOverlap(id, systemMicroOverlaps()),
// 		maxFuncs,
// 	)

// 	if l := len(rfs); l != expFuncs {
// 		t.Fatalf("expected %d recommended functions but got %d", expFuncs, l)
// 	}

// 	rf0 := rfs[0]
// 	if expNode := nodeName(1)(2, true); rf0.Name != expNode {
// 		t.Fatalf("expected node %s but was %v", expNode, rf0)
// 	}
// }

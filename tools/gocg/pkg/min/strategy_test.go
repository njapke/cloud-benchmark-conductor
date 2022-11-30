package min_test

import (
	"fmt"
	"testing"

	"bitbucket.org/sealuzh/gocg/pkg/cg"
	"bitbucket.org/sealuzh/gocg/pkg/min"
	"bitbucket.org/sealuzh/gocg/pkg/overlap"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

func TestGreedyMicroEmpty(t *testing.T) {
	cgRes, ovs := emptyCGOverlapResult()
	selected := min.GreedyMicro(cgRes, ovs, excludeNothing)
	testGreedyEmpty(t, selected)
}

func TestGreedySystemEmpty(t *testing.T) {
	cgRes, ovs := emptyCGOverlapResult()
	selected := min.GreedySystem(cgRes, ovs, excludeNothing)
	testGreedyEmpty(t, selected)
}

func testGreedyEmpty(t *testing.T, mins []min.Selected) {
	expLen := 0
	if l := len(mins); l != expLen {
		t.Fatalf("expected %d micros but got %d", expLen, l)
	}
}

func TestGreedyMicro(t *testing.T) {
	cgRes, ovs := cgOverlapResult(greedyMicro)
	selected := min.GreedyMicro(cgRes, ovs, exclNotProject)
	testGreedy(t, selected)
}

func TestGreedySystem(t *testing.T) {
	cgRes, ovs := cgOverlapResult(greedySystem)
	selected := min.GreedySystem(cgRes, ovs, exclNotProject)
	testGreedy(t, selected)
}

func testGreedy(t *testing.T, mins []min.Selected) {
	// mins expected [t3 t4 t9 t1 t2] or [t3 t4 t9 t2 t1]
	exp1 := []string{
		benchName(3),
		benchName(4),
		benchName(9),
		benchName(1),
		benchName(2),
	}
	exp2 := []string{
		benchName(3),
		benchName(4),
		benchName(9),
		benchName(2),
		benchName(1),
	}

	expLen := 5
	if l := len(mins); l != expLen {
		t.Fatalf("expected %d micros but got %d", expLen, l)
	}

	for i, selected := range mins {
		b := selected.Benchmark
		e1 := exp1[i]
		e2 := exp2[i]
		if b != e1 && b != e2 {
			t.Fatalf("expected '%s' or '%s' at position %d but was '%s'", e1, e2, i, b)
		}
	}
}

// helpers

const projPrefix = "proj/"

func excludeNothing(n graph.Node, level int) bool {
	return false
}

var exclNotProject = cg.ExclusionNot(cg.IsProject(projPrefix))

func emptyCGOverlapResult() (*cg.Result, *overlap.System) {
	return nil, &overlap.System{
		Name:   "TestSystem",
		Micros: make(map[string]*overlap.NodeResult),
		Total:  nil, // not needed by min.StrategyFuncs
	}
}

type greedyType int

const (
	greedyMicro greedyType = iota
	greedySystem
)

// see Table 2 in Chen and Lau - A simulation study on some heuristics for test suite reduction (IST'98)
func cgOverlapResult(t greedyType) (*cg.Result, *overlap.System) {
	systemName := "TestSystem"

	ider := cg.NewIDer()

	t1Name, t1g, t1o := graphAndOverlaps(t, ider, systemName, 1, []int64{1, 3, 21, 22, 24})
	t2Name, t2g, t2o := graphAndOverlaps(t, ider, systemName, 2, []int64{2, 4, 21})
	t3Name, t3g, t3o := graphAndOverlaps(t, ider, systemName, 3, []int64{3, 4, 5, 6, 7, 8, 23, 26})
	t4Name, t4g, t4o := graphAndOverlaps(t, ider, systemName, 4, []int64{9, 10, 11, 12, 23})
	t5Name, t5g, t5o := graphAndOverlaps(t, ider, systemName, 5, []int64{5, 6, 7, 9, 10, 22, 27, 28})
	t6Name, t6g, t6o := graphAndOverlaps(t, ider, systemName, 6, []int64{11, 13, 14, 22, 23, 24, 25, 29})
	t7Name, t7g, t7o := graphAndOverlaps(t, ider, systemName, 7, []int64{12, 15, 21})
	t8Name, t8g, t8o := graphAndOverlaps(t, ider, systemName, 8, []int64{8, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30})
	t9Name, t9g, t9o := graphAndOverlaps(t, ider, systemName, 9, []int64{13, 14, 15, 25, 26})
	t10Name, t10g, t10o := graphAndOverlaps(t, ider, systemName, 10, []int64{5, 9, 13, 24, 26, 28})
	t11Name, t11g, t11o := graphAndOverlaps(t, ider, systemName, 11, []int64{6, 11, 21})
	t12Name, t12g, t12o := graphAndOverlaps(t, ider, systemName, 12, []int64{7, 10, 14, 25, 27})

	mcgs := map[string]graph.WeightedDirected{
		t1Name:  t1g,
		t2Name:  t2g,
		t3Name:  t3g,
		t4Name:  t4g,
		t5Name:  t5g,
		t6Name:  t6g,
		t7Name:  t7g,
		t8Name:  t8g,
		t9Name:  t9g,
		t10Name: t10g,
		t11Name: t11g,
		t12Name: t12g,
	}

	cgRes := &cg.Result{
		IDer:     ider,
		System:   systemName,
		MicroCGs: mcgs,
	}

	mos := map[string]*overlap.NodeResult{
		t1Name:  t1o,
		t2Name:  t2o,
		t3Name:  t3o,
		t4Name:  t4o,
		t5Name:  t5o,
		t6Name:  t6o,
		t7Name:  t7o,
		t8Name:  t8o,
		t9Name:  t9o,
		t10Name: t10o,
		t11Name: t11o,
		t12Name: t12o,
	}

	ovl := &overlap.System{
		Name:   systemName,
		Micros: mos,
		Total:  nil, // not needed by min.StrategyFuncs
	}

	return cgRes, ovl
}

func benchName(nr int) string {
	return fmt.Sprintf("%st%d", projPrefix, nr)
}

func funcName(nr int64) string {
	var fn string
	if nr <= 20 {
		fn = fmt.Sprintf("%sr%d", projPrefix, nr)
	} else {
		fn = fmt.Sprintf("r%d", nr)
	}
	return fn
}

func graphAndOverlaps(t greedyType, ider *cg.IDer, systemName string, microNumber int, overlaps []int64) (string, *simple.WeightedDirectedGraph, *overlap.NodeResult) {
	microName := benchName(microNumber)

	// create graph
	g := simple.NewWeightedDirectedGraph(-1, -2)
	g.AddNode(cg.NewFunction(ider, microName))

	ovs := make(map[int64]int)
	for _, ov := range overlaps {
		fn := funcName(ov)
		id := ider.ID(fn)
		ovs[id] = 1
		g.AddNode(cg.NewFunction(ider, fn))
	}

	// create overlaps
	nrs := &overlap.NodeResult{
		SystemName: systemName,
		MicroName:  microName,
	}

	switch t {
	case greedyMicro:
		nrs.MicroNodes = ovs
	case greedySystem:
		nrs.OverlappingNodes = ovs
	}

	return microName, g, nrs
}

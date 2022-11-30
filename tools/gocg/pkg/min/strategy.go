package min

import (
	"sort"
	"time"

	"bitbucket.org/sealuzh/gocg/internal/set"
	"bitbucket.org/sealuzh/gocg/pkg/cg"
	"bitbucket.org/sealuzh/gocg/pkg/overlap"
	"gonum.org/v1/gonum/graph"
)

// StrategyFunc defines a benchmark minimization strategy where the result is the minimized suite with elements in order of selection
// and the number of additional nodes for every element
type StrategyFunc func(cgRes *cg.Result, overlaps *overlap.System, excl cg.ExclusionFunc) []Selected

var _ StrategyFunc = GreedyMicro
var _ StrategyFunc = GreedySystem

type Selected struct {
	Benchmark       string
	AdditionalNodes int
	appBenchTime time.Duration
}

func GreedyMicro(cgRes *cg.Result, overlaps *overlap.System, excl cg.ExclusionFunc) []Selected {
	return internalGreedy(cgRes, overlaps, overlap.SelectMicroNodes, excl)
}

func GreedySystem(cgRes *cg.Result, overlaps *overlap.System, excl cg.ExclusionFunc) []Selected {
	return internalGreedy(cgRes, overlaps, overlap.SelectOverlappingNodes, excl)
}

func internalGreedy(cgRes *cg.Result, overlaps *overlap.System, sel overlap.NodesSelector, excl cg.ExclusionFunc) []Selected {
	remaining := greedyMicrosFrom(cgRes, overlaps, sel, excl)
	selected := make([]Selected, 0, len(remaining))
	sort.Sort(sort.Reverse(remaining))

	for !remaining.EmptyOrDone() {
		next := remaining[0]
		nrNodes := len(next.Nodes)

		selected = append(selected, Selected{
			Benchmark:       next.MicroKey,
			AdditionalNodes: nrNodes,
			appBenchTime: next.appBenchTime,
		})
		remaining = remaining[1:]
		remaining.RemoveNodes(next.Nodes)
		sort.Sort(sort.Reverse(remaining))
	}

	return selected
}

type greedyMicros []*greedyMicro

func greedyMicrosFrom(cgRes *cg.Result, overlaps *overlap.System, sel overlap.NodesSelector, excl cg.ExclusionFunc) greedyMicros {
	gms := make(greedyMicros, 0, len(overlaps.Micros))
	for key, ovl := range overlaps.Micros {
		g := cgRes.MicroCGs[key]
		nodes := ovl.Nodes(sel)
		gm := newGreedyMicro(g, cgRes.SystemCG, key, ovl.MicroName, nodes, excl)
		gms = append(gms, gm)
	}
	return gms
}

func (g greedyMicros) Len() int {
	return len(g)
}

func (g greedyMicros) Less(i, j int) bool {
	m1 := g[i]
	m2 := g[j]
	return len(m1.Nodes) < len(m2.Nodes)
}

func (g greedyMicros) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}

func (g greedyMicros) EmptyOrDone() bool {
	return len(g) == 0 || len(g[0].Nodes) == 0
}

func (g greedyMicros) RemoveNodes(nodes map[int64]struct{}) {
	for _, m := range g {
		m.Nodes = set.ComplementInt64(m.Nodes, nodes)
	}
}

type greedyMicro struct {
	MicroKey  string
	MicroName string
	Nodes     map[int64]struct{}
	appBenchTime time.Duration
}

func newGreedyMicro(g graph.Graph, appG graph.Graph, key, name string, nodes map[int64]int, excl cg.ExclusionFunc) *greedyMicro {
	gm := &greedyMicro{
		MicroKey:  key,
		MicroName: name,
		Nodes:     make(map[int64]struct{}),
		appBenchTime: 0.0,
	}

	for nid := range nodes {
		n := g.Node(nid).(*cg.Function) // should never panic
		if !excl(n, -1) {
			gm.Nodes[nid] = struct{}{}
		}

		appIterator := appG.Nodes()
		for appIterator.Next() {
			myAppNode := appIterator.Node()
			if myAppNode != nil {
				myAppFunction := myAppNode.(*cg.Function)
				if myAppFunction.Name == n.Name {
					gm.appBenchTime += myAppFunction.FunctionTime
				}
			}
		}
	}

	return gm
}

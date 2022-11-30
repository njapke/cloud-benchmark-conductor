package rec_test

import (
	"fmt"
	"time"

	"bitbucket.org/sealuzh/gocg/pkg/cg"
	"bitbucket.org/sealuzh/gocg/pkg/overlap"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

const projectPrefix = "project/"

var projectPrefixSlice []string = []string{projectPrefix}

// helper

func printGraph(g graph.Graph) {
	fmt.Println("-------------------")
	fmt.Println("Nodes:")

	nodes := g.Nodes()
	for nodes.Next() {
		n := nodes.Node()
		fmt.Println(n)
	}

	fmt.Println("-------------------")
	fmt.Println("Edges:")

	fmt.Println("-------------------")
}

func newGraph(twoRoots, connected, addProjectNodes bool) (graph.WeightedDirected, *cg.IDer) {
	g := simple.NewWeightedDirectedGraph(-1, -2)
	ider := cg.NewIDer()

	pns := map[int]map[int]struct{}{}

	if addProjectNodes {
		pns = projectNodes()
	}

	subgraphNodes(g, ider, 1, 7, pns[1], subgraph1Edges)

	// add second subgraph (with own root)
	if twoRoots {
		subgraphNodes(g, ider, 2, 12, pns[2], subgraph2Edges)
	}

	// add connecting edge(s) for subgraphs
	if connected {
		subgraphEdges(
			g,
			ider,
			[]edge{{nodeName(1)(5, true), nodeName(2)(2, true)}},
			time.Microsecond,
		)
	}

	return g, ider
}

func projectNodes() map[int]map[int]struct{} {
	pns := map[int]map[int]struct{}{}

	pns[1] = map[int]struct{}{
		2: struct{}{},
		3: struct{}{},
		4: struct{}{},
		5: struct{}{},
		6: struct{}{},
	}

	pns[2] = map[int]struct{}{
		2: struct{}{},
		3: struct{}{},
		5: struct{}{},
		6: struct{}{},
		7: struct{}{},
	}

	return pns
}

func systemMicroOverlaps() map[int]map[int]struct{} {
	ovs := map[int]map[int]struct{}{}

	ovs[1] = map[int]struct{}{
		3: struct{}{},
		6: struct{}{},
	}

	ovs[2] = map[int]struct{}{}

	return ovs
}

func nodeName(subgraph int) func(nodeNumber int, isProject bool) string {
	return func(nodeNumber int, isProject bool) string {
		var pp string
		if isProject {
			pp = projectPrefix
		}
		return fmt.Sprintf("%sf%d_%d", pp, subgraph, nodeNumber)
	}
}

func subgraphNodes(g *simple.WeightedDirectedGraph, ider *cg.IDer, subgraphID int, nrFuncs int, projectNodes map[int]struct{}, edgesFunc func(*simple.WeightedDirectedGraph, *cg.IDer, []*cg.Function, time.Duration)) {
	et := time.Microsecond

	nn := nodeName(subgraphID)

	fs := make([]*cg.Function, nrFuncs)
	for i := 0; i < nrFuncs; i++ {
		_, isProjectNode := projectNodes[i]
		fs[i] = cg.NewFunction(ider, nn(i, isProjectNode))
		g.AddNode(fs[i])
	}

	edgesFunc(g, ider, fs, et)
}

func subgraph1Edges(g *simple.WeightedDirectedGraph, ider *cg.IDer, fs []*cg.Function, et time.Duration) {
	edges := []edge{
		{fs[0].Name, fs[1].Name},
		{fs[0].Name, fs[2].Name},
		{fs[1].Name, fs[2].Name},
		{fs[1].Name, fs[3].Name},
		{fs[1].Name, fs[4].Name},
		{fs[2].Name, fs[4].Name},
		{fs[2].Name, fs[5].Name},
		{fs[3].Name, fs[1].Name},
		{fs[3].Name, fs[6].Name},
		{fs[4].Name, fs[5].Name},
		{fs[5].Name, fs[2].Name},
	}

	subgraphEdges(g, ider, edges, et)
}

func subgraph2Edges(g *simple.WeightedDirectedGraph, ider *cg.IDer, fs []*cg.Function, et time.Duration) {
	edges := []edge{
		{fs[0].Name, fs[1].Name},
		{fs[1].Name, fs[2].Name},
		{fs[1].Name, fs[3].Name},
		{fs[1].Name, fs[4].Name},
		{fs[2].Name, fs[4].Name},
		{fs[2].Name, fs[5].Name},
		{fs[2].Name, fs[6].Name},
		{fs[3].Name, fs[7].Name},
		{fs[4].Name, fs[8].Name},
		{fs[3].Name, fs[7].Name},
		{fs[5].Name, fs[9].Name},
		{fs[6].Name, fs[9].Name},
		{fs[7].Name, fs[10].Name},
		{fs[7].Name, fs[11].Name},
		{fs[9].Name, fs[6].Name},
	}

	subgraphEdges(g, ider, edges, et)
}

type edge struct {
	From string
	To   string
}

func subgraphEdges(g *simple.WeightedDirectedGraph, ider *cg.IDer, edges []edge, et time.Duration) {
	for _, e := range edges {
		fromID := ider.ID(e.From)
		toID := ider.ID(e.To)
		fromNode := g.Node(fromID).(*cg.Function)
		toNode := g.Node(toID).(*cg.Function)
		e := cg.NewCall(fromNode, toNode, et)
		g.SetWeightedEdge(e)
	}
}

func noOverlap() *overlap.System {
	return &overlap.System{
		Total: &overlap.NodeResult{
			OverlappingNodes: make(map[int64]int),
		},
	}
}

func newOverlap(ider *cg.IDer, overlaps map[int]map[int]struct{}) *overlap.System {
	ons := make(map[int64]int)
	for subgraph, ovs := range overlaps {
		for on := range ovs {
			nn := nodeName(subgraph)(on, true)
			id := ider.ID(nn)
			ons[id] = 1
		}
	}

	return &overlap.System{
		Total: &overlap.NodeResult{
			SystemName:       "system1",
			MicroName:        "all",
			OverlappingNodes: ons,
		},
	}
}

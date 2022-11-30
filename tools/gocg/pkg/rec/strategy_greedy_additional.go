package rec

import (
	"fmt"
	"reflect"
	"sort"

	"bitbucket.org/sealuzh/gocg/internal/set"
	"bitbucket.org/sealuzh/gocg/pkg/cg"
	"bitbucket.org/sealuzh/gocg/pkg/overlap"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/traverse"
)

var _ StrategyFunc = StrategyGreedyAdditional

func StrategyGreedyAdditional(projects []string, callgraphs *cg.Result, overlaps *overlap.System, count int) []Function {
	return strategyGreedy(projects, callgraphs, overlaps, count, []reachabilitiesUpdateFunc{ruFilterCoveredByNewFunc, ruAdditional})
}

func strategyGreedy(projects []string, callgraphs *cg.Result, overlaps *overlap.System, count int, rus []reachabilitiesUpdateFunc) []Function {
	scg := callgraphs.SystemCG
	recFuncs := make([]Function, 0, count)

	isProj := cg.IsProjects(projects)
	isOvl := overlap.IsOverlapping(overlaps)

	rs := reachabilitiesAll(projects, scg, isProj, isOvl)

	exclude := cg.ExclusionOr(
		cg.ExclusionNot(isProj),
		isOvl,
	)

	rs = addLevelFromRoot(scg, rs, exclude)

	if len(rs) == 0 {
		return recFuncs
	}

	for len(recFuncs) < count && len(rs) > 0 {
		// sort functions according to their reachabilities count
		sort.Sort(sort.Reverse(rs))

		// recommended function
		rf := rs[0]
		// remove recommended function
		rs = rs[1:]

		additionalNodes := len(rf.ReachableProject)
		// do not pick a method that does not cover new parts
		if additionalNodes == 0 {
			continue
		}

		rff := rf.Function
		recFuncs = append(recFuncs, Function{
			Function:        rff,
			AdditionalNodes: additionalNodes,
		})
		// remaining functions
		rs = reachabilitiesUpdate(rs, rf, rus)
	}

	return recFuncs
}

func addLevelFromRoot(g graph.Directed, frs []*functionReachabilities, excluded cg.ExclusionFunc) []*functionReachabilities {
	rootNodes := cg.RootNodes(g)
	nodeToLevel := cg.Levels(g, rootNodes, excluded).NodeToLevel

	for _, fr := range frs {
		id := fr.Function.ID()
		level, ok := nodeToLevel[id]
		if !ok {
			panic(fmt.Sprintf("could not retrieve level for node %v", fr.Function))
		}
		fr.Level = level
	}

	return frs
}

type functionReachabilitiesSlice []*functionReachabilities

func (s functionReachabilitiesSlice) Len() int {
	return len(s)
}

// Less assesses whether an element in functionReachabilitiesSlice (in this order)
// (1) has fewer called project nodes,
// (2) is at a deeper callgraph level,
// (3) has fewer called non-project nodes
func (s functionReachabilitiesSlice) Less(i, j int) bool {
	n1 := s[i]
	n2 := s[j]

	// project reachable sizes (higher is better)
	ln1pr := len(n1.ReachableProject)
	ln2pr := len(n2.ReachableProject)
	switch {
	case ln1pr < ln2pr:
		return true
	case ln1pr > ln2pr:
		return false
	}

	// different levels (lower is better)
	switch {
	case n1.Level < n2.Level:
		return false
	case n1.Level > n2.Level:
		return true
	}

	// non-project reachable sizes (higher is better)
	ln1npr := len(n1.ReachableOther)
	ln2npr := len(n2.ReachableOther)
	switch {
	case ln1npr < ln2npr:
		return true
	case ln1npr > ln2npr:
		return false
	}

	// sort alphabetically if all other metrics are equal
	return n1.Function.Name < n2.Function.Name
}

func (s functionReachabilitiesSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type functionReachabilities struct {
	Function         *cg.Function
	Level            int
	ReachableProject set.Int64 // key is the node ID
	ReachableOther   set.Int64 // key is the node ID
}

func (fr *functionReachabilities) String() string {
	return fmt.Sprintf("functionReachabilities(Function=%v,Level=%d,ReachableProject=%v,ReachableOther=%v)", fr.Function, fr.Level, fr.ReachableProject, fr.ReachableOther)
}

func reachabilitiesAll(projects []string, g graph.WeightedDirected, isProject cg.ExclusionFunc, excluded cg.ExclusionFunc) functionReachabilitiesSlice {
	var frs []*functionReachabilities
	nodes := g.Nodes()
	for nodes.Next() {
		n := nodes.Node()

		f := n.(*cg.Function) // panics if not a *cg.Function -> which should never happen

		// filter excluded and non-project nodes
		if excluded(f, -1) || !isProject(f, -1) {
			continue
		}

		fr := reachabilitiesFunc(g, isProject, excluded, f)
		frs = append(frs, fr)
	}
	return frs
}

func reachabilitiesFunc(g graph.WeightedDirected, isProject cg.ExclusionFunc, excluded cg.ExclusionFunc, f *cg.Function) *functionReachabilities {
	fr := &functionReachabilities{
		Function:         f,
		ReachableProject: make(set.Int64),
		ReachableOther:   make(set.Int64),
	}

	u := func(n graph.Node, d int) bool {
		func() {
			if _, ok := n.(*cg.Function); ok {
				// do not add excluded functions
				if excluded(n, d) {
					return
				}
				var m set.Int64
				// check if a non-project function
				if isProject(n, d) {
					m = fr.ReachableProject
				} else {
					m = fr.ReachableOther
				}
				m[n.ID()] = struct{}{}
			} else {
				panic(fmt.Sprintf("graph node not of type *cg.Function, was %v", reflect.TypeOf(n)))
			}
		}()

		return false
	}

	bfs := traverse.BreadthFirst{}
	bfs.Walk(g, f, u)
	return fr
}

type reachabilitiesUpdateFunc func(f, newFunc *functionReachabilities) *functionReachabilities

func reachabilitiesUpdate(rs []*functionReachabilities, newFunc *functionReachabilities, updateFuncs []reachabilitiesUpdateFunc) functionReachabilitiesSlice {
	var nrs []*functionReachabilities
	for _, r := range rs {
		var nr *functionReachabilities
		for _, f := range updateFuncs {
			nr = f(r, newFunc)
		}

		if nr != nil {
			nrs = append(nrs, nr)
		}
	}
	return nrs
}

func ruFilterCoveredByNewFunc(r, nf *functionReachabilities) *functionReachabilities {
	if r == nil {
		return nil
	}
	// is covered by new function
	_, coveredByNewFunc := nf.ReachableProject[r.Function.ID()]
	if coveredByNewFunc {
		return nil
	}
	return r
}

func ruAdditional(r, nf *functionReachabilities) *functionReachabilities {
	if r == nil {
		return nil
	}

	nfSet := set.Int64{
		nf.Function.ID(): struct{}{},
	}

	return &functionReachabilities{
		Function:         r.Function,
		Level:            r.Level,
		ReachableProject: set.ComplementInt64(set.ComplementInt64(r.ReachableProject, nf.ReachableProject), nfSet),
		ReachableOther:   set.ComplementInt64(r.ReachableOther, nf.ReachableOther),
	}
}

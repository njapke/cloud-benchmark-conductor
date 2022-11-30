package rec

import (
	"fmt"

	"bitbucket.org/sealuzh/gocg/pkg/cg"
	"bitbucket.org/sealuzh/gocg/pkg/overlap"
	"gonum.org/v1/gonum/graph"
)

var _ StrategyFunc = StrategyRootNodeBFSNonOverlapping

// StrategyRootNodeBFSNonOverlapping is the strategy that returns (at most) `count` functions that should be benchmarked, based on whether they have not been covered (are non-overlapping).
// It uses a breadth-first search to find these functions.
func StrategyRootNodeBFSNonOverlapping(projects []string, callgraphs *cg.Result, overlaps *overlap.System, count int) []Function {
	scg := callgraphs.SystemCG
	rns := cg.RootNodes(scg)

	exclude := cg.ExclusionOr(
		cg.ExclusionNot(cg.IsProjects(projects)),
		overlap.IsOverlapping(overlaps),
	)

	levelNodes := cg.Levels(scg, rns, exclude).Slice

	picked := pick(scg, levelNodes, 3)
	functions := make([]Function, len(picked))
	for i, p := range picked {
		functions[i] = Function{
			Function:        p,
			AdditionalNodes: -1,
		}
	}

	return functions
}

func pick(g graph.WeightedDirected, levelNodes [][]graph.Node, count int) []*cg.Function {
	// pick functions until count is reached where callers have not already been picked
	// start from first level and go levels down
	var picked []*cg.Function

	pickMutuallyExclusive := func(current []*cg.Function, candidate *cg.Function) []*cg.Function {
		// check if candidate has picked
		ok := inCG(g, current, candidate)
		if ok {
			return current
		}

		newCurrent := append(current, candidate)
		// lenNewCurrent := len(newCurrent)
		// if lenNewCurrent == 1 {
		// 	return newCurrent
		// }

		return newCurrent

		// TODO: check mutual exclusivity
		// for i, currentNode := range current[:lenNewCurrent-1] {
		// 	altCurrent := append(current[:i], current[i+1:]...)
		// 	icg := inCG(g, altCurrent, currentNode)
		// 	if icg {
		// 	}
		// }
	}

PickLoop:
	for _, nodes := range levelNodes {
		for _, node := range nodes {
			f := node.(*cg.Function) // panics if not *cg.Function -> should never be the case
			picked = pickMutuallyExclusive(picked, f)
			if pickedLen := len(picked); pickedLen == count {
				break PickLoop
			}
		}
	}

	return picked
}

func inCG(g graph.Directed, picked []*cg.Function, node *cg.Function) bool {
	if len(picked) == 0 {
		return false
	}

	pickedSet := map[int64]struct{}{}
	for _, p := range picked {
		id := p.ID()
		// check precondition that picked does not contain the same function/node twice
		if _, contained := pickedSet[id]; contained {
			panic(fmt.Sprintf("picked slice contains a function/node multiple times: %v", p))
		}
		pickedSet[id] = struct{}{}
	}

	var currentNode graph.Node = node
	seen := map[int64]struct{}{
		node.ID(): struct{}{},
	}

	addIfNotInCG := func(q []*cg.Function, callers graph.Nodes) (newQueue []*cg.Function, inCG bool) {
		newQueue = q
		for callers.Next() {
			n := callers.Node()
			f := n.(*cg.Function) // panics if a node of another typoe was returned -> should not happen

			id := f.ID()

			// check if already picked node is a caller
			if _, alreadyPicked := pickedSet[id]; alreadyPicked {
				return nil, true
			}

			// add caller to queue
			if _, alreadySeen := seen[id]; !alreadySeen {
				newQueue = append(newQueue, f)
				seen[id] = struct{}{}
			}
		}

		return newQueue, false
	}

	callerQueue, inCG := addIfNotInCG(make([]*cg.Function, 0, 10), g.To(currentNode.ID()))
	if inCG {
		return true
	}

	for len(callerQueue) > 0 {
		// get next node
		currentNode = callerQueue[0]
		// remove node from queue
		callerQueue = callerQueue[1:]

		callers := g.To(currentNode.ID())

		callerQueue, inCG = addIfNotInCG(callerQueue, callers)
		if inCG {
			return true
		}
	}

	return false
}

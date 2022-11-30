package overlap

import "fmt"

type System struct {
	Name   string
	Micros map[string]*NodeResult
	Total  *NodeResult
}

type NodeResult struct {
	SystemName       string
	MicroName        string
	SystemNodes      map[int64]int
	MicroNodes       map[int64]int
	OverlappingNodes map[int64]int
	OverlappingPerc  float64
}

func (r *NodeResult) CalcPerc() {
	nrSysNodes := len(r.SystemNodes)
	if nrSysNodes == 0 {
		r.OverlappingPerc = 0
	} else {
		r.OverlappingPerc = float64(len(r.OverlappingNodes)) / float64(len(r.SystemNodes))
	}
}

func (r *NodeResult) AddResult(other *NodeResult) {
	r.SystemNodes = mergeSets(r.SystemNodes, other.SystemNodes)
	r.MicroNodes = mergeSets(r.MicroNodes, other.MicroNodes)
	r.OverlappingNodes = mergeSets(r.OverlappingNodes, other.OverlappingNodes)
}

type NodesSelector int

const (
	SelectSystemNodes NodesSelector = iota
	SelectMicroNodes
	SelectOverlappingNodes
)

func (r *NodeResult) Nodes(sel NodesSelector) map[int64]int {
	var nodes map[int64]int

	switch sel {
	case SelectSystemNodes:
		nodes = r.SystemNodes
	case SelectMicroNodes:
		nodes = r.MicroNodes
	case SelectOverlappingNodes:
		nodes = r.OverlappingNodes
	default:
		panic(fmt.Sprintf("invalid NodesSelector %v", sel))
	}

	return nodes
}

func mergeSets(s1, s2 map[int64]int) map[int64]int {
	out := make(map[int64]int)
	for s1ID, count := range s1 {
		out[s1ID] = count
	}

	for s2ID, count := range s2 {
		prevCount, ok := out[s2ID]
		if ok {
			out[s2ID] = prevCount + count
		} else {
			out[s2ID] = count
		}
	}

	return out
}

package cg

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/traverse"
)

// rootNodes returns all root nodes in g.
func RootNodes(g graph.Directed) []graph.Node {
	var ret []graph.Node
	nodes := g.Nodes()
	for nodes.Next() {
		n := nodes.Node()
		callers := g.To(n.ID())
		if callers.Len() == 0 {
			// no callers -> must be a root node
			ret = append(ret, n)
		}
	}
	return ret
}

type ExclusionFunc func(n graph.Node, level int) bool

func ExclusionAnd(f1, f2 ExclusionFunc) ExclusionFunc {
	return func(n graph.Node, level int) bool {
		return f1(n, level) && f2(n, level)
	}
}

func ExclusionOr(f1, f2 ExclusionFunc) ExclusionFunc {
	return func(n graph.Node, level int) bool {
		return f1(n, level) || f2(n, level)
	}
}

func ExclusionNot(f ExclusionFunc) ExclusionFunc {
	return func(n graph.Node, level int) bool {
		return !f(n, level)
	}
}

type NodeLevels struct {
	Slice        [][]graph.Node
	LevelToNodes map[int][]graph.Node
	NodeToLevel  map[int64]int // key is node ID
}

func Levels(g graph.Directed, froms []graph.Node, exclude ExclusionFunc) *NodeLevels {
	ret := &NodeLevels{
		Slice:        [][]graph.Node{},
		LevelToNodes: make(map[int][]graph.Node),
		NodeToLevel:  make(map[int64]int),
	}

	if len(froms) == 0 {
		return ret
	}

	addLevel := func(level int) {
		l := len(ret.Slice)
		for i := l; i <= level; i++ {
			ret.Slice = append(ret.Slice, []graph.Node{})
			ret.LevelToNodes[i] = []graph.Node{}
		}
	}

	add := func(n graph.Node, level int) {
		ret.Slice[level] = append(ret.Slice[level], n)
		ret.LevelToNodes[level] = append(ret.LevelToNodes[level], []graph.Node{n}...)
		previousLevel, exists := ret.NodeToLevel[n.ID()]
		if !exists || exists && previousLevel > level {
			ret.NodeToLevel[n.ID()] = level
		}
	}

	u := func(n graph.Node, d int) bool {
		addLevel(d)
		if l := len(ret.Slice); d >= l {
			panic("should never happen because of call to addLevel")
		}

		if !exclude(n, d) {
			add(n, d)
		}

		return false
	}

	for _, from := range froms {
		bfs := traverse.BreadthFirst{}
		bfs.Walk(g, from, u)
	}

	return ret
}

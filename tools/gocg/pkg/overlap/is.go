package overlap

import (
	"bitbucket.org/sealuzh/gocg/pkg/cg"
	"gonum.org/v1/gonum/graph"
)

func IsOverlapping(overlaps *System) cg.ExclusionFunc {
	return func(n graph.Node, level int) bool {
		_, overlapping := overlaps.Total.OverlappingNodes[n.ID()]
		return overlapping
	}
}

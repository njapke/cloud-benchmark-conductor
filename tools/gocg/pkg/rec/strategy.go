package rec

import (
	"bitbucket.org/sealuzh/gocg/pkg/cg"
	"bitbucket.org/sealuzh/gocg/pkg/overlap"
)

// StrategyFunc functions return recommonded functions to be benchmarked based on a CG Result and an overlap.
// The same function as Strategy.Recommend.
type StrategyFunc func(projects []string, callgraphs *cg.Result, overlaps *overlap.System, count int) []Function

type Function struct {
	Function        *cg.Function
	AdditionalNodes int
}

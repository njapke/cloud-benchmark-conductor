package overlap

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"bitbucket.org/sealuzh/gocg/pkg/cg"
	"bitbucket.org/sealuzh/gocg/pkg/dir"
	"bitbucket.org/sealuzh/gocg/pkg/profile"
	"gonum.org/v1/gonum/graph"
)

const (
	benchNameAll = "ALL"
)

type overlapType int

func (t overlapType) String() string {
	var s string
	switch t {
	case typeNode:
		s = "node"
	case typeEdge:
		s = "edge"
	default:
		panic(fmt.Sprintf("invalid type %d", t))
	}
	return s
}

const (
	typeNode overlapType = iota
	typeEdge
)

func StructuralsWrite(projects []string, cgRes []*cg.Result, projectOnlyNodes bool, out io.Writer, writeHeader bool) error {
	systemOverlap, err := Structurals(projects, cgRes, projectOnlyNodes)
	if err != nil {
		return fmt.Errorf("could not get structural overlaps: %w", err)
	}

	for i, cgr := range cgRes {
		actualWriteHeader := writeHeader && i == 0

		system := cgr.System
		overlap := systemOverlap[i]

		err := structuralWrite(projects, system, overlap, projectOnlyNodes, out, actualWriteHeader, &cgr.Config)
		if err != nil {
			return fmt.Errorf("could not get structural overlap of config %+v: %w", cgr.Config, err)
		}
	}

	return nil
}

func structuralWrite(projects []string, system string, overlap *System, projectOnlyNodes bool, out io.Writer, writeHeader bool, outConfig *profile.OutConfig) error {
	w, err := csvWriter(out, writeHeader)
	if err != nil {
		return err
	}

	for _, mo := range overlap.Micros {
		err = csvWriteMicro(w, true, projects, projectOnlyNodes, typeNode, mo)
		if err != nil {
			return fmt.Errorf("could not write CSV: %w", err)
		}
	}

	err = csvWrite(w, true, projects, overlap.Total.SystemName, overlap.Total.MicroName, outConfig, projectOnlyNodes, typeNode, overlap.Total)
	if err != nil {
		return fmt.Errorf("could not write total overlap to CSV: %w", err)
	}

	w.Flush()

	return nil
}

// Structurals computes structural overlaps for all CG results (cgRes)
// cgRes and the returned slice are of same length, and the cgRes at index i corresponds to the overlap at index i
func Structurals(projects []string, cgRes []*cg.Result, projectOnlyNodes bool) ([]*System, error) {
	ret := make([]*System, 0, 10)
	for _, cgr := range cgRes {
		o, err := Structural(projects, cgr, projectOnlyNodes)
		if err != nil {
			return nil, fmt.Errorf("could not get structural overlap of config %+v: %w", cgr.Config, err)
		}

		// ensure system consistency
		if cgr.System != o.Name {
			panic(fmt.Sprintf("System names do not match: %s != %s", cgr.System, o.Total.SystemName))
		}

		ret = append(ret, o)
	}

	// ensure order consistency
	lCG := len(cgRes)
	lRet := len(ret)
	if lCG != lRet {
		panic(fmt.Sprintf("lengths of cgRes and overlaps not equal: %d != %d", lCG, lRet))
	}

	return ret, nil
}

// Structural returns the overlap of a systme benchmark (cgRes) with all microbenchmarks
func Structural(projects []string, cgRes *cg.Result, projectOnlyNodes bool) (*System, error) {
	return systemBenchNode(projects, cgRes.System, cgRes.SystemCG, cgRes.MicroCGs, projectOnlyNodes)
}

func systemBenchNode(projects []string, system string, systemCG graph.WeightedDirected, microCGs cg.Map, projectOnly bool) (*System, error) {
	total := &NodeResult{
		SystemName:       system,
		MicroName:        benchNameAll,
		SystemNodes:      make(map[int64]int),
		MicroNodes:       make(map[int64]int),
		OverlappingNodes: make(map[int64]int),
	}

	microOverlaps := make(map[string]*NodeResult)

	for micro, microCG := range microCGs {
		nodeOverlapResult, err := systemMicroBenchNode(projects, system, micro, systemCG, microCG, projectOnly)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not get overlaps for microbenchmark file '%s': %v\n", micro, err)
		}
		total.AddResult(nodeOverlapResult)
		microOverlaps[micro] = nodeOverlapResult
	}

	total.CalcPerc()

	return &System{
		Name:   system,
		Micros: microOverlaps,
		Total:  total,
	}, nil
}

func systemMicroBenchNode(projects []string, system, micro string, systemCG, microCG graph.WeightedDirected, projectOnly bool) (*NodeResult, error) {
	systemNodes := validNodes(projects, projectOnly, systemCG)
	microNodes := validNodes(projects, projectOnly, microCG)

	overlappingNodes := make(map[int64]int)
	for systemID := range systemNodes {
		if _, matching := microNodes[systemID]; matching {
			overlappingNodes[systemID] = overlappingNodes[systemID] + 1
		}
	}

	res := &NodeResult{
		SystemName:       system,
		MicroName:        micro,
		SystemNodes:      systemNodes,
		MicroNodes:       microNodes,
		OverlappingNodes: overlappingNodes,
	}
	res.CalcPerc()

	return res, nil
}

func validNodes(projects []string, projectOnly bool, graph graph.WeightedDirected) map[int64]int {
	nodes := make(map[int64]int)
	allNodes := graph.Nodes()

NodesLoop:
	for allNodes.Next() {
		genNode := allNodes.Node()
		node, ok := genNode.(*cg.Function)
		if !ok {
			panic("node not of type *cg.Function")
		}

		if projectOnly {
			// only add nodes with project prefix
			var projectNode bool
		ProjectLoop:
			for _, project := range projects {
				if strings.HasPrefix(node.Name, project) {
					projectNode = true
					break ProjectLoop
				}
			}

			if !projectNode {
				continue NodesLoop
			}
		}
		nodes[node.ID()] = 1
	}
	return nodes
}

func csvWriteMicro(w *csv.Writer, flush bool, projects []string, projectOnly bool, overlapType overlapType, res *NodeResult) error {
	system := res.SystemName
	micro := res.MicroName
	microBench, config, err := dir.ParseFileNameConfig(micro)
	if err != nil {
		return err
	}

	return csvWrite(w, flush, projects, system, microBench, config, projectOnly, overlapType, res)
}

func csvWrite(w *csv.Writer, flush bool, projects []string, system, micro string, config *profile.OutConfig, projectOnly bool, overlapType overlapType, res *NodeResult) error {
	err := w.Write([]string{
		strings.Join(projects, ","),
		system,
		micro,
		strconv.Itoa(config.NodeCount),
		fmt.Sprintf("%.5f", config.NodeFraction),
		strconv.FormatBool(projectOnly),
		strconv.Itoa(len(res.SystemNodes)),
		strconv.Itoa(len(res.MicroNodes)),
		overlapType.String(),
		strconv.Itoa(len(res.OverlappingNodes)),
		fmt.Sprintf("%.5f", res.OverlappingPerc),
	})
	if err != nil {
		return err
	}
	w.Flush()
	return nil
}

func csvWriter(out io.Writer, writeHeader bool) (*csv.Writer, error) {
	w := csv.NewWriter(out)
	if w == nil {
		return nil, fmt.Errorf("could not get CSV writer")
	}
	w.Comma = ';'

	if writeHeader {
		err := w.Write([]string{"project", "system", "micro", "config_node_count", "config_node_fraction", "project_only", "system_nodes", "micro_nodes", "overlap_type", "overlap_nodes", "overlap_perc"})
		if err != nil {
			return nil, fmt.Errorf("could not write header: %v", err)
		}
		w.Flush()
	}

	return w, nil
}

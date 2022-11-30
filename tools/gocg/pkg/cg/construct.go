package cg

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bitbucket.org/sealuzh/gocg/pkg/dir"
	"bitbucket.org/sealuzh/gocg/pkg/profile"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

// Map maps from filenames to callgraphs
type Map = map[string]graph.WeightedDirected

// Result contains the Map and the IDer
type Result struct {
	Config   profile.OutConfig
	System   string
	SystemCG graph.WeightedDirected
	MicroCGs Map
	IDer     *IDer
}

func FromDotsSystemMicro(systemDir, microDir string) ([]*Result, error) {
	results := make([]*Result, 0, 10)

	systemDirFile, err := os.Open(systemDir)
	if err != nil {
		return nil, fmt.Errorf("could not open system dir: %v", err)
	}
	defer systemDirFile.Close()

	systemFileInfos, err := systemDirFile.Readdir(-1)
	if err != nil {
		return nil, fmt.Errorf("could not read system file infos: %v", err)
	}

	for _, systemFileInfo := range systemFileInfos {
		if systemFileInfo.IsDir() {
			continue
		}

		sfn := systemFileInfo.Name()
		if !strings.HasSuffix(sfn, ".dot") {
			continue
		}

		_, systemFileConfig, err := dir.ParseFileNameConfig(sfn)
		if err != nil {
			return nil, fmt.Errorf("could not get system filename/config: %v", err)
		}

		res, err := FromDotsSystemMicroConfig(systemDir, systemFileInfo, microDir, systemFileConfig)
		if err != nil {
			return nil, fmt.Errorf("could not get call graphs for config %+v: %v", *systemFileConfig, err)
		}
		results = append(results, res)
	}

	return results, nil
}

func FromDotsSystemMicroConfig(systemDir string, systemFileInfo os.FileInfo, microDir string, config *profile.OutConfig) (*Result, error) {
	ider := NewIDer()

	system, c, err := dir.ParseFileNameConfig(systemFileInfo.Name())
	if err != nil {
		return nil, fmt.Errorf("could not get system filename/config: %v", err)
	}
	if *config != *c {
		return nil, fmt.Errorf("configs do not match: %+v != %+v", *config, *c)
	}

	systemCG, err := FromDot(systemDir, systemFileInfo, ider)
	if err != nil {
		return nil, fmt.Errorf("could not get system call graph: %v", err)
	}

	microCGs, _, err := FromDotsConfig([]string{microDir}, ider, *config)
	if err != nil {
		return nil, fmt.Errorf("could not get micro call graphs: %v", err)
	}

	return &Result{
		Config:   *config,
		System:   system,
		SystemCG: systemCG,
		MicroCGs: microCGs,
		IDer:     ider,
	}, nil
}

func FromDotsConfig(dirs []string, ider *IDer, config profile.OutConfig) (Map, *IDer, error) {
	if ider == nil {
		ider = NewIDer()
	}

	cgs := make(Map)
	for _, path := range dirs {
		dir, err := os.Open(path)
		if err != nil {
			return nil, nil, fmt.Errorf("could not open dots dir: %v", err)
		}
		defer dir.Close()

		fileInfos, err := dir.Readdir(-1)
		if err != nil {
			return nil, nil, fmt.Errorf("could not get FileInfos: %v", err)
		}

		for _, fileInfo := range fileInfos {
			if fileInfo.IsDir() {
				continue
			}

			fn := fileInfo.Name()

			if !strings.HasSuffix(fn, ".dot") {
				continue
			}

			// only consider files of type (with file suffix)
			if !strings.HasSuffix(fn, config.FileSuffix()) {
				continue
			}

			cg, err := FromDot(path, fileInfo, ider)
			if err != nil {
				return nil, nil, fmt.Errorf("could not get cg for dot file '%s': %v", fn, err)
			}

			cgs[fn] = cg
		}
	}

	return cgs, ider, nil
}

func FromDot(path string, fileInfo os.FileInfo, ider *IDer) (graph.WeightedDirected, error) {
	cg := simple.NewWeightedDirectedGraph(-1, -2)

	fp := filepath.Join(path, fileInfo.Name())
	f, err := os.Open(fp)
	if err != nil {
		return nil, fmt.Errorf("could not open dot file '%s': %v", fp, err)
	}
	defer f.Close()

	dotIDtoID := make(map[string]int64)

	r := bufio.NewReader(f)
Loop:
	for {
		l, err := r.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break Loop
			}
			return nil, fmt.Errorf("error reading line: %v", err)
		}

		if !strings.HasPrefix(l, "N") {
			continue Loop
		}

		// fix to support generic function names
		l = strings.ReplaceAll(l, "[…]", "")
		l = strings.ReplaceAll(l, "[...]", "")
		splitted := strings.Split(l, "[")
		if ls := len(splitted); ls < 2 {
			return nil, fmt.Errorf("invalid line: %s", l)
		}

		nodeEdge := strings.TrimSpace(splitted[0])
		args := strings.TrimSpace(splitted[1])

		if strings.Contains(nodeEdge, "->") {
			edge, err := edgeFrom(cg, dotIDtoID, nodeEdge, args)
			if err != nil {
				return nil, fmt.Errorf("could not get edge: %v", err)
			}
			cg.SetWeightedEdge(edge)
		} else {
			if strings.Contains(args, "tooltip") {
				node, err := nodeFrom(ider, args)
				if err != nil {
					return nil, fmt.Errorf("could not get node: %v", err)
				}
				dotIDtoID[nodeEdge] = node.ID()
				cg.AddNode(node)
			} else {
				fmt.Printf("Error in file %s\n", f.Name())
			}
		}
	}

	return cg, nil
}

func (r *Result) Copy() *Result {
	newSystemCG := simple.NewWeightedDirectedGraph(-1, -2)
	graph.CopyWeighted(newSystemCG, r.SystemCG)

	newMicroCGs := make(Map)
	for k, v := range r.MicroCGs {
		newMicroCG := simple.NewWeightedDirectedGraph(-1, -2)
		graph.CopyWeighted(newMicroCG, v)
		newMicroCGs[k] = newMicroCG
	}

	return &Result{
		Config:   r.Config,
		System:   r.System,
		IDer:     r.IDer.Copy(),
		SystemCG: newSystemCG,
		MicroCGs: newMicroCGs,
	}
}

// nodes (functions)

func nodeFrom(ider *IDer, args string) (*Function, error) {
	name := nameFromDot(args)
	node := NewFunction(ider, name)
	functionTime, totalTime, err := nodeWeightFromDot(args)
	if err != nil {
		return nil, fmt.Errorf("could not parse function '%s' times: %v", name, err)
	}
	node.FunctionTime = functionTime
	node.TotalTime = totalTime
	return node, nil
}

func nodeWeightFromDot(args string) (function, total time.Duration, err error) {
	// label="runtime\nsystemstack\n0.06s (0.14%)\nof 12.99s (30.20%)"
	val := attributeFromDot(args, "label")
	splitted := strings.Split(strings.ReplaceAll(val, "\\n", " "), " ")
	lenSplitted := len(splitted)

	if splitted[lenSplitted-3] == "of" {
		// has outgoing calls
		totalStr := splitted[lenSplitted-2]
		totalTime, err := time.ParseDuration(totalStr)
		if err != nil {
			return 0, 0, fmt.Errorf("could not parse total time '%s': %v", totalStr, err)
		}

		functionStr := splitted[lenSplitted-4]
		if strings.HasPrefix(functionStr, "(") {
			functionStr = splitted[lenSplitted-5]
		}
		functionTime, err := time.ParseDuration(functionStr)
		if err != nil {
			return 0, 0, fmt.Errorf("could not parse function time '%s': %v", functionStr, err)
		}
		return functionTime, totalTime, nil
	}

	// leaf node
	timeStr := splitted[lenSplitted-2]
	time, err := time.ParseDuration(timeStr)
	if err != nil {
		return 0, 0, fmt.Errorf("leaf node - could not parse time '%s': %v", timeStr, err)
	}
	return time, time, nil
}

// edges (calls)

func edgeFrom(cg graph.WeightedDirected, dotIDtoID map[string]int64, nodeEdge, args string) (*Call, error) {
	nodeEdgeSplitted := strings.Split(nodeEdge, "->")
	if l := len(nodeEdgeSplitted); l != 2 {
		return nil, fmt.Errorf("could not get from to IDs from '%s': length expected %d was %d", nodeEdge, 2, l)
	}
	fromDotID := strings.TrimSpace(nodeEdgeSplitted[0])
	toDotID := strings.TrimSpace(nodeEdgeSplitted[1])

	fromNode, err := nodeFromDotID(cg, dotIDtoID, fromDotID, "from")
	if err != nil {
		return nil, fmt.Errorf("could not get from Node: %v", err)
	}

	toNode, err := nodeFromDotID(cg, dotIDtoID, toDotID, "to")
	if err != nil {
		return nil, fmt.Errorf("could not get to Node: %v", err)
	}

	edgeWeight, err := edgeWeightFromDot(args)
	if err != nil {
		return nil, fmt.Errorf("could not parse edge time: %v", err)
	}

	edge := NewCall(fromNode, toNode, edgeWeight)
	return edge, nil
}

func nodeFromDotID(cg graph.WeightedDirected, dotIDtoID map[string]int64, dotID, nodeType string) (*Function, error) {
	id, ok := dotIDtoID[dotID]
	if !ok {
		return nil, fmt.Errorf("%s dot ID '%s' unknown", nodeType, dotID)
	}
	graphNode := cg.Node(id)
	if graphNode == nil {
		return nil, fmt.Errorf("ID '%d' does not exist in graph", id)
	}

	if node, ok := graphNode.(*Function); ok {
		return node, nil
	}
	// could not convert graph.Node to Node
	return nil, fmt.Errorf("could not convert graph.Node to Node for '%s'", dotID)
}

func edgeWeightFromDot(args string) (time.Duration, error) {
	// label=" 25.55s"
	val := attributeFromDot(args, "label")
	trimmed := strings.TrimSpace(val)
	trimmedSplitted := strings.Split(strings.ReplaceAll(trimmed, "\\n", " "), " ")
	dur, err := time.ParseDuration(trimmedSplitted[0])
	if err != nil {
		return 0, err
	}
	return dur, nil
}

// general node and edge functions

func nameFromDot(args string) string {
	idx := strings.Index(args, "tooltip")
	var sb strings.Builder
	var add bool
	for _, c := range args[idx:] {
		if c == '"' {
			add = true
			continue
		}

		if c == ' ' {
			break
		}

		if add {
			sb.WriteRune(c)
		}
	}
	return sb.String()
}

func attributeFromDot(args, attribute string) string {
	idx := strings.Index(args, fmt.Sprintf("%s=", attribute))
	var sb strings.Builder
	var add bool
Loop:
	for _, c := range args[idx:] {
		if c == '"' {
			if !add {
				add = true
				continue
			}
			break Loop
		}

		if add {
			sb.WriteRune(c)
		}
	}
	return sb.String()
}

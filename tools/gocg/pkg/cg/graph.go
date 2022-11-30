package cg

import (
	"fmt"
	"time"

	"gonum.org/v1/gonum/graph"
)

type Function struct {
	id           int64
	Name         string
	FunctionTime time.Duration
	TotalTime    time.Duration
}

func NewFunction(ider *IDer, name string) *Function {
	return &Function{
		id:   ider.ID(name),
		Name: name,
	}
}

func (n *Function) ID() int64 {
	return n.id
}

func (n *Function) String() string {
	return fmt.Sprintf("Function(id=%d,Name=%s,FunctionTime=%v,TotalTime=%v)", n.id, n.Name, n.FunctionTime, n.TotalTime)
}

type Call struct {
	from     graph.Node
	to       graph.Node
	CallTime time.Duration
}

func NewCall(from, to *Function, callTime time.Duration) *Call {
	return &Call{
		from:     from,
		to:       to,
		CallTime: callTime,
	}
}

func (e *Call) From() graph.Node {
	return e.from
}

func (e *Call) To() graph.Node {
	return e.to
}

func (e *Call) ReversedEdge() graph.Edge {
	return &Call{
		from:     e.to,
		to:       e.from,
		CallTime: e.CallTime,
	}
}

func (e *Call) Weight() float64 {
	return float64(e.CallTime)
}

func (e *Call) String() string {
	return fmt.Sprintf("Call(from=%s,to=%s,CallTime=%v)", e.from, e.to, e.CallTime)
}

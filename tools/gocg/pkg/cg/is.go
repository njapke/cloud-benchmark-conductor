package cg

import (
	"strings"

	"gonum.org/v1/gonum/graph"
)

func IsProject(project string) ExclusionFunc {
	return func(n graph.Node, level int) bool {
		f := n.(*Function) // panics -> should never happen
		return strings.HasPrefix(f.Name, project)
	}
}

func IsProjects(projects []string) ExclusionFunc {
	return func(n graph.Node, level int) bool {
		f := n.(*Function) // panics -> should never happen

		var valid bool
	Loop:
		for _, project := range projects {
			if strings.HasPrefix(f.Name, project) {
				valid = true
				break Loop
			}
		}

		return valid
	}
}

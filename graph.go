package gotaskflow

import (
	"sync"

	"github.com/noneback/go-taskflow/utils"
)

type Graph struct {
	name        string
	nodes       []*Node
	joinCounter utils.RC
	entries     []*Node
	scheCond    *sync.Cond
}

func newGraph(name string) *Graph {
	return &Graph{
		name:     name,
		nodes:    make([]*Node, 0),
		scheCond: sync.NewCond(&sync.Mutex{}),
	}
}

func (g *Graph) JoinCounter() int {
	return g.joinCounter.Value()
}

func (g *Graph) Push(n ...*Node) {
	g.nodes = append(g.nodes, n...)
	for _, node := range n {
		node.g = g
	}
}

func (g *Graph) setup() {
	for _, node := range g.nodes {
		g.joinCounter.Increase()
		node.joinCounter.Set(len(node.dependents))

		if len(node.dependents) == 0 {
			g.entries = append(g.entries, node)
		}
	}
}

func (g *Graph) instancelize() {
	for _, node := range g.nodes {
		if subflow, ok := node.ptr.(*Subflow); ok {
			subflow.handle(subflow)
		}
	}
}

func (g *Graph) topologicalSort() ([]*Node, bool) {
	// g.instancelize()
	indegree := map[*Node]int{} // Node -> indegree
	zeros := make([]*Node, 0)   // zero deps
	sorted := make([]*Node, 0, len(g.nodes))

	for _, node := range g.nodes {
		set := map[*Node]struct{}{}
		for _, dep := range node.dependents {
			set[dep] = struct{}{}
		}
		indegree[node] = len(set)
		if len(set) == 0 {
			zeros = append(zeros, node)
		}
	}

	for len(zeros) > 0 {
		node := zeros[0]
		zeros = zeros[1:]
		sorted = append(sorted, node)

		for _, succeesor := range node.successors {
			in := indegree[succeesor]
			in = in - 1
			if in <= 0 { // successor has no deps, put into zeros list
				zeros = append(zeros, succeesor)
			}
			indegree[succeesor] = in
		}
	}

	for _, node := range g.nodes {
		if indegree[node] > 0 {
			return nil, false
		}
	}

	return sorted, true
}

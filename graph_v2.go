package gotaskflow

type TopologicalSortable interface {
	TopologicalSort() ([]TopologicalSortable, bool)
	Dependents() []TopologicalSortable
	Successors() []TopologicalSortable
	// Unfold() []*Node
}

type DAG struct {
	name  string
	nodes []TopologicalSortable
}

func newDAG(name string) *DAG {
	return &DAG{
		name:  name,
		nodes: make([]TopologicalSortable, 0),
	}
}

func (g *DAG) Dependents() []TopologicalSortable {
	return nil
}

func (g *DAG) Successors() []TopologicalSortable {
	return nil
}

func (g *DAG) push(n ...TopologicalSortable) {
	g.nodes = append(g.nodes, n...)
}

// TODO: impl sorting
func (g *DAG) TopologicalSort() ([]*Node, bool) {
	indegree := map[TopologicalSortable]int{} // TopologicalSortable -> indegree
	zeros := make([]TopologicalSortable, 0)   // zero deps
	sorted := make([]*Node, 0, len(g.nodes))

	// calculate indegree
	for _, node := range g.nodes {
		set := map[TopologicalSortable]struct{}{}
		for _, dep := range node.Dependents() {
			set[dep] = struct{}{}
		}
		indegree[node] = len(set)
		if len(set) == 0 {
			zeros = append(zeros, node)
		}
	}

	for len(zeros) > 0 {
		node := zeros[0]
		unfold, ok := node.TopologicalSort()
		if !ok {
			panic("err unfold")
		}

		zeros = zeros[1:]
		sorted = append(sorted, unfold...)

		for _, succeesor := range node.Successors() {
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

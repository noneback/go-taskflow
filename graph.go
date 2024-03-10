package gotaskflow

type Graph struct {
	name  string
	nodes []kNode
}

func newGraph(name string) *Graph {
	return &Graph{
		name:  name,
		nodes: make([]kNode, 0),
	}
}

func (g *Graph) push(n ...kNode) {
	g.nodes = append(g.nodes, n...)
}

// TODO: impl sorting
func (g *Graph) TopologicalSort() ([]kNode, bool) {
	indegree := map[string]int{} // Node -> indegree
	zeros := make([]kNode, 0)    // zero deps
	sorted := make([]kNode, 0, len(g.nodes))

	for _, node := range g.nodes {
		set := map[string]struct{}{}
		for _, dep := range node.Dependents() {
			set[dep.Name()] = struct{}{}
		}
		indegree[node.Name()] = len(set)
		if len(set) == 0 {
			zeros = append(zeros, node)
		}
	}

	for len(zeros) > 0 {
		node := zeros[0]
		zeros = zeros[1:]
		sorted = append(sorted, node)

		for _, succeesor := range node.Successors() {
			in := indegree[succeesor.Name()]
			in = in - 1
			if in <= 0 { // successor has no deps, put into zeros list
				zeros = append(zeros, succeesor)
			}
			indegree[succeesor.Name()] = in
		}
	}

	for _, node := range g.nodes {
		if indegree[node.Name()] > 0 {
			return nil, false
		}
	}

	return sorted, true
}

package gotaskflow

type Graph struct {
	name  string
	nodes []*Node
}

func newGraph(name string) *Graph {
	return &Graph{
		name:  name,
		nodes: make([]*Node, 0),
	}
}

func (g *Graph) push(n ...*Node) {
	g.nodes = append(g.nodes, n...)
}

// TODO: impl sorting
func (g *Graph) TopologicalSort() ([]*Node, bool) {
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

func (g *Graph) Dependents() []TopologicalSortable {
	depends := make([]TopologicalSortable, len(g.nodes))
	for i, d := range g.nodes {
		depends[i] = d // 将 *Node 转换为 TopologicalSortable
	}
	return depends
}

func (g *Graph) Successors() []TopologicalSortable {
	succs := make([]TopologicalSortable, len(n.successors))
	for i, s := range n.successors {
		succs[i] = s // 将 *Node 转换为 TopologicalSortable
	}
	return succs
}

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

func (g *Graph) push(n *Node) {
	g.nodes = append(g.nodes, n)
}

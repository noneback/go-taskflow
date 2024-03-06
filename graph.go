package gotaskflow

type Graph struct {
	nodes []*Node
}

func newGraph() *Graph {
	return &Graph{
		make([]*Node, 0),
	}
}

func (g *Graph) push(n *Node) {
	g.nodes = append(g.nodes, n)
}

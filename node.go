package gotaskflow

type kNodeState uint8

type Node struct {
	successors []*Node
	dependents []*Node
	handle     TaskHandle
	state      kNodeState
}

func newNode() *Node {
	return &Node{
		successors: make([]*Node, 0),
		dependents: make([]*Node, 0),
	}
}

func newNodeWithHandle(f TaskHandle) *Node {
	node := newNode()
	node.handle = f
	return node
}

// set dependency： V deps on N, V is input node
func (n *Node) precede(v *Node) {
	n.successors = append(n.successors, v)
	v.dependents = append(v.dependents, n)
}

package gotaskflow

type kNodeState uint8

const (
	kNodeStateWaiting = 1
	kNodeStateRunning = 2
)

type Node struct {
	name       string
	successors []*Node
	dependents []*Node
	handle     TaskHandle
	state      kNodeState
}

func newNode(name string) *Node {
	return &Node{
		name:       name,
		successors: make([]*Node, 0),
		dependents: make([]*Node, 0),
	}
}

func newNodeWithHandle(name string, f TaskHandle) *Node {
	node := newNode(name)
	node.handle = f
	return node
}

// set dependency： V deps on N, V is input node
func (n *Node) precede(v *Node) {
	n.successors = append(n.successors, v)
	v.dependents = append(v.dependents, n)
}

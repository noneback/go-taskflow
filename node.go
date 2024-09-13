package gotaskflow

type kNodeState int32

const (
	kNodeStateWaiting  = 1
	kNodeStateRunning  = 2
	kNodeStateFinished = 3
)

type Node struct {
	name       string
	successors []TopologicalSortable
	dependents []TopologicalSortable
	handle     TaskHandle
	state      kNodeState
}

func newNode(name string) *Node {
	return &Node{
		name:       name,
		state:      kNodeStateWaiting,
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
func (n *Node) Precede(v *TopologicalSortable) {
	n.successors = append(n.successors, v)
	v.dependents = append(v.dependents, n)
}

func (n *Node) TopologicalSort() ([]TopologicalSortable, bool) {
	return []TopologicalSortable{n}, true
}

// func (n *Node) Dependents() []TopologicalSortable {
// 	return n.dependents
// }
// func (n *Node) Successors() []TopologicalSortable {
// 	return n.successors
// }

func (n *Node) Dependents() []TopologicalSortable {
	depends := make([]TopologicalSortable, len(n.dependents))
	for i, d := range n.dependents {
		depends[i] = d // 将 *Node 转换为 TopologicalSortable
	}
	return depends
}

func (n *Node) Successors() []TopologicalSortable {
	succs := make([]TopologicalSortable, len(n.successors))
	for i, s := range n.successors {
		succs[i] = s // 将 *Node 转换为 TopologicalSortable
	}
	return succs
}

// func (n *Node) Unfold() []*Node {
// 	return []*Node{n}
// }

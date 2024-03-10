package gotaskflow

type ConditionNode struct {
	name       string
	successors []kNode
	dependents []kNode
	state      kNodeState
	handle     ConditionTaskHandle
	mapper     map[int]kNode
}

func newCondNode(name string) *ConditionNode {
	return &ConditionNode{
		name:       name,
		state:      kNodeStateWaiting,
		successors: make([]kNode, 0),
		dependents: make([]kNode, 0),
		mapper:     make(map[int]kNode),
	}
}

func newCondNodeWithHandle(name string, f ConditionTaskHandle) *ConditionNode {
	node := newCondNode(name)
	node.handle = f
	return node
}

func (n *ConditionNode) Handle() interface{} {
	return n.handle
}
func (n *ConditionNode) Dependents() []kNode {
	return n.dependents
}
func (n *ConditionNode) Successors() []kNode {
	return n.successors
}

func (n *ConditionNode) Name() string {
	return n.name
}

func (n *ConditionNode) State() *kNodeState {
	return &n.state
}

// set dependency： V deps on N, V is input node
func (n *ConditionNode) Precede(v kNode) {
	n.successors = append(n.successors, v)
	v.Succeed(n)
}

func (n *ConditionNode) Succeed(v kNode) {
	n.successors = append(n.successors, v)
}

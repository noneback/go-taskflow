package gotaskflow

type kNodeState int32

const (
	kNodeStateWaiting  = 1
	kNodeStateRunning  = 2
	kNodeStateFinished = 3
)

type kNode interface {
	Name() string
	Successors() []kNode
	Dependents() []kNode
	Handle() interface{}
	State() *kNodeState
	Precede(n kNode)
	Succeed(n kNode)
}

type RawNode struct {
	name       string
	successors []kNode
	dependents []kNode
	handle     TaskHandle
	state      kNodeState
}

func newNode(name string) *RawNode {
	return &RawNode{
		name:       name,
		state:      kNodeStateWaiting,
		successors: make([]kNode, 0),
		dependents: make([]kNode, 0),
	}
}

func newNodeWithHandle(name string, f TaskHandle) *RawNode {
	node := newNode(name)
	node.handle = f
	return node
}

func (n *RawNode) Handle() interface{} {
	return n.handle
}
func (n *RawNode) Dependents() []kNode {
	return n.dependents
}
func (n *RawNode) Successors() []kNode {
	return n.successors
}

func (n *RawNode) Name() string {
	return n.name
}

func (n *RawNode) State() *kNodeState {
	return &n.state
}

func (n *RawNode) Succeed(v kNode) {
	n.dependents = append(n.dependents, v)
}

// set dependency： V deps on N, V is input node
func (n *RawNode) Precede(v kNode) {
	n.successors = append(n.successors, v)
	v.Succeed(n)
}

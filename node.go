package gotaskflow

import (
	"sync"

	"github.com/noneback/go-taskflow/utils"
)

type kNodeState int32

const (
	kNodeStateWaiting  = 1
	kNodeStateRunning  = 2
	kNodeStateFinished = 3
)

type NodeType string

const (
	NodeSubflow NodeType = "subflow"
	NodeStatic  NodeType = "task"
)

type Node struct {
	name        string
	successors  []*Node
	dependents  []*Node
	Typ         NodeType
	ptr         interface{}
	rw          *sync.RWMutex
	state       kNodeState
	joinCounter utils.RC
	g           *Graph
}

func (n *Node) JoinCounter() int {
	return n.joinCounter.Value()
}

func (n *Node) drop() {
	// release every deps
	for _, node := range n.successors {
		node.joinCounter.Decrease()
	}

	n.g.joinCounter.Decrease()
}

// set dependency： V deps on N, V is input node
func (n *Node) precede(v *Node) {
	n.successors = append(n.successors, v)
	v.dependents = append(v.dependents, n)
}

func newNode(name string) *Node {
	return &Node{
		name:       name,
		state:      kNodeStateWaiting,
		successors: make([]*Node, 0),
		dependents: make([]*Node, 0),
		rw:         &sync.RWMutex{},
	}
}

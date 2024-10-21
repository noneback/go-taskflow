package gotaskflow

import (
	"sync"
	"sync/atomic"

	"github.com/noneback/go-taskflow/utils"
)

const (
	kNodeStateIdle     = int32(0)
	kNodeStateWaiting  = int32(1)
	kNodeStateRunning  = int32(2)
	kNodeStateFinished = int32(3)
	kNodeStateFailed   = int32(4)
	kNodeStateCanceled = int32(5)
)

type NodeType string

const (
	NodeSubflow   NodeType = "subflow"   // subflow
	NodeStatic    NodeType = "static"    // static
	NodeCondition NodeType = "condition" // static
)

type Node struct {
	name        string
	successors  []*Node
	dependents  []*Node
	Typ         NodeType
	ptr         interface{}
	rw          *sync.RWMutex
	state       atomic.Int32
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
		state:      atomic.Int32{},
		successors: make([]*Node, 0),
		dependents: make([]*Node, 0),
		rw:         &sync.RWMutex{},
	}
}

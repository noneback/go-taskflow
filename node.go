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
	kNodeStateIgnored  = int32(6)
)

type innerNodeType string

const (
	nodeSubflow   innerNodeType = "subflow"   // subflow
	nodeStatic    innerNodeType = "static"    // static
	nodeCondition innerNodeType = "condition" // condition
)

type innerNode struct {
	name        string
	successors  []*innerNode
	dependents  []*innerNode
	Typ         innerNodeType
	ptr         interface{}
	rw          *sync.RWMutex
	state       atomic.Int32
	joinCounter *utils.RC
	g           *eGraph
	priority    TaskPriority
}

func (n *innerNode) JoinCounter() int {
	return n.joinCounter.Value()
}
func (n *innerNode) setup() {
	n.state.Store(kNodeStateIdle)
	if n.Typ == nodeCondition {
		return
	}

	for _, dep := range n.dependents {
		if dep.Typ == nodeCondition {
			continue
		}
		n.joinCounter.Increase()
	}
}

func (n *innerNode) drop() {
	if n.Typ == nodeCondition {
		return
	}
	// release every deps
	for _, node := range n.successors {
		node.joinCounter.Decrease()
	}
}

// set dependency： V deps on N, V is input node
func (n *innerNode) precede(v *innerNode) {
	n.successors = append(n.successors, v)
	v.dependents = append(v.dependents, n)
}

func newNode(name string) *innerNode {
	return &innerNode{
		name:        name,
		state:       atomic.Int32{},
		successors:  make([]*innerNode, 0),
		dependents:  make([]*innerNode, 0),
		rw:          &sync.RWMutex{},
		priority:    NORMAL,
		joinCounter: utils.NewRC(),
	}
}

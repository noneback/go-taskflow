package gotaskflow

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	kNodeStateIdle = int32(iota + 1)
	kNodeStateWaiting
	kNodeStateRunning
	kNodeStateFinished
)

type nodeType string

const (
	nodeSubflow   nodeType = "subflow"   // subflow
	nodeStatic    nodeType = "static"    // static
	nodeCondition nodeType = "condition" // static
)

type innerNode struct {
	name        string
	successors  []*innerNode
	dependents  []*innerNode
	Typ         nodeType
	ptr         interface{}
	mu          *sync.Mutex
	state       atomic.Int32
	joinCounter atomic.Int32
	g           *eGraph
	priority    TaskPriority
}

func (n *innerNode) recyclable() bool {
	return n.joinCounter.Load() == 0
}

func (n *innerNode) ref() {
	n.joinCounter.Add(1)
}

func (n *innerNode) deref() {
	n.joinCounter.Add(-1)
}

func (n *innerNode) setup() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.state.Store(kNodeStateIdle)
	for _, dep := range n.dependents {
		if dep.Typ == nodeCondition {
			continue
		}

		n.ref()
	}
}
func (n *innerNode) drop() {
	// release every deps
	for _, node := range n.successors {
		if n.Typ != nodeCondition {
			node.deref()
		}
	}
}

// set dependency： V deps on N, V is input node
func (n *innerNode) precede(v *innerNode) {
	n.successors = append(n.successors, v)
	v.dependents = append(v.dependents, n)
}

func newNode(name string) *innerNode {
	if len(name) == 0 {
		name = "N_" + strconv.Itoa(time.Now().Nanosecond())
	}
	return &innerNode{
		name:        name,
		state:       atomic.Int32{},
		successors:  make([]*innerNode, 0),
		dependents:  make([]*innerNode, 0),
		mu:          &sync.Mutex{},
		priority:    NORMAL,
		joinCounter: atomic.Int32{},
	}
}

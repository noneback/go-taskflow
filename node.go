package gotaskflow

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/noneback/go-taskflow/utils"
)

const (
	kNodeStateIdle = int32(iota + 1)
	kNodeStateWaiting
	kNodeStateRunning
	kNodeStateFinished
	kNodeStateFailed
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
	for _, dep := range n.dependents {
		if dep.Typ == nodeCondition {
			continue
		}

		n.joinCounter.Increase()
	}
}
func (n *innerNode) drop() {
	// release every deps
	for _, node := range n.successors {
		if n.Typ != nodeCondition {
			node.joinCounter.Decrease()
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
		rw:          &sync.RWMutex{},
		priority:    NORMAL,
		joinCounter: utils.NewRC(),
	}
}

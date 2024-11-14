package gotaskflow

import (
	"sync"
	"sync/atomic"

	"github.com/noneback/go-taskflow/utils"
)

type eGraph struct { // execution graph
	name          string
	nodes         []*innerNode
	joinCounter   *utils.RC
	entries       []*innerNode
	scheCond      *sync.Cond
	instancelized bool
	canceled      atomic.Bool // only changes when task in graph panic
}

func newGraph(name string) *eGraph {
	return &eGraph{
		name:        name,
		nodes:       make([]*innerNode, 0),
		scheCond:    sync.NewCond(&sync.Mutex{}),
		joinCounter: utils.NewRC(),
	}
}

func (g *eGraph) JoinCounter() int {
	return g.joinCounter.Value()
}

func (g *eGraph) reset() {
	g.joinCounter.Set(0)
	g.entries = g.entries[:0]
	for _, n := range g.nodes {
		n.joinCounter.Set(0)
	}
}

func (g *eGraph) push(n ...*innerNode) {
	g.nodes = append(g.nodes, n...)
	for _, node := range n {
		node.g = g
	}
}

func (g *eGraph) setup() {
	g.reset()

	for _, node := range g.nodes {
		node.setup()

		if len(node.dependents) == 0 {
			g.entries = append(g.entries, node)
		}
	}
}

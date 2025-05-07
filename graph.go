package gotaskflow

import (
	"sync"
	"sync/atomic"
)

type eGraph struct { // execution graph
	name         string
	nodes        []*innerNode
	joinCounter  atomic.Int32
	entries      []*innerNode
	scheCond     *sync.Cond
	instantiated bool
	canceled     atomic.Bool // only changes when task in graph panic
}

func newGraph(name string) *eGraph {
	return &eGraph{
		name:        name,
		nodes:       make([]*innerNode, 0),
		scheCond:    sync.NewCond(&sync.Mutex{}),
		joinCounter: atomic.Int32{},
	}
}

func (g *eGraph) ref() {
	g.joinCounter.Add(1)
}

func (g *eGraph) deref() {
	g.scheCond.L.Lock()
	defer g.scheCond.L.Unlock()
	defer g.scheCond.Signal()

	g.joinCounter.Add(-1)
}

func (g *eGraph) reset() {
	g.joinCounter.Store(0)
	g.entries = g.entries[:0]
	for _, n := range g.nodes {
		n.joinCounter.Store(0)
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

func (g *eGraph) recyclable() bool {
	return g.joinCounter.Load() == 0
}

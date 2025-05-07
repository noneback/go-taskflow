package gotaskflow

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type eGraph struct { // execution graph
	name         string
	nodes        []*innerNode
	joinCounter  uint
	entries      []*innerNode
	scheCond     *sync.Cond
	instantiated bool
	rw           *sync.RWMutex
	canceled     atomic.Bool // only changes when task in graph panic
}

func newGraph(name string) *eGraph {
	return &eGraph{
		name:        name,
		nodes:       make([]*innerNode, 0),
		scheCond:    sync.NewCond(&sync.Mutex{}),
		joinCounter: 0,
		rw:          &sync.RWMutex{},
	}
}

func (g *eGraph) ref() {
	g.rw.Lock()
	defer g.rw.Unlock()

	g.joinCounter++
}

func (g *eGraph) deref() {
	g.scheCond.L.Lock()
	defer g.scheCond.L.Unlock()
	defer g.scheCond.Signal()

	g.rw.Lock()
	defer g.rw.Unlock()
	if g.joinCounter == 0 {
		panic(fmt.Sprintf("graph %v ref counter is zero, cannot deref", g.name))
	}
	g.joinCounter--
}

func (g *eGraph) reset() {
	g.joinCounter = 0
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

func (g *eGraph) recyclable(lockup bool) bool {
	if lockup {
		g.rw.RLock()
		defer g.rw.RUnlock()
	}

	return g.joinCounter == 0
}

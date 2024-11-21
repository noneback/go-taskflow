package gotaskflow

import (
	"cmp"
	"fmt"
	"io"
	"log"
	"runtime/debug"
	"slices"
	"sync"
	"time"

	"github.com/noneback/go-taskflow/utils"
)

// Executor schedule and execute taskflow
type Executor interface {
	Wait()                     // Wait block until all tasks finished
	Profile(w io.Writer) error // Profile write flame graph raw text into w
	Run(tf *TaskFlow) Executor // Run start to schedule and execute taskflow
}

type innerExecutorImpl struct {
	concurrency uint
	pool        *utils.Copool
	wq          *utils.Queue[*innerNode]
	wg          *sync.WaitGroup
	profiler    *profiler
}

// NewExecutor return a Executor with a specified max goroutine concurrency(recommend a value bigger than Runtime.NumCPU)
func NewExecutor(concurrency uint) Executor {
	if concurrency == 0 {
		panic("executor concrurency cannot be zero")
	}
	t := newProfiler()
	return &innerExecutorImpl{
		concurrency: concurrency,
		pool:        utils.NewCopool(concurrency),
		wq:          utils.NewQueue[*innerNode](),
		wg:          &sync.WaitGroup{},
		profiler:    t,
	}
}

// Run start to schedule and execute taskflow
func (e *innerExecutorImpl) Run(tf *TaskFlow) Executor {
	tf.forzen = true
	e.scheduleGraph(tf.graph, nil)
	return e
}

func (e *innerExecutorImpl) invokeGraph(g *eGraph, parentSpan *span) {
	for {
		g.scheCond.L.Lock()
		for g.JoinCounter() != 0 && e.wq.Len() == 0 && !g.canceled.Load() {
			g.scheCond.Wait()
		}
		g.scheCond.L.Unlock()

		// tasks can only be executed after sched, and joinCounter incr when sched, so here no need to lock up.
		if g.JoinCounter() == 0 || g.canceled.Load() {
			break
		}

		node := e.wq.Pop() // hang
		e.invokeNode(node, parentSpan)
	}
}

func (e *innerExecutorImpl) sche_successors(node *innerNode) {
	candidate := make([]*innerNode, 0, len(node.successors))

	for _, n := range node.successors {
		if n.JoinCounter() == 0 || n.Typ == nodeCondition {
			// deps all done or condition node
			candidate = append(candidate, n)
		}
	}

	slices.SortFunc(candidate, func(i, j *innerNode) int {
		return cmp.Compare(i.priority, j.priority)
	})
	node.setup()
	e.schedule(candidate...)
}

func (e *innerExecutorImpl) invokeStatic(node *innerNode, parentSpan *span, p *Static) func() {
	return func() {
		span := span{extra: attr{
			typ:  nodeStatic,
			name: node.name,
		}, begin: time.Now(), parent: parentSpan}

		defer func() {
			span.cost = time.Since(span.begin)
			if r := recover(); r != nil {
				node.g.canceled.Store(true)
				log.Printf("[recovered] node %s, panic: %s, stack: %s", node.name, r, debug.Stack())
			} else {
				e.profiler.AddSpan(&span) // remove canceled node span
			}

			node.drop()
			e.sche_successors(node)
			node.g.joinCounter.Decrease()
			e.wg.Done()
			node.g.scheCond.Signal()
		}()

		node.state.Store(kNodeStateRunning)
		p.handle()
		node.state.Store(kNodeStateFinished)
	}
}

func (e *innerExecutorImpl) invokeSubflow(node *innerNode, parentSpan *span, p *Subflow) func() {
	return func() {
		span := span{extra: attr{
			typ:  nodeSubflow,
			name: node.name,
		}, begin: time.Now(), parent: parentSpan}
		defer func() {
			span.cost = time.Since(span.begin)
			if r := recover(); r != nil {
				log.Printf("[recovered] subflow %s, panic: %s, stack: %s", node.name, r, debug.Stack())
				node.g.canceled.Store(true)
				p.g.canceled.Store(true)
			} else {
				e.profiler.AddSpan(&span) // remove canceled node span
			}

			e.scheduleGraph(p.g, &span)
			node.drop()
			e.sche_successors(node)
			node.g.joinCounter.Decrease()
			e.wg.Done()
			node.g.scheCond.Signal()
		}()

		node.state.Store(kNodeStateRunning)
		if !p.g.instancelized {
			p.handle(p)
		}
		p.g.instancelized = true
		node.state.Store(kNodeStateFinished)
	}
}

func (e *innerExecutorImpl) invokeCondition(node *innerNode, parentSpan *span, p *Condition) func() {
	return func() {
		span := span{extra: attr{
			typ:  nodeCondition,
			name: node.name,
		}, begin: time.Now(), parent: parentSpan}

		defer func() {
			span.cost = time.Since(span.begin)
			if r := recover(); r != nil {
				node.g.canceled.Store(true)
				log.Printf("[recovered] node %s, panic: %s, stack: %s", node.name, r, debug.Stack())
			} else {
				e.profiler.AddSpan(&span) // remove canceled node span
			}
			node.drop()
			// e.sche_successors(node)
			node.g.joinCounter.Decrease()
			node.setup()
			e.wg.Done()
			node.g.scheCond.Signal()
		}()

		node.state.Store(kNodeStateRunning)

		choice := p.handle()
		if choice > uint(len(p.mapper)) {
			panic(fmt.Sprintln("condition task failed, successors of condition should be more than precondition choice", p.handle()))
		}
		// do choice and cancel others
		node.state.Store(kNodeStateFinished)
		e.schedule(p.mapper[choice])
	}
}

func (e *innerExecutorImpl) invokeNode(node *innerNode, parentSpan *span) {
	switch p := node.ptr.(type) {
	case *Static:
		e.pool.Go(e.invokeStatic(node, parentSpan, p))
	case *Subflow:
		e.pool.Go(e.invokeSubflow(node, parentSpan, p))
	case *Condition:
		e.pool.Go(e.invokeCondition(node, parentSpan, p))
	default:
		panic("unsupported node")
	}
}

func (e *innerExecutorImpl) schedule(nodes ...*innerNode) {
	for _, node := range nodes {
		if node.g.canceled.Load() {
			node.g.scheCond.Signal()
			log.Printf("node %v is not scheduled, as graph %v is canceled\n", node.name, node.g.name)
			return
		}

		node.g.joinCounter.Increase()
		e.wg.Add(1)
		e.wq.Put(node)
		node.state.Store(kNodeStateWaiting)
		node.g.scheCond.Signal()
	}
}

func (e *innerExecutorImpl) scheduleGraph(g *eGraph, parentSpan *span) {
	g.setup()
	slices.SortFunc(g.entries, func(i, j *innerNode) int {
		return cmp.Compare(i.priority, j.priority)
	})

	e.schedule(g.entries...)
	e.invokeGraph(g, parentSpan)

	g.scheCond.Signal()
}

// Wait: block until all tasks finished
func (e *innerExecutorImpl) Wait() {
	e.wg.Wait()
}

// Profile write flame graph raw text into w
func (e *innerExecutorImpl) Profile(w io.Writer) error {
	return e.profiler.draw(w)
}

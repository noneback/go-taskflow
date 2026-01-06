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
	mu          *sync.Mutex
}

// NewExecutor return a Executor with a specified max goroutine concurrency(recommend a value bigger than Runtime.NumCPU, **MUST** bigger than num(subflows). )
func NewExecutor(concurrency uint) Executor {
	if concurrency == 0 {
		panic("executor concurrency cannot be zero")
	}
	t := newProfiler()
	return &innerExecutorImpl{
		concurrency: concurrency,
		pool:        utils.NewCopool(concurrency),
		wq:          utils.NewQueue[*innerNode](false),
		wg:          &sync.WaitGroup{},
		profiler:    t,
		mu:          &sync.Mutex{},
	}
}

// Run start to schedule and execute taskflow
func (e *innerExecutorImpl) Run(tf *TaskFlow) Executor {
	tf.frozen = true
	e.scheduleGraph(nil, tf.graph, nil)
	return e
}

func (e *innerExecutorImpl) invokeGraph(g *eGraph, parentSpan *span) bool {
	for {
		g.scheCond.L.Lock()
		e.mu.Lock()
		for !g.recyclable() && e.wq.Len() == 0 && !g.canceled.Load() {
			e.mu.Unlock()
			g.scheCond.Wait()
			e.mu.Lock()
		}

		g.scheCond.L.Unlock()

		// tasks can only be executed after sched, and joinCounter incr when sched, so here no need to lock up.
		if g.recyclable() || g.canceled.Load() {
			e.mu.Unlock()
			break
		}
		node := e.wq.Pop()
		e.mu.Unlock()

		e.invokeNode(node, parentSpan)
	}
	return !g.canceled.Load()
}

func (e *innerExecutorImpl) sche_successors(node *innerNode) {
	candidate := make([]*innerNode, 0, len(node.successors))

	for _, n := range node.successors {
		n.mu.Lock()
		if n.recyclable() && n.state.Load() == kNodeStateIdle {
			// deps all done or condition node or task has been sched.
			n.state.Store(kNodeStateWaiting)
			candidate = append(candidate, n)
		}
		n.mu.Unlock()
	}

	slices.SortFunc(candidate, func(i, j *innerNode) int {
		return cmp.Compare(i.priority, j.priority)
	})
	node.setup() // make node repeatable
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
				log.Printf("graph %v is canceled, since static node %v panics", node.g.name, node.name)
				log.Printf("[recovered] static node %s, panic: %v, stack: %s", node.name, r, debug.Stack())
			} else {
				e.profiler.AddSpan(&span) // remove canceled node span
			}

			node.drop()
			e.sche_successors(node)
			node.g.deref()
			e.wg.Done()
		}()
		if !node.g.canceled.Load() {
			node.state.Store(kNodeStateRunning)
			p.handle()
			node.state.Store(kNodeStateFinished)
		}
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
				log.Printf("graph %v is canceled, since subflow %v panics", node.g.name, node.name)
				log.Printf("[recovered] subflow %s, panic: %v, stack: %s", node.name, r, debug.Stack())
				node.g.canceled.Store(true)
				p.g.canceled.Store(true)
			} else {
				e.profiler.AddSpan(&span) // remove canceled node span
			}

			e.scheduleGraph(node.g, p.g, &span)
			node.drop()
			e.sche_successors(node)
			node.g.deref()
			e.wg.Done()
		}()

		if !node.g.canceled.Load() {
			node.state.Store(kNodeStateRunning)
			if !p.g.instantiated {
				p.handle(p)
			}
			p.g.instantiated = true
			node.state.Store(kNodeStateFinished)
		}
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
				log.Printf("graph %v is canceled, since condition node %v panics", node.g.name, node.name)
				log.Printf("[recovered] condition node %s, panic: %v, stack: %s", node.name, r, debug.Stack())
			} else {
				e.profiler.AddSpan(&span) // remove canceled node span
			}
			node.drop()
			// e.sche_successors(node)
			node.g.deref()
			node.setup()
			e.wg.Done()
		}()

		if !node.g.canceled.Load() {
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

func (e *innerExecutorImpl) pushIntoQueue(node *innerNode) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.wq.Put(node)
}

func (e *innerExecutorImpl) schedule(nodes ...*innerNode) {
	for _, node := range nodes {
		if node.g.canceled.Load() {
			// no need
			node.g.scheCond.L.Lock()
			node.g.scheCond.Signal()
			node.g.scheCond.L.Unlock()
			log.Printf("node %v is not scheduled, since graph %v is canceled\n", node.name, node.g.name)
			return
		}
		e.wg.Add(1)
		node.g.scheCond.L.Lock()

		node.g.ref()
		e.pushIntoQueue(node)

		node.g.scheCond.Signal()
		node.g.scheCond.L.Unlock()
	}
}

func (e *innerExecutorImpl) scheduleGraph(parentg, g *eGraph, parentSpan *span) {
	g.setup()
	slices.SortFunc(g.entries, func(i, j *innerNode) int {
		return cmp.Compare(i.priority, j.priority)
	})
	e.schedule(g.entries...)
	if !e.invokeGraph(g, parentSpan) && parentg != nil {
		parentg.canceled.Store(true)
		log.Printf("graph %s canceled, since subgraph %s is canceled\n", parentg.name, g.name)
	}

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

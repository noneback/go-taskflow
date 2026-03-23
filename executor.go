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
	Trace(w io.Writer) error   // Trace write Chrome Trace Event data into w
	Run(tf *TaskFlow) Executor // Run start to schedule and execute taskflow
}

type innerExecutorImpl struct {
	concurrency uint
	pool        *utils.Copool
	wq          *utils.Queue[*innerNode]
	wg          *sync.WaitGroup
	profiler    *profiler
	tracer      *tracer
	mu          *sync.Mutex
}

// NewExecutor returns an Executor with the specified concurrency and options.
// concurrency must be > 0. Recommend concurrency > runtime.NumCPU and MUST > num(subflows).
func NewExecutor(concurrency uint, opts ...Option) Executor {
	if concurrency == 0 {
		panic("executor concurrency cannot be zero")
	}
	e := &innerExecutorImpl{
		concurrency: concurrency,
		wq:          utils.NewQueue[*innerNode](false),
		wg:          &sync.WaitGroup{},
		mu:          &sync.Mutex{},
	}
	for _, opt := range opts {
		opt(e)
	}
	e.pool = utils.NewCopool(e.concurrency)
	return e
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

// record submits the span to active observers.
// ok=true means the node completed without panic; profiler only records successful spans.
func (e *innerExecutorImpl) record(s *span, ok bool) {
	if ok && e.profiler != nil {
		e.profiler.AddSpan(s)
	}
	if e.tracer != nil {
		e.tracer.AddEvent(s)
	}
}

// getDependentNames extracts predecessor task names from a node.
func getDependentNames(node *innerNode) []string {
	if len(node.dependents) == 0 {
		return nil
	}
	names := make([]string, len(node.dependents))
	for i, dep := range node.dependents {
		names[i] = dep.name
	}
	return names
}

func (e *innerExecutorImpl) invokeStatic(node *innerNode, parentSpan *span, p *Static) func() {
	return func() {
		span := span{extra: attr{
			typ:  nodeStatic,
			name: node.name,
		}, begin: time.Now(), parent: parentSpan, dependents: getDependentNames(node)}

		defer func() {
			span.cost = time.Since(span.begin)
			r := recover()
			if r != nil {
				node.g.canceled.Store(true)
				log.Printf("[go-taskflow] graph %q canceled: static task %q panicked: %v\n%s", node.g.name, node.name, r, debug.Stack())
			}
			e.record(&span, r == nil)

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
		}, begin: time.Now(), parent: parentSpan, dependents: getDependentNames(node)}

		defer func() {
			span.cost = time.Since(span.begin)
			r := recover()
			if r != nil {
				log.Printf("[go-taskflow] graph %q canceled: subflow %q panicked: %v\n%s", node.g.name, node.name, r, debug.Stack())
				node.g.canceled.Store(true)
				p.g.canceled.Store(true)
			}
			e.record(&span, r == nil)

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
		}, begin: time.Now(), parent: parentSpan, dependents: getDependentNames(node)}

		defer func() {
			span.cost = time.Since(span.begin)
			r := recover()
			if r != nil {
				node.g.canceled.Store(true)
				log.Printf("[go-taskflow] graph %q canceled: condition task %q panicked: %v\n%s", node.g.name, node.name, r, debug.Stack())
			}
			e.record(&span, r == nil)
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
			// graph already canceled, skip scheduling
			node.g.scheCond.L.Lock()
			node.g.scheCond.Signal()
			node.g.scheCond.L.Unlock()
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
	}

	g.scheCond.Signal()
}

// Wait: block until all tasks finished
func (e *innerExecutorImpl) Wait() {
	e.wg.Wait()
}

// Profile write flame graph raw text into w
func (e *innerExecutorImpl) Profile(w io.Writer) error {
	if e.profiler == nil {
		return nil
	}
	return e.profiler.draw(w)
}

// Trace write Chrome Trace Event data into w
func (e *innerExecutorImpl) Trace(w io.Writer) error {
	if e.tracer == nil {
		return nil
	}
	return e.tracer.draw(w)
}

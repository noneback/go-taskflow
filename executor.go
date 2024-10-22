package gotaskflow

import (
	"context"
	"fmt"
	"io"
	"runtime/debug"
	"sync"
	"time"

	"github.com/noneback/go-taskflow/utils"
)

type Executor interface {
	Wait()                     // Wait until all tasks finished
	Profile(w io.Writer) error // Write flame graph raw text into w
	Run(tf *TaskFlow) Executor // Run taskflow parallally
}

type innerExecutorImpl struct {
	concurrency uint
	pool        *utils.Copool
	wq          *utils.Queue[*innerNode]
	wg          *sync.WaitGroup
	profiler    *profiler
}

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

func (e *innerExecutorImpl) Run(tf *TaskFlow) Executor {
	tf.graph.setup()
	for _, node := range tf.graph.entries {
		e.schedule(node)
	}

	e.profiler.Start()
	defer e.profiler.Stop()
	e.invoke(tf)
	return e
}

func (e *innerExecutorImpl) invokeGraph(g *eGraph, parentSpan *span) {
	ctx := context.Background()
	for {
		g.scheCond.L.Lock()
		for g.JoinCounter() != 0 && e.wq.Len() == 0 && !g.canceled.Load() {
			g.scheCond.Wait()
		}
		g.scheCond.L.Unlock()

		if g.JoinCounter() == 0 || g.canceled.Load() {
			break
		}

		node := e.wq.PeakAndTake() // hang
		e.invokeNode(&ctx, node, parentSpan)
	}
}

func (e *innerExecutorImpl) invoke(tf *TaskFlow) {
	e.invokeGraph(tf.graph, nil)
}

func (e *innerExecutorImpl) invokeNode(ctx *context.Context, node *innerNode, parentSpan *span) {
	// do job
	switch p := node.ptr.(type) {
	case *Static:
		e.pool.Go(func() {
			span := span{extra: attr{
				typ:  NodeStatic,
				name: node.name,
			}, begin: time.Now(), parent: parentSpan}

			defer func() {
				span.end = time.Now()
				span.extra.success = true
				if r := recover(); r != nil {
					node.g.canceled.Store(true)
					fmt.Printf("[recovered] node %s, panic: %s, stack: %s", node.name, r, debug.Stack())
				} else {
					e.profiler.AddSpan(&span) // remove canceled node span
				}

				e.wg.Done()
				node.drop()
				for _, n := range node.successors {
					if n.JoinCounter() == 0 {
						e.schedule(n)
					}
				}
				node.g.scheCond.Signal()
			}()

			node.state.Store(kNodeStateRunning)
			p.handle()
			node.state.Store(kNodeStateFinished)
		})
	case *Subflow:
		e.pool.Go(func() {
			span := span{extra: attr{
				typ:  NodeSubflow,
				name: node.name,
			}, begin: time.Now(), parent: parentSpan}
			defer func() {
				span.end = time.Now()
				span.extra.success = true
				if r := recover(); r != nil {
					fmt.Printf("[recovered] subflow %s, panic: %s, stack: %s", node.name, r, debug.Stack())
					node.g.canceled.Store(true)
					p.g.canceled.Store(true)
				} else {
					e.profiler.AddSpan(&span) // remove canceled node span
				}
				e.wg.Done()
				e.scheduleGraph(p.g, &span)
				node.drop()

				for _, n := range node.successors {
					if n.JoinCounter() == 0 {
						e.schedule(n)
					}
				}
				node.g.scheCond.Signal()
			}()

			node.state.Store(kNodeStateRunning)
			if !p.g.instancelized {
				p.handle(p)
			}
			p.g.instancelized = true
			node.state.Store(kNodeStateFinished)
		})
	case *Condition:
		e.pool.Go(func() {
			span := span{extra: attr{
				typ:  NodeCondition,
				name: node.name,
			}, begin: time.Now(), parent: parentSpan}

			defer func() {
				span.end = time.Now()
				span.extra.success = true
				if r := recover(); r != nil {
					node.g.canceled.Store(true)
					fmt.Printf("[recovered] node %s, panic: %s, stack: %s", node.name, r, debug.Stack())
				} else {
					e.profiler.AddSpan(&span) // remove canceled node span
				}
				e.wg.Done()
				node.drop()
				for _, n := range node.successors {
					if n.JoinCounter() == 0 {
						e.schedule(n)
					}
				}
				node.g.scheCond.Signal()
			}()

			node.state.Store(kNodeStateRunning)

			choice := p.handle()
			if choice > uint(len(p.mapper)) {
				panic(fmt.Sprintln("condition task failed", p.handle()))
			}

			for idx, v := range p.mapper {
				if idx == choice {
					continue
				}
				v.state.Store(kNodeStateCanceled)
			}
			// do choice and cancel others
			node.state.Store(kNodeStateFinished)
		})

	default:
		panic("unsupported node")
	}
}

func (e *innerExecutorImpl) schedule(node *innerNode) {
	if node.g.canceled.Load() {
		node.g.scheCond.Signal()
		fmt.Printf("node %v is not scheduled, as graph %v is canceled\n", node.name, node.g.name)
		return
	}

	if node.state.Load() == kNodeStateCanceled {
		node.g.scheCond.Signal()
		fmt.Printf("node %v is canceled\n", node.name)
		for _, v := range node.successors {
			v.state.Store(kNodeStateCanceled)
		}

		return
	}

	node.g.joinCounter.Increase()
	e.wg.Add(1)
	e.wq.Put(node)
	node.state.Store(kNodeStateWaiting)
	node.g.scheCond.Signal()
}

func (e *innerExecutorImpl) scheduleGraph(g *eGraph, parentSpan *span) {
	g.setup()
	for _, node := range g.entries {
		e.schedule(node)
	}

	e.invokeGraph(g, parentSpan)

	g.scheCond.Signal()
}

func (e *innerExecutorImpl) Wait() {
	e.wg.Wait()
}

func (e *innerExecutorImpl) Profile(w io.Writer) error {
	return e.profiler.draw(w)
}

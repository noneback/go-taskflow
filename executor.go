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
	Wait()
	Profile(w io.Writer) error
	Run(tf *TaskFlow) Executor
}

type ExecutorImpl struct {
	concurrency uint
	pool        *utils.Copool
	wq          *utils.Queue[*Node]
	wg          *sync.WaitGroup
	profiler    *Profiler
}

func NewExecutor(concurrency uint) Executor {
	if concurrency == 0 {
		panic("executor concrurency cannot be zero")
	}
	t := newTracer()
	return &ExecutorImpl{
		concurrency: concurrency,
		pool:        utils.NewCopool(concurrency),
		wq:          utils.NewQueue[*Node](),
		wg:          &sync.WaitGroup{},
		profiler:    t,
	}
}

func (e *ExecutorImpl) Run(tf *TaskFlow) Executor {
	tf.graph.setup()
	for _, node := range tf.graph.entries {
		e.schedule(node)
	}

	e.profiler.Start()
	defer e.profiler.Stop()
	e.invoke(tf)
	return e
}

func (e *ExecutorImpl) invokeGraph(g *Graph, parentSpan *span) {
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

func (e *ExecutorImpl) invoke(tf *TaskFlow) {
	e.invokeGraph(tf.graph, nil)
}

func (e *ExecutorImpl) invokeNode(ctx *context.Context, node *Node, parentSpan *span) {
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
					fmt.Println("[recovered] node", node.name, "panic:", r, debug.Stack())
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
					fmt.Println("[recovered] subflow", node.name, "panic:", r, debug.Stack())
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
	default:
		panic("unsupported node")
	}
}

func (e *ExecutorImpl) schedule(node *Node) {
	if node.g.canceled.Load() {
		node.g.scheCond.Signal()
		fmt.Println("node cannot be scheduled, cuz graph canceled", node.name)
		return
	}

	e.wg.Add(1)
	e.wq.Put(node)
	node.state.Store(kNodeStateWaiting)
	node.g.scheCond.Signal()
}

func (e *ExecutorImpl) scheduleGraph(g *Graph, parentSpan *span) {
	g.setup()
	for _, node := range g.entries {
		e.schedule(node)
	}

	e.invokeGraph(g, parentSpan)

	g.scheCond.Signal()
}

func (e *ExecutorImpl) Wait() {
	e.wg.Wait()
}

func (e *ExecutorImpl) Profile(w io.Writer) error {
	return e.profiler.draw(w)
}

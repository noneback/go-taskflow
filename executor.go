package gotaskflow

import (
	"context"
	"fmt"
	"io"
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
	t := NewTracer()
	return &ExecutorImpl{
		concurrency: concurrency,
		pool:        utils.NewCopool(concurrency),
		wq:          utils.NewQueue[*Node](),
		wg:          &sync.WaitGroup{},
		profiler:    t,
	}
}

func (e *ExecutorImpl) Run(tf *TaskFlow) Executor {
	// register graceful exit when pinc
	e.pool.SetPanicHandler(func(ctx *context.Context, i interface{}) {

	})

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
		for g.JoinCounter() != 0 && e.wq.Len() == 0 {
			g.scheCond.Wait()
		}
		g.scheCond.L.Unlock()

		if g.JoinCounter() == 0 {
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
				e.profiler.AddSpan(&span)
			}()

			defer e.wg.Done()
			node.state.Store(kNodeStateRunning)
			defer node.state.Store(kNodeStateFinished)

			p.handle()
			node.drop()
			for _, n := range node.successors {
				// fmt.Println("put", n.Name)
				if n.JoinCounter() == 0 {
					e.schedule(n)
				}
			}
			node.g.scheCond.Signal()
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
				e.profiler.AddSpan(&span)
			}()

			defer e.wg.Done()
			node.state.Store(kNodeStateRunning)
			defer node.state.Store(kNodeStateFinished)

			if !p.g.instancelized {
				p.handle(p)
			}
			p.g.instancelized = true

			e.scheduleGraph(p.g, &span)
			node.drop()

			for _, n := range node.successors {
				if n.JoinCounter() == 0 {
					e.schedule(n)
				}
			}

			node.g.scheCond.Signal()
		})
	default:
		fmt.Println("exit: ", node.name)
		panic("do nothing")
	}
}

func (e *ExecutorImpl) schedule(node *Node) {
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

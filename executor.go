package gotaskflow

import (
	"context"
	"fmt"
	"sync"

	"github.com/noneback/go-taskflow/utils"
)

type Executor interface {
	Wait()
	// WaitForAll()
	Run(tf *TaskFlow) error
	// Observe()
}

type ExecutorImpl struct {
	concurrency uint
	pool        *utils.Copool
	wq          *utils.Queue[*Node]
	wg          *sync.WaitGroup
}

func NewExecutor(concurrency uint) Executor {
	if concurrency == 0 {
		panic("executor concrurency cannot be zero")
	}
	return &ExecutorImpl{
		concurrency: concurrency,
		pool:        utils.NewCopool(concurrency),
		wq:          utils.NewQueue[*Node](),
		wg:          &sync.WaitGroup{},
	}
}

func (e *ExecutorImpl) Run(tf *TaskFlow) error {
	tf.graph.setup()

	for _, node := range tf.graph.entries {
		e.schedule(node)
	}

	e.invoke(tf)
	return nil
}

func (e *ExecutorImpl) invoke_graph(g *Graph) {
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
		e.invoke_node(&ctx, node)
	}
}

func (e *ExecutorImpl) invoke(tf *TaskFlow) {
	e.invoke_graph(tf.graph)
}

func (e *ExecutorImpl) invoke_node(ctx *context.Context, node *Node) {
	// do job
	switch p := node.ptr.(type) {
	case *Static:
		e.pool.Go(func() {
			defer e.wg.Done()
			p.handle(ctx)

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
			defer e.wg.Done()

			if !p.g.instancelized {
				p.handle(p)
			}
			p.g.instancelized = true

			e.schedule_graph(p.g)
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
	node.g.scheCond.Signal()
}

func (e *ExecutorImpl) schedule_graph(g *Graph) {
	g.setup()
	for _, node := range g.entries {
		e.schedule(node)
	}

	e.invoke_graph(g)

	g.scheCond.Signal()
}

func (e *ExecutorImpl) Wait() {
	e.wg.Wait()
}

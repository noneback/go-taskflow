package gotaskflow

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Executor interface {
	Wait()
	// WaitForAll()
	Run(tf *TaskFlow) error
	// Observe()
}

type ExecutorImpl struct {
	concurrency int
	pool        Pool
	wg          *sync.WaitGroup
}

func NewExecutor(concurrency int) Executor {
	return &ExecutorImpl{
		concurrency: concurrency,
		pool:        NewTaskPool(int32(concurrency)),
		wg:          &sync.WaitGroup{},
	}
}

func (e *ExecutorImpl) Run(tf *TaskFlow) error {
	nodes, ok := tf.graph.TopologicalSort()
	if !ok {
		return ErrTaskFlowIsCyclic
	}

	ctx := context.Background()

	for _, node := range nodes {
		e.schedule(ctx, node)
	}
	return nil
}

func (e *ExecutorImpl) schedule(ctx context.Context, node kNode) {
	waitting := make(map[string]kNode)
	for _, dep := range node.Dependents() {
		waitting[dep.Name()] = dep
	}

	for len(waitting) > 0 {
		for name, dep := range waitting {
			if atomic.LoadInt32((*int32)(dep.State())) == kNodeStateFinished {
				delete(waitting, name)
			}
			// fmt.Println("Not Ready", name)
		}
		time.Sleep(time.Microsecond * 100)
	}

	e.wg.Add(1)
	e.pool.CtxGo(ctx, func() {
		defer e.wg.Done()
		atomic.StoreInt32((*int32)(node.State()), kNodeStateRunning)
		if handle, ok := node.Handle().(TaskHandle); ok {
			handle(&ctx)
		}

		if handle, ok := node.Handle().(ConditionTaskHandle); ok {
			val := handle(&ctx)
			fmt.Println(val)
		}

		atomic.StoreInt32((*int32)(node.State()), kNodeStateFinished)
	})
}

func (e *ExecutorImpl) Wait() {
	e.wg.Wait()
}

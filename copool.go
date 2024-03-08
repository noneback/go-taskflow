package gotaskflow

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

type Pool interface {
	// SetCap sets the goroutine capacity of the pool.
	SetCap(cap int32)
	// Go executes f.
	Go(f func())
	// CtxGo executes f and accepts the context.
	CtxGo(ctx context.Context, f func())
	// SetPanicHandler sets the panic handler.
	SetPanicHandler(f func(context.Context, interface{}))
}

var (
	taskPool   sync.Pool
	workerPool sync.Pool
)

func init() {
	taskPool.New = newCotask
	workerPool.New = newCoworker
}

func newCoworker() interface{} {
	return &coworker{}
}

type coworker struct {
	pool *pool
}

func (w *coworker) close() {
	w.pool.decWorkerCount()
}

func (w *coworker) run() {
	go func() {
		for {
			w.pool.taskLocker.Lock()
			if w.pool.taskQ.Len() == 0 {
				w.close()
				w.pool.taskLocker.Unlock()
				w.recycle()
				return
			}
			t := w.pool.taskQ.PeakAndTake()
			w.pool.taskLocker.Unlock()

			func() {
				defer func() {
					if r := recover(); r != nil {
						if w.pool.panicHandler != nil {
							w.pool.panicHandler(t.ctx, r)
						} else {
							msg := fmt.Sprintf("[ERROR] GOPOOL: panic in pool: %v: %s", r, debug.Stack())
							log.Println(msg)
						}
					}
				}()
				t.f()
			}()
			t.recycle()
		}
	}()

}
func (w *coworker) recycle() {
	w.zero()
	workerPool.Put(w)
}
func (w *coworker) zero() {
	w.pool = nil
}

func newCotask() interface{} {
	return &cotask{}
}

type cotask struct {
	ctx context.Context
	f   func()
}

func (ct *cotask) zero() {
	ct.ctx = nil
	ct.f = nil
}

func (ct *cotask) recycle() {
	ct.zero()
	taskPool.Put(ct)
}

type pool struct {
	panicHandler   func(context.Context, interface{})
	cap            int32
	ScaleThreshold int32
	taskQ          *Queue[*cotask]
	taskLocker     *sync.Mutex
	workerCnt      int32
}

func (p *pool) SetCap(cap int32) {
	atomic.StoreInt32(&p.cap, cap)
}

// Go executes f.
func (p *pool) Go(f func()) {
	p.CtxGo(context.Background(), f)
}

// CtxGo executes f and accepts the context.
func (p *pool) CtxGo(ctx context.Context, f func()) {
	t := taskPool.Get().(*cotask)
	t.ctx = ctx
	t.f = f

	p.taskQ.Put(t)

	if (p.taskQ.Len() >= p.ScaleThreshold && p.workerCount() <= atomic.LoadInt32(&p.cap)) || p.workerCount() == 0 {
		p.incWorkerCount()
		w := workerPool.Get().(*coworker)
		w.pool = p
		w.run()
	}

}

func (p *pool) workerCount() int32 {
	return atomic.LoadInt32(&p.workerCnt)
}

func (p *pool) incWorkerCount() {
	atomic.AddInt32(&p.workerCnt, 1)
}

func (p *pool) decWorkerCount() {
	atomic.AddInt32(&p.workerCnt, -1)
}

// SetPanicHandler sets the panic handler.
func (p *pool) SetPanicHandler(f func(context.Context, interface{})) {
	p.panicHandler = f
}

func NewTaskPool(cap int32) Pool {
	return &pool{
		panicHandler:   nil,
		ScaleThreshold: 32,
		cap:            cap,
		workerCnt:      0,
		taskLocker:     &sync.Mutex{},
		taskQ:          NewQueue[*cotask](),
	}
}

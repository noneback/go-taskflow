package utils

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
)

type cotask struct {
	ctx *context.Context
	f   func()
}

func (ct *cotask) zero() {
	ct.ctx = nil
	ct.f = nil
}

type Copool struct {
	panicHandler func(*context.Context, interface{})
	cap          uint
	taskQ        *Queue[*cotask]
	corun        *RC
	coworker     *RC
	mu           *sync.Mutex
	taskObjPool  *ObjectPool[*cotask]
}

// NewCopool return a goroutinue pool with specified cap
func NewCopool(cap uint) *Copool {
	return &Copool{
		panicHandler: nil,
		taskQ:        NewQueue[*cotask](),
		cap:          cap,
		corun:        NewRC(),
		coworker:     NewRC(),
		mu:           &sync.Mutex{},
		taskObjPool: NewObjectPool(func() *cotask {
			return &cotask{}
		}),
	}
}

// Go executes f.
func (cp *Copool) Go(f func()) {
	ctx := context.Background()
	cp.CtxGo(&ctx, f)
}

// CtxGo executes f and accepts the context.
func (cp *Copool) CtxGo(ctx *context.Context, f func()) {
	cp.corun.Increase()
	task := cp.taskObjPool.Get()
	task.f = func() {
		defer func() {
			if r := recover(); r != nil {
				if cp.panicHandler != nil {
					cp.panicHandler(ctx, r)
				} else {
					msg := fmt.Sprintf("[panic] copool: %v: %s", r, debug.Stack())
					fmt.Println(msg)
					os.Exit(-1)
				}
			}
		}()
		defer cp.corun.Decrease()
		f()
	}

	task.ctx = ctx

	cp.taskQ.Put(task)
	if cp.coworker.Value() == 0 || cp.taskQ.Len() != 0 && cp.coworker.Value() < int(cp.cap) {
		go func() {
			cp.coworker.Increase()
			defer cp.coworker.Decrease()

			for {
				cp.mu.Lock()
				if cp.taskQ.Len() == 0 {
					cp.mu.Unlock()
					return
				}

				task := cp.taskQ.PeakAndTake()
				cp.mu.Unlock()
				task.f()
				task.zero()
				cp.taskObjPool.Put(task)
			}

		}()
	}

}

// SetPanicHandler sets the panic handler.
func (cp *Copool) SetPanicHandler(f func(*context.Context, interface{})) *Copool {
	cp.panicHandler = f
	return cp
}

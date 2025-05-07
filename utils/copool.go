package utils

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"sync/atomic"
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
	corun        atomic.Int32
	coworker     atomic.Int32
	mu           *sync.Mutex
	taskObjPool  *ObjectPool[*cotask]
}

// NewCopool return a goroutine pool with specified cap
func NewCopool(cap uint) *Copool {
	return &Copool{
		panicHandler: nil,
		taskQ:        NewQueue[*cotask](false),
		cap:          cap,
		corun:        atomic.Int32{},
		coworker:     atomic.Int32{},
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
	cp.corun.Add(1)
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
		defer cp.corun.Add(-1)
		f()
	}

	task.ctx = ctx
	cp.mu.Lock()
	cp.taskQ.Put(task)

	if cp.coworker.Load() == 0 || cp.taskQ.Len() != 0 && uint(cp.coworker.Load()) < uint(cp.cap) {
		cp.mu.Unlock()
		cp.coworker.Add(1)

		go func() {
			defer cp.coworker.Add(-1)

			for {
				cp.mu.Lock()
				if cp.taskQ.Len() == 0 {
					cp.mu.Unlock()
					return
				}

				task := cp.taskQ.Pop()
				cp.mu.Unlock()
				task.f()
				task.zero()
				cp.taskObjPool.Put(task)
			}

		}()
	} else {
		cp.mu.Unlock()
	}
}

// SetPanicHandler sets the panic handler.
func (cp *Copool) SetPanicHandler(f func(*context.Context, interface{})) *Copool {
	cp.panicHandler = f
	return cp
}

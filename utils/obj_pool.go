package utils

import "sync"

// ObjectPool with Type
type ObjectPool[T any] struct {
	pool sync.Pool
}

func NewObjectPool[T any](creator func() T) *ObjectPool[T] {
	return &ObjectPool[T]{
		pool: sync.Pool{
			New: func() any { return creator() },
		},
	}
}

func (p *ObjectPool[T]) Get() T {
	return p.pool.Get().(T)
}

func (p *ObjectPool[T]) Put(x T) {
	p.pool.Put(x)
}

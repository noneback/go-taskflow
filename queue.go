package gotaskflow

import (
	"sync"

	"github.com/eapache/queue/v2"
)

// thread safe
type Queue[T any] struct {
	q       *queue.Queue[T]
	rwMutex *sync.RWMutex
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		q:       queue.New[T](),
		rwMutex: &sync.RWMutex{},
	}
}

func (q *Queue[T]) Peak() T {
	q.rwMutex.Lock()
	defer q.rwMutex.Unlock()
	return q.q.Peek()
}

func (q *Queue[T]) Len() int32 {
	q.rwMutex.RLock()
	defer q.rwMutex.RUnlock()
	return int32(q.q.Length())
}

func (q *Queue[T]) Put(data T) {
	q.rwMutex.Lock()
	defer q.rwMutex.Unlock()
	q.q.Add(data)
}

func (q *Queue[T]) PeakAndTake() T {
	q.rwMutex.Lock()
	defer q.rwMutex.Unlock()
	return q.q.Remove()
}

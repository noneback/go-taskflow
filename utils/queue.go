package utils

import (
	"sync"

	"github.com/eapache/queue/v2"
)

// thread safe Queue
type Queue[T any] struct {
	q  *queue.Queue[T]
	mu *sync.Mutex
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		q:  queue.New[T](),
		mu: &sync.Mutex{},
	}
}

func (q *Queue[T]) Peak() T {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.q.Peek()
}

func (q *Queue[T]) Len() int32 {
	q.mu.Lock()
	defer q.mu.Unlock()
	return int32(q.q.Length())
}

func (q *Queue[T]) Put(data T) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.q.Add(data)
}

func (q *Queue[T]) PeakAndTake() T {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.q.Remove()
}

// NOTE: CODE BASE IS COPIED FROM https://github.com/eapache/queue/blob/main/v2/queue.go, modified to make it thread safe

package utils

import (
	"sync"
)

// minQueueLen is smallest capacity that queue may have.
// Must be power of 2 for bitwise modulus: x % n == x & (n - 1).
const minQueueLen = 16

// Queue represents a single instance of the queue data structure.
type Queue[V any] struct {
	buf               []*V
	head, tail, count int
	rw                *sync.RWMutex
	tsafe             bool
}

// New constructs and returns a new Queue.
func NewQueue[V any](threadSafe bool) *Queue[V] {
	return &Queue[V]{
		buf:   make([]*V, minQueueLen),
		rw:    &sync.RWMutex{},
		tsafe: threadSafe,
	}
}

// Length returns the number of elements currently stored in the queue.
func (q *Queue[V]) Len() int {
	if q.tsafe {
		q.rw.RLock()
		defer q.rw.RUnlock()
	}

	return q.count
}

// resizes the queue to fit exactly twice its current contents
// this can result in shrinking if the queue is less than half-full
func (q *Queue[V]) resize() {
	newBuf := make([]*V, q.count<<1)

	if q.tail > q.head {
		copy(newBuf, q.buf[q.head:q.tail])
	} else {
		n := copy(newBuf, q.buf[q.head:])
		copy(newBuf[n:], q.buf[:q.tail])
	}

	q.head = 0
	q.tail = q.count
	q.buf = newBuf
}

// Add puts an element on the end of the queue.
func (q *Queue[V]) Put(elem V) {
	if q.tsafe {
		q.rw.Lock()
		defer q.rw.Unlock()
	}

	if q.count == len(q.buf) {
		q.resize()
	}

	q.buf[q.tail] = &elem
	// bitwise modulus
	q.tail = (q.tail + 1) & (len(q.buf) - 1)
	q.count++
}

// Top returns the element at the head of the queue. This call panics
// if the queue is empty.
func (q *Queue[V]) Top() V {
	if q.tsafe {
		q.rw.RLock()
		defer q.rw.RUnlock()
	}

	if q.count <= 0 {
		panic("queue: Peek() called on empty queue")
	}
	return *(q.buf[q.head])
}

// Get returns the element at index i in the queue. If the index is
// invalid, the call will panic. This method accepts both positive and
// negative index values. Index 0 refers to the first element, and
// index -1 refers to the last.
func (q *Queue[V]) Get(i int) V {
	if q.tsafe {
		q.rw.RLock()
		defer q.rw.RUnlock()
	}

	// If indexing backwards, convert to positive index.
	if i < 0 {
		i += q.count
	}
	if i < 0 || i >= q.count {
		panic("queue: Get() called with index out of range")
	}
	// bitwise modulus
	return *(q.buf[(q.head+i)&(len(q.buf)-1)])
}

// Remove removes and returns the element from the front of the queue. If the
// queue is empty, the call will panic.
func (q *Queue[V]) Pop() V {
	if q.tsafe {
		q.rw.Lock()
		defer q.rw.Unlock()
	}

	if q.count <= 0 {
		panic("queue: Remove() called on empty queue")
	}
	ret := q.buf[q.head]
	q.buf[q.head] = nil
	// bitwise modulus
	q.head = (q.head + 1) & (len(q.buf) - 1)
	q.count--
	// Resize down if buffer 1/4 full.
	if len(q.buf) > minQueueLen && (q.count<<2) == len(q.buf) {
		q.resize()
	}
	return *ret
}

// Remove removes and returns the element from the front of the queue. If the
// queue is empty, the call will panic.
func (q *Queue[V]) TryPop() (V, bool) {
	if q.tsafe {
		q.rw.Lock()
		defer q.rw.Unlock()
	}

	if q.count <= 0 {
		var tmp V
		return tmp, false
	}
	ret := q.buf[q.head]
	q.buf[q.head] = nil
	// bitwise modulus
	q.head = (q.head + 1) & (len(q.buf) - 1)
	q.count--
	// Resize down if buffer 1/4 full.
	if len(q.buf) > minQueueLen && (q.count<<2) == len(q.buf) {
		q.resize()
	}

	return *ret, true
}

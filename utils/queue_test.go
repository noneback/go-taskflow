/*
NOTE: CODE BASE IS COPYED FROM https://github.com/eapache/queue/blob/main/v2/queue.go, modified to make it thread safe

Package queue provides a fast, ring-buffer queue based on the version suggested by Dariusz Górecki.
Using this instead of other, simpler, queue implementations (slice+append or linked list) provides
substantial memory and time benefits, and fewer GC pauses.

The queue implemented here is as fast as it is for an additional reason: it is *not* thread-safe.
*/

package utils

import "testing"

func TestQueueSimple(t *testing.T) {
	q := NewQueue[int]()

	for i := 0; i < minQueueLen; i++ {
		q.Put(i)
	}
	for i := 0; i < minQueueLen; i++ {
		if q.Top() != i {
			t.Error("peek", i, "had value", q.Top())
		}
		x := q.Pop()
		if x != i {
			t.Error("remove", i, "had value", x)
		}
	}
}

func TestQueueWrapping(t *testing.T) {
	q := NewQueue[int]()

	for i := 0; i < minQueueLen; i++ {
		q.Put(i)
	}
	for i := 0; i < 3; i++ {
		q.Pop()
		q.Put(minQueueLen + i)
	}

	for i := 0; i < minQueueLen; i++ {
		if q.Top() != i+3 {
			t.Error("peek", i, "had value", q.Top())
		}
		q.Pop()
	}
}

func TestQueueLen(t *testing.T) {
	q := NewQueue[int]()

	if q.Len() != 0 {
		t.Error("empty queue length not 0")
	}

	for i := 0; i < 1000; i++ {
		q.Put(i)
		if q.Len() != i+1 {
			t.Error("adding: queue with", i, "elements has length", q.Len())
		}
	}
	for i := 0; i < 1000; i++ {
		q.Pop()
		if q.Len() != 1000-i-1 {
			t.Error("removing: queue with", 1000-i-i, "elements has length", q.Len())
		}
	}
}

func TestQueueGet(t *testing.T) {
	q := NewQueue[int]()

	for i := 0; i < 1000; i++ {
		q.Put(i)
		for j := 0; j < q.Len(); j++ {
			if q.Get(j) != j {
				t.Errorf("index %d doesn't contain %d", j, j)
			}
		}
	}
}

func TestQueueGetNegative(t *testing.T) {
	q := NewQueue[int]()

	for i := 0; i < 1000; i++ {
		q.Put(i)
		for j := 1; j <= q.Len(); j++ {
			if q.Get(-j) != q.Len()-j {
				t.Errorf("index %d doesn't contain %d", -j, q.Len()-j)
			}
		}
	}
}

func TestQueueGetOutOfRangePanics(t *testing.T) {
	q := NewQueue[int]()

	q.Put(1)
	q.Put(2)
	q.Put(3)

	AssertPanics(t, "should panic when negative index", func() {
		q.Get(-4)
	})

	AssertPanics(t, "should panic when index greater than length", func() {
		q.Get(4)
	})
}

func TestQueuePeekOutOfRangePanics(t *testing.T) {
	q := NewQueue[any]()

	AssertPanics(t, "should panic when peeking empty queue", func() {
		q.Top()
	})

	q.Put(1)
	q.Pop()

	AssertPanics(t, "should panic when peeking emptied queue", func() {
		q.Top()
	})
}

func TestQueuePopOutOfRangePanics(t *testing.T) {
	q := NewQueue[int]()

	AssertPanics(t, "should panic when removing empty queue", func() {
		q.Pop()
	})

	q.Put(1)
	q.Pop()

	AssertPanics(t, "should panic when removing emptied queue", func() {
		q.Pop()
	})
}


// WARNING: Go's benchmark utility (go test -bench .) increases the number of
// iterations until the benchmarks take a reasonable amount of time to run; memory usage
// is *NOT* considered. On a fast CPU, these benchmarks can fill hundreds of GB of memory
// (and then hang when they start to swap). You can manually control the number of iterations
// with the `-benchtime` argument. Passing `-benchtime 1000000x` seems to be about right.

func BenchmarkQueueSerial(b *testing.B) {
	q := NewQueue[any]()
	for i := 0; i < b.N; i++ {
		q.Put(nil)
	}
	for i := 0; i < b.N; i++ {
		q.Top()
		q.Pop()
	}
}

func BenchmarkQueueGet(b *testing.B) {
	q := NewQueue[int]()
	for i := 0; i < b.N; i++ {
		q.Put(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Get(i)
	}
}

func BenchmarkQueueTickTock(b *testing.B) {
	q := NewQueue[any]()
	for i := 0; i < b.N; i++ {
		q.Put(nil)
		q.Top()
		q.Pop()
	}
}

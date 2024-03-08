package gotaskflow

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const benchmarkTimes = 10000

func DoCopyStack(a, b int) int {
	if b < 100 {
		return DoCopyStack(0, b+1)
	}
	return 0
}

func testFunc() {
	DoCopyStack(0, 0)
}

func TestPool(t *testing.T) {
	p := NewTaskPool(100)
	var n int32
	var wg sync.WaitGroup
	for i := 0; i < 2000; i++ {
		// fmt.Print(i)
		wg.Add(1)
		p.Go(func() {
			defer wg.Done()
			atomic.AddInt32(&n, 1)
		})
	}
	wg.Wait()
	if n != 2000 {
		t.Error(n)
	}
}

func testPanic() {
	n := 0
	fmt.Println(1 / n)
}

func TestPoolPanic(t *testing.T) {
	p := NewTaskPool(100)
	var wg sync.WaitGroup
	p.Go(testPanic)
	wg.Wait()
	time.Sleep(time.Second)
}

func BenchmarkPool(b *testing.B) {
	p := NewTaskPool(100)
	var wg sync.WaitGroup
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(benchmarkTimes)
		for j := 0; j < benchmarkTimes; j++ {
			p.Go(func() {
				testFunc()
				wg.Done()
			})
		}
		wg.Wait()
	}
}

func BenchmarkGo(b *testing.B) {
	var wg sync.WaitGroup
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(benchmarkTimes)
		for j := 0; j < benchmarkTimes; j++ {
			go func() {
				testFunc()
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

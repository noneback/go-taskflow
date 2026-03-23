package gotaskflow_test

import (
	"fmt"
	_ "net/http/pprof"
	"os"
	"sync/atomic"
	"testing"
	"time"

	gotaskflow "github.com/noneback/go-taskflow"
	"github.com/noneback/go-taskflow/utils"
)

// =============================================================================
// Test Helper Functions
// =============================================================================

// testTaskSimple creates a simple task function that logs execution.
func testTaskSimple(name string, t *testing.T) func() {
	return func() {
		t.Logf("Executing task: %s", name)
	}
}

// =============================================================================
// Basic TaskFlow Tests
// =============================================================================

func TestTaskFlow(t *testing.T) {
	// A -> B, C -> B, A1 -> B; A1 -> C, B1 -> C; C1 is independent
	executor := gotaskflow.NewExecutor(10)
	tf := gotaskflow.NewTaskFlow("G")

	A := tf.NewTask("A", testTaskSimple("A", t))
	B := tf.NewTask("B", testTaskSimple("B", t))
	C := tf.NewTask("C", testTaskSimple("C", t))
	A1 := tf.NewTask("A1", testTaskSimple("A1", t))
	B1 := tf.NewTask("B1", testTaskSimple("B1", t))
	_ = tf.NewTask("C1", testTaskSimple("C1", t))

	A.Precede(B)
	C.Precede(B)
	A1.Precede(B)
	C.Succeed(A1)
	C.Succeed(B1)

	t.Run("TestViz", func(t *testing.T) {
		if err := tf.Dump(os.Stdout); err != nil {
			t.Fatalf("Failed to dump taskflow: %v", err)
		}
	})

	executor.Run(tf).Wait()
}

func TestSubflow(t *testing.T) {
	executor := gotaskflow.NewExecutor(10)
	tf := gotaskflow.NewTaskFlow("G")

	A := tf.NewTask("A", testTaskSimple("A", t))
	B := tf.NewTask("B", testTaskSimple("B", t))
	C := tf.NewTask("C", testTaskSimple("C", t))
	A1 := tf.NewTask("A1", testTaskSimple("A1", t))
	B1 := tf.NewTask("B1", testTaskSimple("B1", t))
	C1 := tf.NewTask("C1", testTaskSimple("C1", t))

	A.Precede(B)
	C.Precede(B)
	C.Succeed(A1)
	C.Succeed(B1)

	subflow := tf.NewSubflow("sub1", func(sf *gotaskflow.Subflow) {
		A2 := sf.NewTask("A2", testTaskSimple("A2", t))
		B2 := sf.NewTask("B2", testTaskSimple("B2", t))
		C2 := sf.NewTask("C2", testTaskSimple("C2", t))
		A2.Precede(B2)
		C2.Precede(B2)

		cond := sf.NewCondition("cond", func() uint {
			return 0
		})

		ssub := sf.NewSubflow("sub in sub", func(sf *gotaskflow.Subflow) {
			sf.NewTask("done", func() {
				t.Log("done in nested subflow")
			})
		})

		cond.Precede(ssub, cond)
	})

	subflow2 := tf.NewSubflow("sub2", func(sf *gotaskflow.Subflow) {
		A3 := sf.NewTask("A3", testTaskSimple("A3", t))
		B3 := sf.NewTask("B3", testTaskSimple("B3", t))
		C3 := sf.NewTask("C3", testTaskSimple("C3", t))
		A3.Precede(B3)
		C3.Precede(B3)
	})

	subflow.Precede(subflow2)
	C1.Precede(subflow)
	C1.Succeed(C)

	executor.Run(tf).Wait()

	if err := tf.Dump(os.Stdout); err != nil {
		t.Errorf("Failed to dump: %v", err)
	}
}

// =============================================================================
// Error Robustness Tests
// =============================================================================

func TestTaskflowPanic(t *testing.T) {
	executor := gotaskflow.NewExecutor(10)
	tf := gotaskflow.NewTaskFlow("G")

	A := tf.NewTask("A", testTaskSimple("A", t))
	B := tf.NewTask("B", testTaskSimple("B", t))
	C := tf.NewTask("C", func() {
		t.Log("C")
		panic("panic C")
	})
	A.Precede(B)
	C.Precede(B)

	executor.Run(tf).Wait()
	// Test should not hang or crash
}

func TestSubflowPanic(t *testing.T) {
	executor := gotaskflow.NewExecutor(10)
	tf := gotaskflow.NewTaskFlow("G")

	A := tf.NewTask("A", testTaskSimple("A", t))
	B := tf.NewTask("B", testTaskSimple("B", t))
	C := tf.NewTask("C", testTaskSimple("C", t))
	A.Precede(B)
	C.Precede(B)

	subflow := tf.NewSubflow("sub1", func(sf *gotaskflow.Subflow) {
		A2 := sf.NewTask("A2", func() {
			t.Log("A2")
			time.Sleep(1 * time.Second)
		})
		B2 := sf.NewTask("B2", testTaskSimple("B2", t))
		C2 := sf.NewTask("C2", func() {
			t.Log("C2")
			panic("C2 panicked")
		})
		A2.Precede(B2)
		panic("subflow panic")
		C2.Precede(B2)
	})

	subflow.Precede(B)
	executor.Run(tf).Wait()

	if err := tf.Dump(os.Stdout); err != nil {
		t.Logf("Dump error: %v", err)
	}
}

// =============================================================================
// Condition Tests
// =============================================================================

func TestTaskflowCondition(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		executor := gotaskflow.NewExecutor(10)
		tf := gotaskflow.NewTaskFlow("G")

		A := tf.NewTask("A", testTaskSimple("A", t))
		B := tf.NewTask("B", testTaskSimple("B", t))
		C := tf.NewTask("C", testTaskSimple("C", t))
		A.Precede(B)
		C.Precede(B)

		fail := tf.NewTask("failed", func() {
			t.Log("Failed - should not execute")
			t.Fail()
		})
		success := tf.NewTask("success", testTaskSimple("success", t))

		cond := tf.NewCondition("cond", func() uint { return 0 })
		B.Precede(cond)
		cond.Precede(success, fail)

		suc := tf.NewSubflow("sub1", func(sf *gotaskflow.Subflow) {
			A2 := sf.NewTask("A2", testTaskSimple("A2", t))
			B2 := sf.NewTask("B2", testTaskSimple("B2", t))
			C2 := sf.NewTask("C2", testTaskSimple("C2", t))
			A2.Precede(B2)
			C2.Precede(B2)
		}).Priority(gotaskflow.HIGH)
		fs := tf.NewTask("fail_single", func() { t.Log("it should be canceled") })
		fail.Precede(fs, suc)

		if err := tf.Dump(os.Stdout); err != nil {
			t.Logf("Dump error: %v", err)
		}
		executor.Run(tf).Wait()
	})

	t.Run("normal-1", func(t *testing.T) {
		executor := gotaskflow.NewExecutor(10)
		tf := gotaskflow.NewTaskFlow("G")

		A := tf.NewTask("A", testTaskSimple("A", t))
		B := tf.NewTask("B", testTaskSimple("B", t))
		C := tf.NewTask("C", testTaskSimple("C", t))
		A.Precede(B)
		C.Precede(B)

		fail := tf.NewTask("failed", func() {
			t.Log("Failed - should not execute")
			t.Fail()
		})
		success := tf.NewTask("success", testTaskSimple("success", t))

		cond := tf.NewCondition("cond", func() uint { return 0 })
		B.Precede(cond)
		cond.Precede(success)
		cond.Precede(fail)

		executor.Run(tf).Wait()
	})

	t.Run("multiple tasks preceding condition", func(t *testing.T) {
		executor := gotaskflow.NewExecutor(10)
		tf := gotaskflow.NewTaskFlow("G")

		A := tf.NewTask("A", testTaskSimple("A", t))
		B := tf.NewTask("B", testTaskSimple("B", t))
		C := tf.NewTask("C", testTaskSimple("C", t))

		fail := tf.NewTask("failed", func() {
			t.Log("Failed - should not execute")
			t.Fail()
		})
		success := tf.NewTask("success", testTaskSimple("success", t))

		cond := tf.NewCondition("cond", func() uint { return 0 })
		A.Precede(cond)
		B.Precede(cond)
		C.Precede(cond)
		cond.Precede(success)
		cond.Precede(fail)

		executor.Run(tf).Wait()
	})

	t.Run("start with condition node", func(t *testing.T) {
		i := 0
		tf := gotaskflow.NewTaskFlow("G")

		cond := tf.NewCondition("cond", func() uint {
			if i == 0 {
				return 0
			} else {
				return 1
			}
		})

		zero := tf.NewTask("zero", testTaskSimple("zero", t))
		one := tf.NewTask("one", testTaskSimple("one", t))
		cond.Precede(zero, one)

		executor := gotaskflow.NewExecutor(10)
		executor.Run(tf).Wait()

		if err := tf.Dump(os.Stdout); err != nil {
			t.Errorf("Dump error: %v", err)
		}
	})
}

// =============================================================================
// Loop Tests
// =============================================================================

func TestTaskflowLoop(t *testing.T) {
	executor := gotaskflow.NewExecutor(10)

	t.Run("normal", func(t *testing.T) {
		i := 0
		tf := gotaskflow.NewTaskFlow("G")

		init := tf.NewTask("init", func() {
			i = 0
			t.Log("i=0")
		})
		cond := tf.NewCondition("while i < 5", func() uint {
			if i < 5 {
				return 0
			} else {
				return 1
			}
		})
		body := tf.NewTask("body", func() {
			i += 1
			t.Logf("i++ = %d", i)
		})
		back := tf.NewCondition("back", func() uint {
			t.Log("back")
			return 0
		})
		done := tf.NewTask("done", func() {
			t.Log("done")
		})

		init.Precede(cond)
		cond.Precede(body, done)
		body.Precede(back)
		back.Precede(cond)

		executor.Run(tf).Wait()

		if i < 5 {
			t.Errorf("Expected i >= 5, got %d", i)
		}
	})

	t.Run("simple loop", func(t *testing.T) {
		i := 0
		tf := gotaskflow.NewTaskFlow("G")

		init := tf.NewTask("init", func() {
			i = 0
		})
		cond := tf.NewCondition("cond", func() uint {
			i++
			t.Logf("i++ = %d", i)
			if i > 2 {
				return 0
			} else {
				return 1
			}
		})

		done := tf.NewTask("done", func() {
			t.Log("done")
		})

		init.Precede(cond)
		cond.Precede(done, cond)

		executor.Run(tf).Wait()

		if i <= 2 {
			t.Errorf("Expected i > 2, got %d", i)
		}
	})
}

// =============================================================================
// Priority Tests
// =============================================================================

func TestTaskflowPriority(t *testing.T) {
	executor := gotaskflow.NewExecutor(1)
	q := utils.NewQueue[byte](true)
	tf := gotaskflow.NewTaskFlow("G")

	tf.NewTask("B", func() {
		t.Log("B")
		q.Put('B')
	}).Priority(gotaskflow.NORMAL)

	tf.NewTask("C", func() {
		t.Log("C")
		q.Put('C')
	}).Priority(gotaskflow.HIGH)

	A := tf.NewTask("A", func() {
		t.Log("A")
		q.Put('A')
	}).Priority(gotaskflow.LOW)

	A.Precede(
		tf.NewTask("A2", func() {
			t.Log("A2")
			q.Put('a')
		}).Priority(gotaskflow.LOW),
		tf.NewTask("B2", func() {
			t.Log("B2")
			q.Put('b')
		}).Priority(gotaskflow.HIGH),
		tf.NewTask("C2", func() {
			t.Log("C2")
			q.Put('c')
		}).Priority(gotaskflow.NORMAL),
	)

	executor.Run(tf).Wait()

	expected := []byte{'C', 'B', 'A', 'b', 'c', 'a'}
	for _, val := range expected {
		real := q.Pop()
		t.Logf("Expected: %c, Got: %c", val, real)
		if val != real {
			t.Fatalf("Priority order mismatch: expected %c, got %c", val, real)
		}
	}
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func TestTaskflowNotInFlow(t *testing.T) {
	executor := gotaskflow.NewExecutor(10)
	tf := gotaskflow.NewTaskFlow("tf")

	task := tf.NewTask("init", testTaskSimple("init", t))
	var cnt atomic.Int32
	for i := 0; i < 10; i++ {
		task.Precede(tf.NewTask("test", func() {
			cnt.Add(1)
		}))
	}

	executor.Run(tf).Wait()

	if cnt.Load() != 10 {
		t.Errorf("Expected cnt = 10, got %d", cnt.Load())
	}
}

func TestTaskflowFrozen(t *testing.T) {
	executor := gotaskflow.NewExecutor(10)
	tf := gotaskflow.NewTaskFlow("G")

	A := tf.NewTask("A", testTaskSimple("A", t))
	B := tf.NewTask("B", testTaskSimple("B", t))
	C := tf.NewTask("C", testTaskSimple("C", t))
	A.Precede(B)
	C.Precede(B)

	executor.Run(tf).Wait()

	utils.AssertPanics(t, "frozen", func() {
		tf.NewTask("tt", testTaskSimple("tt", t))
	})
}

// =============================================================================
// Stress Tests
// =============================================================================

func TestLoopRunManyTimes(t *testing.T) {
	executor := gotaskflow.NewExecutor(10)
	tf := gotaskflow.NewTaskFlow("G")
	var count atomic.Int32

	add := func(name string) func() {
		return func() {
			t.Logf("%s", name)
			count.Add(1)
		}
	}

	A := tf.NewTask("A", add("A"))
	B := tf.NewTask("B", add("B"))
	C := tf.NewTask("C", add("C"))
	A.Precede(B)
	C.Precede(B)

	// Reduced iterations for faster testing
	iterations := 100

	t.Run("static", func(t *testing.T) {
		for i := 0; i < iterations; i++ {
			if cnt := count.Load(); cnt%3 != 0 {
				t.Errorf("static unexpected count %d at iteration %d", cnt, i)
				return
			}
			executor.Run(tf).Wait()
		}
	})

	tf.Reset()
	count.Store(0)

	sf := tf.NewSubflow("sub", func(sf *gotaskflow.Subflow) {
		t.Log("sub")
		A1 := sf.NewTask("A1", add("A1"))
		B1 := sf.NewTask("B1", add("B1"))
		C1 := sf.NewTask("C1", add("C1"))
		A1.Precede(B1)
		C1.Precede(B1)
	})
	additional := tf.NewTask("additional", add("Additional"))
	B.Precede(sf)
	additional.Precede(sf)

	t.Run("subflow", func(t *testing.T) {
		for i := 0; i < iterations; i++ {
			if cnt := count.Load(); cnt%7 != 0 {
				t.Errorf("subflow unexpected count %d at iteration %d", cnt, i)
				return
			}
			executor.Run(tf).Wait()
		}
	})

	tf.Reset()
	count.Store(0)

	cond := tf.NewCondition("if count %10 % 7 == 0", func() uint {
		if count.Load()%7 == 0 {
			return 0
		} else {
			return 1
		}
	})
	plus7 := tf.NewTask("7 plus 7", func() {
		count.Add(7)
	})
	cond.Precede(plus7, tf.NewTask("7 minus 3", func() {
		t.Logf("%d", count.Load())
		t.Log("should not minus 3")
	}))
	sf.Precede(cond)

	t.Run("condition", func(t *testing.T) {
		for i := 0; i < iterations; i++ {
			if cnt := count.Load(); cnt%7 != 0 {
				t.Errorf("condition unexpected count %d at iteration %d", cnt, i)
				return
			}
			executor.Run(tf).Wait()
		}
	})

	tf.Reset()
	count.Store(0)
	cond2 := tf.NewCondition("bigger than 10000", func() uint {
		if count.Load() > 10000 {
			return 1
		} else {
			return 0
		}
	})

	plus7.Precede(cond2)
	done := tf.NewTask("done", func() {
		if count.Load() < 10000 {
			t.Fail()
		}
	})

	newPlus7 := tf.NewTask("new plus 7", func() {
		count.Add(7)
	})
	cond2.Precede(newPlus7, done)
	newPlus7.Precede(cond2)

	t.Run("loop", func(t *testing.T) {
		for i := 0; i < iterations; i++ {
			if cnt := count.Load(); cnt%7 != 0 {
				t.Errorf("loop unexpected count %d at iteration %d", cnt, i)
				return
			}
			executor.Run(tf).Wait()
		}
	})
}

func TestSequencialTaskingPanic(t *testing.T) {
	executor := gotaskflow.NewExecutor(1)
	tf := gotaskflow.NewTaskFlow("test")
	q := utils.NewQueue[string](true)

	tf.NewTask("task1", func() {
		q.Put("panic")
		t.Log("task1")
		panic(1)
	})
	tf.NewTask("task2", func() {
		q.Put("2")
		t.Log("task2")
	})
	tf.NewTask("task3", func() {
		q.Put("3")
		t.Log("task3")
	})

	executor.Run(tf).Wait()

	if q.Top() != "panic" {
		t.Error("Expected panic task to execute first")
	}
}

func TestDeadlock(t *testing.T) {
	// BUG: https://github.com/noneback/go-taskflow/issues/99
	executor := gotaskflow.NewExecutor(1)
	N := 10 // Reduced for faster testing

	t.Run("linear chain", func(t *testing.T) {
		tf := gotaskflow.NewTaskFlow("G1")
		prev := tf.NewTask("N0", func() {})
		for i := 1; i < 32; i++ {
			next := tf.NewTask(fmt.Sprintf("N%d", i), func() {})
			prev.Precede(next)
			prev = next
		}

		for i := 0; i < N; i++ {
			executor.Run(tf).Wait()
		}
	})

	t.Run("layered graph", func(t *testing.T) {
		tf := gotaskflow.NewTaskFlow("G2")
		layersCount := 8
		layerNodesCount := 8

		var curLayer, upperLayer []*gotaskflow.Task

		for i := 0; i < layersCount; i++ {
			for j := 0; j < layerNodesCount; j++ {
				task := tf.NewTask(fmt.Sprintf("N%d", i*layersCount+j), func() {})

				for _, t := range upperLayer {
					t.Precede(task)
				}

				curLayer = append(curLayer, task)
			}

			upperLayer = curLayer
			curLayer = []*gotaskflow.Task{}
		}

		for i := 0; i < N; i++ {
			executor.Run(tf).Wait()
		}
	})
}

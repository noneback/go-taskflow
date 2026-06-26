package benchmark

import (
	"fmt"
	"runtime"
	"testing"

	gotaskflow "github.com/noneback/go-taskflow"
)

// --- Topology scaling: measure scheduling overhead across graph shapes and sizes ---

func BenchmarkConcurrent(b *testing.B) {
	for _, n := range []int{8, 32, 128, 512} {
		b.Run(fmt.Sprintf("N%d", n), func(b *testing.B) {
			exec := gotaskflow.NewExecutor(uint(runtime.NumCPU()))
			tf := gotaskflow.NewTaskFlow("concurrent")
			for i := 0; i < n; i++ {
				tf.NewTask(fmt.Sprintf("T%d", i), func() {})
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				exec.Run(tf).Wait()
			}
		})
	}
}

func BenchmarkSerial(b *testing.B) {
	for _, n := range []int{8, 32, 128, 512} {
		b.Run(fmt.Sprintf("N%d", n), func(b *testing.B) {
			exec := gotaskflow.NewExecutor(uint(runtime.NumCPU()))
			tf := gotaskflow.NewTaskFlow("serial")
			prev := tf.NewTask("T0", func() {})
			for i := 1; i < n; i++ {
				next := tf.NewTask(fmt.Sprintf("T%d", i), func() {})
				prev.Precede(next)
				prev = next
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				exec.Run(tf).Wait()
			}
		})
	}
}

func BenchmarkDiamond(b *testing.B) {
	exec := gotaskflow.NewExecutor(uint(runtime.NumCPU()))
	tf := gotaskflow.NewTaskFlow("diamond")
	source := tf.NewTask("source", func() {})
	left := tf.NewTask("left", func() {})
	right := tf.NewTask("right", func() {})
	leftMid := tf.NewTask("left_mid", func() {})
	rightMid := tf.NewTask("right_mid", func() {})
	sink := tf.NewTask("sink", func() {})
	source.Precede(left, right)
	left.Precede(leftMid)
	right.Precede(rightMid)
	sink.Succeed(leftMid, rightMid)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exec.Run(tf).Wait()
	}
}

func BenchmarkDenseLayers(b *testing.B) {
	for _, layers := range []int{4, 8} {
		for _, width := range []int{4, 8} {
			b.Run(fmt.Sprintf("L%dxW%d", layers, width), func(b *testing.B) {
				exec := gotaskflow.NewExecutor(uint(runtime.NumCPU()))
				tf := gotaskflow.NewTaskFlow("dense_layers")
				var curLayer, prevLayer []*gotaskflow.Task
				for l := 0; l < layers; l++ {
					for w := 0; w < width; w++ {
						task := tf.NewTask(fmt.Sprintf("T%d_%d", l, w), func() {})
						for _, p := range prevLayer {
							p.Precede(task)
						}
						curLayer = append(curLayer, task)
					}
					prevLayer = curLayer
					curLayer = nil
				}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					exec.Run(tf).Wait()
				}
			})
		}
	}
}

// --- Feature benchmarks: paths unique to subflow and condition nodes ---

func BenchmarkSubflow(b *testing.B) {
	exec := gotaskflow.NewExecutor(uint(runtime.NumCPU()))
	tf := gotaskflow.NewTaskFlow("subflow")
	setup := tf.NewTask("setup", func() {})
	sub := tf.NewSubflow("sub", func(sf *gotaskflow.Subflow) {
		s1 := sf.NewTask("sub_1", func() {})
		s2 := sf.NewTask("sub_2", func() {})
		s3 := sf.NewTask("sub_3", func() {})
		s1.Precede(s3)
		s2.Precede(s3)
	})
	teardown := tf.NewTask("teardown", func() {})
	setup.Precede(sub)
	sub.Precede(teardown)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exec.Run(tf).Wait()
	}
}

func BenchmarkCondition(b *testing.B) {
	exec := gotaskflow.NewExecutor(uint(runtime.NumCPU()))
	tf := gotaskflow.NewTaskFlow("condition")
	entry := tf.NewTask("entry", func() {})
	cond := tf.NewCondition("cond", func() uint { return 0 })
	branchA := tf.NewTask("branch_a", func() {})
	branchB := tf.NewTask("branch_b", func() {})
	entry.Precede(cond)
	cond.Precede(branchA, branchB)
	_ = branchA
	_ = branchB
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exec.Run(tf).Wait()
	}
}

// --- Concurrency scaling: fixed topology, varying executor concurrency ---

func BenchmarkConcurrencyScaling(b *testing.B) {
	const taskCount = 64
	numCPU := runtime.NumCPU()
	for _, c := range []int{1, numCPU, numCPU * 4} {
		b.Run(fmt.Sprintf("C%d", c), func(b *testing.B) {
			exec := gotaskflow.NewExecutor(uint(c))
			tf := gotaskflow.NewTaskFlow("conc_scaling")
			for i := 0; i < taskCount; i++ {
				tf.NewTask(fmt.Sprintf("T%d", i), func() {})
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				exec.Run(tf).Wait()
			}
		})
	}
}

// --- Graph construction: allocation cost of building DAGs ---

func BenchmarkGraphBuild(b *testing.B) {
	for _, n := range []int{32, 128, 512} {
		b.Run(fmt.Sprintf("N%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				tf := gotaskflow.NewTaskFlow("build")
				prev := tf.NewTask("T0", func() {})
				for j := 1; j < n; j++ {
					next := tf.NewTask(fmt.Sprintf("T%d", j), func() {})
					prev.Precede(next)
					prev = next
				}
			}
		})
	}
}

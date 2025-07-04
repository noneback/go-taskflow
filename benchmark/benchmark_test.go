package benchmark

import (
	"fmt"
	"testing"

	gotaskflow "github.com/noneback/go-taskflow"
)

var executor = gotaskflow.NewExecutor(1)

func BenchmarkC32(b *testing.B) {
	tf := gotaskflow.NewTaskFlow("G")
	for i := 0; i < 32; i++ {
		tf.NewTask(fmt.Sprintf("N%d", i), func() {})
	}

	for i := 0; i < b.N; i++ {
		executor.Run(tf).Wait()
	}
}

func BenchmarkS32(b *testing.B) {
	tf := gotaskflow.NewTaskFlow("G")
	prev := tf.NewTask("N0", func() {})
	for i := 1; i < 32; i++ {
		next := tf.NewTask(fmt.Sprintf("N%d", i), func() {})
		prev.Precede(next)
		prev = next
	}

	for i := 0; i < b.N; i++ {
		executor.Run(tf).Wait()
	}
}

func BenchmarkC6(b *testing.B) {
	tf := gotaskflow.NewTaskFlow("G")
	n0 := tf.NewTask("N0", func() {})
	n1 := tf.NewTask("N1", func() {})
	n2 := tf.NewTask("N2", func() {})
	n3 := tf.NewTask("N3", func() {})
	n4 := tf.NewTask("N4", func() {})
	n5 := tf.NewTask("N5", func() {})

	n0.Precede(n1, n2)
	n1.Precede(n3)
	n2.Precede(n4)
	n5.Succeed(n3, n4)

	for i := 0; i < b.N; i++ {
		executor.Run(tf).Wait()
	}
}

func BenchmarkC8x8(b *testing.B) {
	tf := gotaskflow.NewTaskFlow("G")

	layersCount := 8
	layerNodesCount := 8

	var curLayer, upperLayer []*gotaskflow.Task

	for i := 0; i < layersCount; i++ {
		for j := 0; j < layerNodesCount; j++ {
			task := tf.NewTask(fmt.Sprintf("N%d", i*layersCount+j), func() {})

			for i := range upperLayer {
				upperLayer[i].Precede(task)
			}

			curLayer = append(curLayer, task)
		}

		upperLayer = curLayer
		curLayer = []*gotaskflow.Task{}
	}

	for i := 0; i < b.N; i++ {
		executor.Run(tf).Wait()
	}
}

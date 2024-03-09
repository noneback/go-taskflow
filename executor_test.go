package gotaskflow_test

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"testing"

	gotaskflow "github.com/noneback/go-taskflow"
)

func TestExecutor(t *testing.T) {
	executor := gotaskflow.NewExecutor(runtime.NumCPU() - 1)
	tf := gotaskflow.NewTaskFlow("G")
	A, B, C :=
		gotaskflow.NewTask("A", func(ctx *context.Context) {
			fmt.Println("A")
		}),
		gotaskflow.NewTask("B", func(ctx *context.Context) {
			fmt.Println("B")
		}),
		gotaskflow.NewTask("C", func(ctx *context.Context) {
			fmt.Println("C")
		})

	A1, B1, C1 :=
		gotaskflow.NewTask("A1", func(ctx *context.Context) {
			fmt.Println("A1")
		}),
		gotaskflow.NewTask("B1", func(ctx *context.Context) {
			fmt.Println("B1")
		}),
		gotaskflow.NewTask("C1", func(ctx *context.Context) {
			fmt.Println("C1")
		})
	A.Precede(B)
	C.Precede(B)
	A1.Precede(B)
	C.Succeed(A1)
	C.Succeed(B1)

	tf.Push(A, B, C)
	tf.Push(A1, B1, C1)

	if err := tf.Visualize(os.Stdout); err != nil {
		panic(err)
	}
	executor.Run(tf)
	executor.Wait()
}

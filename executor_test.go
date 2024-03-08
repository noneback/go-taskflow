package gotaskflow_test

import (
	"context"
	"fmt"
	"io/fs"
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
	f, err := os.OpenFile("out.dot", os.O_RDWR|os.O_CREATE, fs.FileMode(os.O_TRUNC))
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := tf.Visualize(f); err != nil {
		panic(err)
	}
	executor.Run(tf)
	executor.Wait()
}

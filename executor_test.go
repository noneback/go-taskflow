package gotaskflow_test

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	gotaskflow "github.com/noneback/go-taskflow"
)

func TestExecutor(t *testing.T) {
	executor := gotaskflow.NewExecutor(uint(runtime.NumCPU()))
	tf := gotaskflow.NewTaskFlow("G")
	A, B, C :=
		gotaskflow.NewTask("A", func() {
			fmt.Println("A")
		}),
		gotaskflow.NewTask("B", func() {
			fmt.Println("B")
		}),
		gotaskflow.NewTask("C", func() {
			fmt.Println("C")
		})

	A1, B1, C1 :=
		gotaskflow.NewTask("A1", func() {
			fmt.Println("A1")
		}),
		gotaskflow.NewTask("B1", func() {
			fmt.Println("B1")
		}),
		gotaskflow.NewTask("C1", func() {
			fmt.Println("C1")
		})
	A.Precede(B)
	C.Precede(B)
	A1.Precede(B)
	C.Succeed(A1)
	C.Succeed(B1)

	tf.Push(A, B, C)
	tf.Push(A1, B1, C1)

	executor.Run(tf).Wait()
	executor.Profile(os.Stdout)
}

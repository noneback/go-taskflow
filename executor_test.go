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
		tf.NewTask("A", func() {
			fmt.Println("A")
		}),
		tf.NewTask("B", func() {
			fmt.Println("B")
		}),
		tf.NewTask("C", func() {
			fmt.Println("C")
		})

	A1, B1, _ :=
		tf.NewTask("A1", func() {
			fmt.Println("A1")
		}),
		tf.NewTask("B1", func() {
			fmt.Println("B1")
		}),
		tf.NewTask("C1", func() {
			fmt.Println("C1")
		})
	A.Precede(B)
	C.Precede(B)
	A1.Precede(B)
	C.Succeed(A1)
	C.Succeed(B1)

	executor.Run(tf).Wait()
	executor.Profile(os.Stdout)
}

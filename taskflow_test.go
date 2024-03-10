package gotaskflow_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	gotaskflow "github.com/noneback/go-taskflow"
)

func TestTaskFlow(t *testing.T) {
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

	tf := gotaskflow.NewTaskFlow("G")
	tf.Push(A, B, C)
	tf.Push(A1, B1, C1)

	t.Run("TestViz", func(t *testing.T) {
		if err := tf.Visualize(os.Stdout); err != nil {
			panic(err)
		}
	})
}

func TestConditionedTaskFlow(t *testing.T) {
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

	tf := gotaskflow.NewTaskFlow("G")
	tf.Push(A, B, C)
	tf.Push(A1, B1, C1)

	A2, B2, C2 :=
		gotaskflow.NewTask("A2", func(ctx *context.Context) {
			fmt.Println("A2")
		}),
		gotaskflow.NewTask("B2", func(ctx *context.Context) {
			fmt.Println("B2")
		}),
		gotaskflow.NewTask("C2", func(ctx *context.Context) {
			fmt.Println("C2")
		})

	i := 0
	cond := gotaskflow.NewConditionTask("cond1", func(ctx *context.Context) int {
		return i % 3
	})

	cond.SetMapper(map[int]gotaskflow.TaskInterface{
		1: A2,
		2: B2,
		3: C2,
	})

	tf.Push(cond)
	if err := tf.Visualize(os.Stdout); err != nil {
		panic(err)
	}

}

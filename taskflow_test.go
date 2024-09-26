package gotaskflow_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"testing"

	"github.com/felixge/fgprof"

	gotaskflow "github.com/noneback/go-taskflow"
)

var exector = gotaskflow.NewExecutor(10)

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
		if err := gotaskflow.Visualizer.Visualize(tf, os.Stdout); err != nil {
			panic(err)
		}
	})
	err := exector.Run(tf)
	if err != nil {
		panic(err)
	}
}

func TestSubflow(t *testing.T) {
	http.DefaultServeMux.Handle("/debug/fgprof", fgprof.Handler())

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

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

	subflow := gotaskflow.NewSubflow("sub1", func(sf *gotaskflow.Subflow) {
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
		A2.Precede(B2)
		C2.Precede(B2)
		sf.Push(A2, B2, C2)
	})

	subflow2 := gotaskflow.NewSubflow("sub2", func(sf *gotaskflow.Subflow) {
		A3, B3, C3 :=
			gotaskflow.NewTask("A3", func(ctx *context.Context) {
				fmt.Println("A3")
			}),
			gotaskflow.NewTask("B3", func(ctx *context.Context) {
				fmt.Println("B3")
			}),
			gotaskflow.NewTask("C3", func(ctx *context.Context) {
				fmt.Println("C3")
				// time.Sleep(10 * time.Second)
			})
		A3.Precede(B3)
		C3.Precede(B3)
		sf.Push(A3, B3, C3)
	})

	subflow.Precede(B)
	subflow.Precede(subflow2)

	tf := gotaskflow.NewTaskFlow("G")
	tf.Push(A, B, C)
	tf.Push(A1, B1, C1, subflow, subflow2)
	exector.Run(tf)
	exector.Wait()
	if err := gotaskflow.Visualizer.Visualize(tf, os.Stdout); err != nil {
		log.Fatal(err)
	}
	// tf.Reset()
	// exector.Run(tf)
	// exector.Wait()

	// if err := tf.Visualize(os.Stdout); err != nil {
	// 	panic(err)
	// }
}

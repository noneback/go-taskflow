package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	gotaskflow "github.com/noneback/go-taskflow"
)

func main() {
	executor := gotaskflow.NewExecutor(uint(runtime.NumCPU()-1) * 10000)
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

	subflow := tf.NewSubflow("sub1", func(sf *gotaskflow.Subflow) {
		A2, B2, C2 :=
			sf.NewTask("A2", func() {
				fmt.Println("A2")
			}),
			sf.NewTask("B2", func() {
				fmt.Println("B2")
			}),
			sf.NewTask("C2", func() {
				fmt.Println("C2")
			})
		A2.Precede(B2)
		C2.Precede(B2)

	})

	subflow2 := tf.NewSubflow("sub2", func(sf *gotaskflow.Subflow) {
		A3, B3, C3 :=
			tf.NewTask("A3", func() {
				fmt.Println("A3")
			}),
			tf.NewTask("B3", func() {
				fmt.Println("B3")
			}),
			tf.NewTask("C3", func() {
				fmt.Println("C3")
				// time.Sleep(10 * time.Second)
			})
		A3.Precede(B3)
		C3.Precede(B3)
	})

	cond := tf.NewCondition("binary", func() uint {
		return uint(time.Now().Second() % 2)
	})
	B.Precede(cond)
	cond.Precede(subflow, subflow2)
	executor.Run(tf).Wait()
	fmt.Println("Print DOT")
	if err := tf.Dump(os.Stdout); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Print Flamegraph")
	if err := executor.Profile(os.Stdout); err != nil {
		log.Fatal(err)
	}
}

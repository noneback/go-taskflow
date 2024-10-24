package gotaskflow_test

import (
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"testing"
	"time"

	gotaskflow "github.com/noneback/go-taskflow"
)

var exector = gotaskflow.NewExecutor(10)

func TestTaskFlow(t *testing.T) {
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

	tf := gotaskflow.NewTaskFlow("G")
	tf.Push(A, B, C)
	tf.Push(A1, B1, C1)

	t.Run("TestViz", func(t *testing.T) {
		if err := gotaskflow.Visualize(tf, os.Stdout); err != nil {
			panic(err)
		}
	})

	exector.Run(tf).Wait()
	fmt.Print("########### second times")
	exector.Run(tf).Wait()
}

func TestSubflow(t *testing.T) {
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

	subflow := gotaskflow.NewSubflow("sub1", func(sf *gotaskflow.Subflow) {
		A2, B2, C2 :=
			gotaskflow.NewTask("A2", func() {
				fmt.Println("A2")
			}),
			gotaskflow.NewTask("B2", func() {
				fmt.Println("B2")
			}),
			gotaskflow.NewTask("C2", func() {
				fmt.Println("C2")
			})
		A2.Precede(B2)
		C2.Precede(B2)
		sf.Push(A2, B2, C2)
	})

	subflow2 := gotaskflow.NewSubflow("sub2", func(sf *gotaskflow.Subflow) {
		A3, B3, C3 :=
			gotaskflow.NewTask("A3", func() {
				fmt.Println("A3")
			}),
			gotaskflow.NewTask("B3", func() {
				fmt.Println("B3")
			}),
			gotaskflow.NewTask("C3", func() {
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
	if err := gotaskflow.Visualize(tf, os.Stdout); err != nil {
		log.Fatal(err)
	}
	exector.Profile(os.Stdout)
	// exector.Wait()

	// if err := tf.Visualize(os.Stdout); err != nil {
	// 	panic(err)
	// }
}

// ERROR robust testing
func TestTaskflowPanic(t *testing.T) {
	A, B, C :=
		gotaskflow.NewTask("A", func() {
			fmt.Println("A")
		}),
		gotaskflow.NewTask("B", func() {
			fmt.Println("B")
		}),
		gotaskflow.NewTask("C", func() {
			fmt.Println("C")
			panic("panic C")
		})
	A.Precede(B)
	C.Precede(B)
	tf := gotaskflow.NewTaskFlow("G")
	tf.Push(A, B, C)

	exector.Run(tf).Wait()
}

func TestSubflowPanic(t *testing.T) {
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
	A.Precede(B)
	C.Precede(B)

	subflow := gotaskflow.NewSubflow("sub1", func(sf *gotaskflow.Subflow) {
		A2, B2, C2 :=
			gotaskflow.NewTask("A2", func() {
				fmt.Println("A2")
				time.Sleep(1 * time.Second)
			}),
			gotaskflow.NewTask("B2", func() {
				fmt.Println("B2")
			}),
			gotaskflow.NewTask("C2", func() {
				fmt.Println("C2")
				panic("C2 paniced")
			})
		sf.Push(A2, B2, C2)
		A2.Precede(B2)
		panic("subflow panic")
		C2.Precede(B2)
	})

	subflow.Precede(B)

	tf := gotaskflow.NewTaskFlow("G")
	tf.Push(A, B, C)
	tf.Push(subflow)
	exector.Run(tf)
	exector.Wait()
	if err := gotaskflow.Visualize(tf, os.Stdout); err != nil {
		fmt.Errorf("%v", err)
	}
	exector.Profile(os.Stdout)
}

func TestTaskflowCondition(t *testing.T) {
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
	A.Precede(B)
	C.Precede(B)
	tf := gotaskflow.NewTaskFlow("G")
	tf.Push(A, B, C)
	fail, success := gotaskflow.NewTask("failed", func() {
		fmt.Println("Failed")
		t.Fail()
	}), gotaskflow.NewTask("success", func() {
		fmt.Println("success")
	})

	cond := gotaskflow.NewCondition("cond", func() uint { return 0 })
	B.Precede(cond)
	cond.Precede(success, fail)

	suc := gotaskflow.NewSubflow("sub1", func(sf *gotaskflow.Subflow) {
		A2, B2, C2 :=
			gotaskflow.NewTask("A2", func() {
				fmt.Println("A2")
			}),
			gotaskflow.NewTask("B2", func() {
				fmt.Println("B2")
			}),
			gotaskflow.NewTask("C2", func() {
				fmt.Println("C2")
			})
		sf.Push(A2, B2, C2)
		A2.Precede(B2)
		C2.Precede(B2)
	})
	fs := gotaskflow.NewTask("fail_single", func() {
		fmt.Println("it should be canceled")
	})
	fail.Precede(fs, suc)
	// success.Precede(suc)
	tf.Push(cond, success, fail, fs, suc)
	exector.Run(tf).Wait()

	if err := gotaskflow.Visualize(tf, os.Stdout); err != nil {
		fmt.Errorf("%v", err)
	}
	exector.Profile(os.Stdout)
}

func TestTaskflowLoop(t *testing.T) {
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
	A.Precede(B)
	C.Precede(B)
	tf := gotaskflow.NewTaskFlow("G")
	tf.Push(A, B, C)
	zero := gotaskflow.NewTask("zero", func() {
		fmt.Println("zero")
	})
	counter := uint(0)
	cond := gotaskflow.NewCondition("cond", func() uint {
		counter += 1
		return counter % 3
	})
	B.Precede(cond)
	cond.Precede(cond, cond, zero)

	tf.Push(cond, zero)
	exector.Run(tf).Wait()

	if err := gotaskflow.Visualize(tf, os.Stdout); err != nil {
		fmt.Errorf("%v", err)
	}
	exector.Profile(os.Stdout)
}

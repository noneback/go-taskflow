package gotaskflow_test

import (
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"testing"
	"time"

	gotaskflow "github.com/noneback/go-taskflow"
	"github.com/noneback/go-taskflow/utils"
)

var executor = gotaskflow.NewExecutor(10)

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

	executor.Run(tf).Wait()
	fmt.Print("########### second times")
	executor.Run(tf).Wait()
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
	executor.Run(tf)
	executor.Wait()
	if err := gotaskflow.Visualize(tf, os.Stdout); err != nil {
		log.Fatal(err)
	}
	executor.Profile(os.Stdout)
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

	executor.Run(tf).Wait()
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
	executor.Run(tf)
	executor.Wait()
	if err := gotaskflow.Visualize(tf, os.Stdout); err != nil {
		fmt.Errorf("%v", err)
	}
	executor.Profile(os.Stdout)
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
	executor.Run(tf).Wait()

	if err := gotaskflow.Visualize(tf, os.Stdout); err != nil {
		fmt.Errorf("%v", err)
	}
	executor.Profile(os.Stdout)
}

func TestTaskflowLoop(t *testing.T) {
	i := 0
	tf := gotaskflow.NewTaskFlow("G")
	init, cond, body, back, done :=
		gotaskflow.NewTask("init", func() {
			i = 0
			fmt.Println("i=0")
		}),
		gotaskflow.NewCondition("while i < 5", func() uint {
			if i < 5 {
				return 0
			} else {
				return 1
			}
		}),
		gotaskflow.NewTask("i++", func() {
			i += 1
			fmt.Println("i++ =", i)
		}),
		gotaskflow.NewCondition("back", func() uint {
			fmt.Println("back")
			return 0
		}),
		gotaskflow.NewTask("done", func() {
			fmt.Println("done")
		})

	tf.Push(init, cond, body, back, done)

	init.Precede(cond)
	cond.Precede(body, done)
	body.Precede(back)
	back.Precede(cond)

	executor.Run(tf).Wait()

	// if err := gotaskflow.Visualize(tf, os.Stdout); err != nil {
	// 	fmt.Printf("%v", err)
	// }
	// exector.Profile(os.Stdout)
}

func TestTaskflowPriority(t *testing.T) {
	exector := gotaskflow.NewExecutor(uint(2))
	q := utils.NewQueue[byte]()
	tf := gotaskflow.NewTaskFlow("G")
	B, C :=
		gotaskflow.NewTask("B", func() {
			fmt.Println("B")
			q.Put('B')
		}).Priority(gotaskflow.NORMAL),
		gotaskflow.NewTask("C", func() {
			fmt.Println("C")
			q.Put('C')
		}).Priority(gotaskflow.HIGH)
	suc := gotaskflow.NewSubflow("sub1", func(sf *gotaskflow.Subflow) {
		A2, B2, C2 :=
			gotaskflow.NewTask("A2", func() {
				fmt.Println("A2")
				q.Put('a')
			}).Priority(gotaskflow.LOW),
			gotaskflow.NewTask("B2", func() {
				fmt.Println("B2")
				q.Put('b')
			}).Priority(gotaskflow.HIGH),
			gotaskflow.NewTask("C2", func() {
				fmt.Println("C2")
				q.Put('c')
			}).Priority(gotaskflow.NORMAL)
		sf.Push(A2, B2, C2)
	}).Priority(gotaskflow.LOW)

	tf.Push(B, C, suc)
	exector.Run(tf).Wait()

	for _, val := range []byte{'C', 'B', 'b', 'c', 'a'} {
		real := q.PeakAndTake()
		fmt.Printf("%c ", real)
		if val != real {
			t.Fail()
		}
	}
}

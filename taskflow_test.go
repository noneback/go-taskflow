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

type rgChain[R comparable] struct {
	rgs []*rgroup[R]
}

func newRgChain[R comparable]() *rgChain[R] {
	return &rgChain[R]{
		rgs: make([]*rgroup[R], 0),
	}
}

func (c *rgChain[R]) grouping(rs ...R) {
	g := newRg[R]()
	g.push(rs...)
	c.rgs = append(c.rgs, g)
}

// result group
type rgroup[R comparable] struct {
	pre, next *rgroup[R]
	elems     map[R]struct{}
}

func newRg[R comparable]() *rgroup[R] {
	return &rgroup[R]{
		elems: make(map[R]struct{}),
	}
}

func (g *rgroup[R]) push(rs ...R) {
	for _, r := range rs {
		g.elems[r] = struct{}{}
	}
}

func (g *rgroup[R]) chain(successor *rgroup[R]) {
	g.next = successor
	successor.pre = g.next
}

func (g *rgroup[R]) contains(r R) bool {
	_, ok := g.elems[r]
	return ok
}

func checkTopology[R comparable](t *testing.T, q *utils.Queue[R], chain *rgChain[R]) {
	for _, g := range chain.rgs {
		for len(g.elems) != 0 {
			node := q.Pop()
			if g.contains(node) {
				delete(g.elems, node)
			} else {
				fmt.Println("failed in", node)
				t.Fail()
			}
		}
	}
}

var executor = gotaskflow.NewExecutor(10)

func TestTaskFlow(t *testing.T) {
	q := utils.NewQueue[string]()
	tf := gotaskflow.NewTaskFlow("G")
	A, B, C :=
		tf.NewTask("A", func() {
			fmt.Println("A")
			q.Put("A")
		}),
		tf.NewTask("B", func() {
			fmt.Println("B")
			q.Put("B")
		}),
		tf.NewTask("C", func() {
			fmt.Println("C")
			q.Put("C")
		})

	A1, B1, _ :=
		tf.NewTask("A1", func() {
			fmt.Println("A1")
			q.Put("A1")
		}),
		tf.NewTask("B1", func() {
			fmt.Println("B1")
			q.Put("B1")
		}),
		tf.NewTask("C1", func() {
			fmt.Println("C1")
			q.Put("C1")
		})
	chains := newRgChain[string]()
	chains.grouping("C1", "A1", "B1", "A", "C")
	chains.grouping("B")

	A.Precede(B)
	C.Precede(B)
	A1.Precede(B)
	C.Succeed(A1)
	C.Succeed(B1)

	t.Run("TestViz", func(t *testing.T) {
		if err := tf.Dump(os.Stdout); err != nil {
			panic(err)
		}
	})

	executor.Run(tf).Wait()
	if q.Len() != 6 {
		t.Fail()
	}

	// checkTopology(t, q, chains)
}

func TestSubflow(t *testing.T) {
	q := utils.NewQueue[string]()
	// chains := newRgChain[string]()
	tf := gotaskflow.NewTaskFlow("G")
	A, B, C :=
		tf.NewTask("A", func() {
			fmt.Println("A")
			q.Put("A")
		}),
		tf.NewTask("B", func() {
			fmt.Println("B")
			q.Put("B")
		}),
		tf.NewTask("C", func() {
			fmt.Println("C")
			q.Put("C")
		})

	A1, B1, C1 :=
		tf.NewTask("A1", func() {
			fmt.Println("A1")
			q.Put("A1")
		}),
		tf.NewTask("B1", func() {
			fmt.Println("B1")
			q.Put("B1")
		}),
		tf.NewTask("C1", func() {
			fmt.Println("C1")
			q.Put("C1")
		})
	A.Precede(B)
	C.Precede(B)
	C.Succeed(A1)
	C.Succeed(B1)

	subflow := tf.NewSubflow("sub1", func(sf *gotaskflow.Subflow) {
		A2, B2, C2 :=
			sf.NewTask("A2", func() {
				fmt.Println("A2")
				q.Put("A2")
			}),
			sf.NewTask("B2", func() {
				fmt.Println("B2")
				q.Put("B2")
			}),
			sf.NewTask("C2", func() {
				fmt.Println("C2")
				q.Put("C2")
			})
		A2.Precede(B2)
		C2.Precede(B2)
		cond := sf.NewCondition("cond", func() uint {
			return 0
		})

		ssub := sf.NewSubflow("sub in sub", func(sf *gotaskflow.Subflow) {
			sf.NewTask("done", func() {
				fmt.Println("done")
			})
		})

		cond.Precede(ssub, cond)

	})

	subflow2 := tf.NewSubflow("sub2", func(sf *gotaskflow.Subflow) {
		A3, B3, C3 :=
			sf.NewTask("A3", func() {
				fmt.Println("A3")
				q.Put("A3")
			}),
			sf.NewTask("B3", func() {
				fmt.Println("B3")
				q.Put("B3")
			}),
			sf.NewTask("C3", func() {
				fmt.Println("C3")
				q.Put("C3")
				// time.Sleep(10 * time.Second)
			})
		A3.Precede(B3)
		C3.Precede(B3)

	})

	subflow.Precede(subflow2)
	C1.Precede(subflow)
	C1.Succeed(C)

	executor.Run(tf)
	executor.Wait()
	if err := tf.Dump(os.Stdout); err != nil {
		log.Fatal(err)
	}
	executor.Profile(os.Stdout)

	chain := newRgChain[string]()

	// Group 1 - Top-level nodes
	chain.grouping("A1", "B1", "A")
	chain.grouping("C")
	chain.grouping("B", "C1")
	chain.grouping("A2", "C2")
	chain.grouping("B2")

	// Group 2 - Connections under A, B, C
	chain.grouping("A3", "C3")
	chain.grouping("B3")

	// validate
	if q.Len() != 12 {
		t.Fail()
	}
	// checkTopology(t, q, chain)
}

// ERROR robust testing
func TestTaskflowPanic(t *testing.T) {
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
			panic("panic C")
		})
	A.Precede(B)
	C.Precede(B)

	executor.Run(tf).Wait()
}

func TestSubflowPanic(t *testing.T) {
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
	A.Precede(B)
	C.Precede(B)

	subflow := tf.NewSubflow("sub1", func(sf *gotaskflow.Subflow) {
		A2, B2, C2 :=
			tf.NewTask("A2", func() {
				fmt.Println("A2")
				time.Sleep(1 * time.Second)
			}),
			tf.NewTask("B2", func() {
				fmt.Println("B2")
			}),
			tf.NewTask("C2", func() {
				fmt.Println("C2")
				panic("C2 paniced")
			})
		A2.Precede(B2)
		panic("subflow panic")
		C2.Precede(B2)
	})

	subflow.Precede(B)
	executor.Run(tf)
	executor.Wait()
	if err := tf.Dump(os.Stdout); err != nil {
		fmt.Errorf("%v", err)
	}
	executor.Profile(os.Stdout)
}

func TestTaskflowCondition(t *testing.T) {

	q := utils.NewQueue[string]()
	chain := newRgChain[string]()
	tf := gotaskflow.NewTaskFlow("G")
	t.Run("normal", func(t *testing.T) {
		A, B, C :=
			tf.NewTask("A", func() {
				fmt.Println("A")
				q.Put("A")
			}),
			tf.NewTask("B", func() {
				fmt.Println("B")
				q.Put("B")
			}),
			tf.NewTask("C", func() {
				fmt.Println("C")
				q.Put("C")
			})
		A.Precede(B)
		C.Precede(B)

		fail, success := tf.NewTask("failed", func() {
			fmt.Println("Failed")
			q.Put("failed")
			t.Fail()
		}), tf.NewTask("success", func() {
			fmt.Println("success")
			q.Put("success")
		})

		cond := tf.NewCondition("cond", func() uint {
			q.Put("cond")
			return 0
		})
		B.Precede(cond)
		cond.Precede(success, fail)

		suc := tf.NewSubflow("sub1", func(sf *gotaskflow.Subflow) {
			A2, B2, C2 :=
				sf.NewTask("A2", func() {
					fmt.Println("A2")
					q.Put("A2")
				}),
				sf.NewTask("B2", func() {
					fmt.Println("B2")
					q.Put("B2")
				}),
				sf.NewTask("C2", func() {
					fmt.Println("C2")
					q.Put("C2")
				})
			A2.Precede(B2)
			C2.Precede(B2)
		})
		fs := tf.NewTask("fail_single", func() {
			fmt.Println("it should be canceled")
			q.Put("fail_single")
		})
		fail.Precede(fs, suc)
		// success.Precede(suc)
		if err := tf.Dump(os.Stdout); err != nil {
			fmt.Errorf("%v", err)
		}
		executor.Run(tf).Wait()

		executor.Profile(os.Stdout)
		chain.grouping("A", "C")
		chain.grouping("B")
		chain.grouping("cond")
		chain.grouping("success")

		checkTopology(t, q, chain)

	})

	t.Run("start with condion node", func(t *testing.T) {
		i := 0
		tf := gotaskflow.NewTaskFlow("G")

		cond := tf.NewCondition("cond", func() uint {
			if i == 0 {
				return 0
			} else {
				return 1
			}
		})

		zero, one := tf.NewTask("zero", func() {
			fmt.Println("zero")
		}), tf.NewTask("one", func() {
			fmt.Println("one")
		})
		cond.Precede(zero, one)

		executor.Run(tf).Wait()

		if err := tf.Dump(os.Stdout); err != nil {
			log.Fatal(err)
		}
		executor.Profile(os.Stdout)

	})

}

func TestTaskflowLoop(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		i := 0
		tf := gotaskflow.NewTaskFlow("G")
		init, cond, body, back, done :=
			tf.NewTask("init", func() {
				i = 0
				fmt.Println("i=0")
			}),
			tf.NewCondition("while i < 5", func() uint {
				if i < 5 {
					return 0
				} else {
					return 1
				}
			}),
			tf.NewTask("body", func() {
				i += 1
				fmt.Println("i++ =", i)
			}),
			tf.NewCondition("back", func() uint {
				fmt.Println("back")
				return 0
			}),
			tf.NewTask("done", func() {
				fmt.Println("done")
			})

		init.Precede(cond)
		cond.Precede(body, done)
		body.Precede(back)
		back.Precede(cond)
		if err := tf.Dump(os.Stdout); err != nil {
			// log.Fatal(err)
		}
		executor.Run(tf).Wait()
		if i < 5 {
			t.Fail()
		}

		executor.Profile(os.Stdout)
	})

	t.Run("simple loop", func(t *testing.T) {
		i := 0
		tf := gotaskflow.NewTaskFlow("G")
		init := tf.NewTask("init", func() {
			i = 0
		})
		cond := tf.NewCondition("cond", func() uint {
			i++
			fmt.Println("i++ =", i)
			if i > 2 {
				return 0
			} else {
				return 1
			}
		})

		done := tf.NewTask("done", func() {
			fmt.Println("done")
		})

		init.Precede(cond)
		cond.Precede(done, cond)

		executor.Run(tf).Wait()
		if i <= 2 {
			t.Fail()
		}

		if err := tf.Dump(os.Stdout); err != nil {
			// log.Fatal(err)
		}
		executor.Profile(os.Stdout)
	})
}

func TestTaskflowPriority(t *testing.T) {
	executor := gotaskflow.NewExecutor(uint(2))
	q := utils.NewQueue[byte]()
	tf := gotaskflow.NewTaskFlow("G")
	tf.NewTask("B", func() {
		fmt.Println("B")
		q.Put('B')
	}).Priority(gotaskflow.NORMAL)

	tf.NewTask("C", func() {
		fmt.Println("C")
		q.Put('C')
	}).Priority(gotaskflow.HIGH)
	tf.NewSubflow("sub1", func(sf *gotaskflow.Subflow) {
		sf.NewTask("A2", func() {
			fmt.Println("A2")
			q.Put('a')
		}).Priority(gotaskflow.LOW)
		sf.NewTask("B2", func() {
			fmt.Println("B2")
			q.Put('b')
		}).Priority(gotaskflow.HIGH)
		sf.NewTask("C2", func() {
			fmt.Println("C2")
			q.Put('c')
		}).Priority(gotaskflow.NORMAL)
	}).Priority(gotaskflow.LOW)

	executor.Run(tf).Wait()

	for _, val := range []byte{'C', 'B', 'b', 'c', 'a'} {
		real := q.Pop()
		fmt.Printf("%c ", real)
		if val != real {
			t.Fail()
		}
	}
}

func TestTaskflowNotInFlow(t *testing.T) {
	tf := gotaskflow.NewTaskFlow("tf")
	task := tf.NewTask("init", func() {
		fmt.Println("task init")
	})
	cnt := 0
	for i := 0; i < 10; i++ {
		task.Precede(tf.NewTask("test", func() {
			fmt.Println(cnt)
			cnt++
		}))
	}

	executor.Run(tf).Wait()
}

func TestTaskflowFrozen(t *testing.T) {
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
	A.Precede(B)
	C.Precede(B)

	executor.Run(tf).Wait()
	utils.AssertPanics(t, "frozen", func() {
		tf.NewTask("tt", func() {
			fmt.Println("should not")
		})
	})
}

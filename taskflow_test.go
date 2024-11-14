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
			node := q.PeakAndTake()
			if g.contains(node) {
				delete(g.elems, node)
			} else {
				fmt.Println(node)
				t.Fail()
			}
		}
	}
}

var executor = gotaskflow.NewExecutor(10)

func TestTaskFlow(t *testing.T) {
	q := utils.NewQueue[string]()

	A, B, C :=
		gotaskflow.NewTask("A", func() {
			fmt.Println("A")
			q.Put("A")
		}),
		gotaskflow.NewTask("B", func() {
			fmt.Println("B")
			q.Put("B")
		}),
		gotaskflow.NewTask("C", func() {
			fmt.Println("C")
			q.Put("C")
		})

	A1, B1, C1 :=
		gotaskflow.NewTask("A1", func() {
			fmt.Println("A1")
			q.Put("A1")
		}),
		gotaskflow.NewTask("B1", func() {
			fmt.Println("B1")
			q.Put("B1")
		}),
		gotaskflow.NewTask("C1", func() {
			fmt.Println("C1")
			q.Put("C1")
		})
	chains := newRgChain[string]()
	chains.grouping("C1", "A1", "B1", "A")
	chains.grouping("C")
	chains.grouping("B")

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

	// validate
	checkTopology(t, q, chains)
}

func TestSubflow(t *testing.T) {
	q := utils.NewQueue[string]()
	// chains := newRgChain[string]()

	A, B, C :=
		gotaskflow.NewTask("A", func() {
			fmt.Println("A")
			q.Put("A")
		}),
		gotaskflow.NewTask("B", func() {
			fmt.Println("B")
			q.Put("B")
		}),
		gotaskflow.NewTask("C", func() {
			fmt.Println("C")
			q.Put("C")
		})

	A1, B1, C1 :=
		gotaskflow.NewTask("A1", func() {
			fmt.Println("A1")
			q.Put("A1")
		}),
		gotaskflow.NewTask("B1", func() {
			fmt.Println("B1")
			q.Put("B1")
		}),
		gotaskflow.NewTask("C1", func() {
			fmt.Println("C1")
			q.Put("C1")
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
				q.Put("A2")
			}),
			gotaskflow.NewTask("B2", func() {
				fmt.Println("B2")
				q.Put("B2")
			}),
			gotaskflow.NewTask("C2", func() {
				fmt.Println("C2")
				q.Put("C2")
			})
		A2.Precede(B2)
		C2.Precede(B2)
		sf.Push(A2, B2, C2)
	})

	subflow2 := gotaskflow.NewSubflow("sub2", func(sf *gotaskflow.Subflow) {
		A3, B3, C3 :=
			gotaskflow.NewTask("A3", func() {
				fmt.Println("A3")
				q.Put("A3")
			}),
			gotaskflow.NewTask("B3", func() {
				fmt.Println("B3")
				q.Put("B3")
			}),
			gotaskflow.NewTask("C3", func() {
				fmt.Println("C3")
				q.Put("C3")
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

	chain := newRgChain[string]()

	// Group 1 - Top-level nodes
	chain.grouping("C1", "A1", "B1", "C1", "A", "A2", "C2", "B2")

	// Group 2 - Connections under A, B, C
	chain.grouping("C", "A3", "C3",
		"B3")

	chain.grouping("B")
	// validate
	if q.Len() != 12 {
		t.Fail()
	}
	checkTopology(t, q, chain)
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
	q := utils.NewQueue[string]()
	chain := newRgChain[string]()
	t.Run("normal", func(t *testing.T) {
		A, B, C :=
			gotaskflow.NewTask("A", func() {
				fmt.Println("A")
				q.Put("A")
			}),
			gotaskflow.NewTask("B", func() {
				fmt.Println("B")
				q.Put("B")
			}),
			gotaskflow.NewTask("C", func() {
				fmt.Println("C")
				q.Put("C")
			})
		A.Precede(B)
		C.Precede(B)
		tf := gotaskflow.NewTaskFlow("G")
		tf.Push(A, B, C)
		fail, success := gotaskflow.NewTask("failed", func() {
			fmt.Println("Failed")
			q.Put("failed")
			t.Fail()
		}), gotaskflow.NewTask("success", func() {
			fmt.Println("success")
			q.Put("success")
		})

		cond := gotaskflow.NewCondition("cond", func() uint {
			q.Put("cond")
			return 0
		})
		B.Precede(cond)
		cond.Precede(success, fail)

		suc := gotaskflow.NewSubflow("sub1", func(sf *gotaskflow.Subflow) {
			A2, B2, C2 :=
				gotaskflow.NewTask("A2", func() {
					fmt.Println("A2")
					q.Put("A2")
				}),
				gotaskflow.NewTask("B2", func() {
					fmt.Println("B2")
					q.Put("B2")
				}),
				gotaskflow.NewTask("C2", func() {
					fmt.Println("C2")
					q.Put("C2")
				})
			sf.Push(A2, B2, C2)
			A2.Precede(B2)
			C2.Precede(B2)
		})
		fs := gotaskflow.NewTask("fail_single", func() {
			fmt.Println("it should be canceled")
			q.Put("fail_single")
		})
		fail.Precede(fs, suc)
		// success.Precede(suc)
		tf.Push(cond, success, fail, fs, suc)
		executor.Run(tf).Wait()

		if err := gotaskflow.Visualize(tf, os.Stdout); err != nil {
			fmt.Errorf("%v", err)
		}
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

		cond := gotaskflow.NewCondition("cond", func() uint {
			if i == 0 {
				return 0
			} else {
				return 1
			}
		})

		zero, one := gotaskflow.NewTask("zero", func() {
			fmt.Println("zero")
		}), gotaskflow.NewTask("one", func() {
			fmt.Println("one")
		})
		cond.Precede(zero, one)

		tf.Push(zero, one, cond)
		executor.Run(tf).Wait()

		if err := gotaskflow.Visualize(tf, os.Stdout); err != nil {
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
			gotaskflow.NewTask("body", func() {
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
		if i < 5 {
			t.Fail()
		}

		if err := gotaskflow.Visualize(tf, os.Stdout); err != nil {
			// log.Fatal(err)
		}
		executor.Profile(os.Stdout)
	})

	t.Run("simple loop", func(t *testing.T) {
		i := 0
		tf := gotaskflow.NewTaskFlow("G")
		init := gotaskflow.NewTask("init", func() {
			i = 0
		})
		cond := gotaskflow.NewCondition("cond", func() uint {
			i++
			fmt.Println("i++ =", i)
			if i > 2 {
				return 0
			} else {
				return 1
			}
		})

		done := gotaskflow.NewTask("done", func() {
			fmt.Println("done")
		})

		init.Precede(cond)
		cond.Precede(done, cond)

		tf.Push(done, cond, init)
		executor.Run(tf).Wait()
		if i <= 2 {
			t.Fail()
		}

		if err := gotaskflow.Visualize(tf, os.Stdout); err != nil {
			// log.Fatal(err)
		}
		executor.Profile(os.Stdout)
	})
}

func TestTaskflowPriority(t *testing.T) {
	executor := gotaskflow.NewExecutor(uint(2))
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
	executor.Run(tf).Wait()

	for _, val := range []byte{'C', 'B', 'b', 'c', 'a'} {
		real := q.PeakAndTake()
		fmt.Printf("%c ", real)
		if val != real {
			t.Fail()
		}
	}
}

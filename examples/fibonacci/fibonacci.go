package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	gotaskflow "github.com/noneback/go-taskflow"
)

func main() {
	n := 10
	fib := make([]int, n+1)

	executor := gotaskflow.NewExecutor(uint(runtime.NumCPU()))
	tf := gotaskflow.NewTaskFlow("fibonacci")

	f0 := tf.NewTask("F0", func() { fib[0] = 1 })
	f1 := tf.NewTask("F1", func() { fib[1] = 1 })

	tasks := []*gotaskflow.Task{f0, f1}
	for k := 2; k <= n; k++ {
		k := k
		task := tf.NewTask(fmt.Sprintf("F%d", k), func() {
			fib[k] = fib[k-1] + fib[k-2]
		})
		tasks[k-1].Precede(task) // F(n-1) -> F(n)
		tasks[k-2].Precede(task) // F(n-2) -> F(n)
		tasks = append(tasks, task)
	}
	if err := tf.Dump(os.Stdout); err != nil {
		log.Fatal(err)
	}
	executor.Run(tf).Wait()

	fmt.Printf("F(%d) = %d\n", n+1, fib[n])
}

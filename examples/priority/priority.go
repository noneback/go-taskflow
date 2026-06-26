package main

import (
	"fmt"
	"sync"

	gotaskflow "github.com/noneback/go-taskflow"
)

func main() {
	executor := gotaskflow.NewExecutor(2, gotaskflow.WithProfiler())
	tf := gotaskflow.NewTaskFlow("priority-demo")

	var mu sync.Mutex
	var order []string

	tf.NewTask("normal_task", func() {
		mu.Lock()
		order = append(order, "normal_task")
		mu.Unlock()
		fmt.Println("normal_task")
	}).Priority(gotaskflow.NORMAL)

	tf.NewTask("high_task", func() {
		mu.Lock()
		order = append(order, "high_task")
		mu.Unlock()
		fmt.Println("high_task")
	}).Priority(gotaskflow.HIGH)

	tf.NewSubflow("low_subflow", func(sf *gotaskflow.Subflow) {
		sf.NewTask("sub_low", func() {
			mu.Lock()
			order = append(order, "sub_low")
			mu.Unlock()
			fmt.Println("  sub_low")
		}).Priority(gotaskflow.LOW)
		sf.NewTask("sub_high", func() {
			mu.Lock()
			order = append(order, "sub_high")
			mu.Unlock()
			fmt.Println("  sub_high")
		}).Priority(gotaskflow.HIGH)
		sf.NewTask("sub_normal", func() {
			mu.Lock()
			order = append(order, "sub_normal")
			mu.Unlock()
			fmt.Println("  sub_normal")
		}).Priority(gotaskflow.NORMAL)
	}).Priority(gotaskflow.LOW)

	executor.Run(tf).Wait()

	fmt.Println("\nexecution order:", order)
}

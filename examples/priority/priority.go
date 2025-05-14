package main

import (
	"fmt"

	gotaskflow "github.com/noneback/go-taskflow"
	"github.com/noneback/go-taskflow/utils"
)

func main() {
	exector := gotaskflow.NewExecutor(uint(2))
	q := utils.NewQueue[byte](true)
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

	exector.Run(tf).Wait()
}

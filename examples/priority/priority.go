package priority

import (
	"fmt"

	gotaskflow "github.com/noneback/go-taskflow"
	"github.com/noneback/go-taskflow/utils"
)

func main() {
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
}

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

	executor := gotaskflow.NewExecutor(uint(runtime.NumCPU()) * 100)
	i := 0
	tf := gotaskflow.NewTaskFlow("G")
	init, cond, body, back, done :=
		gotaskflow.NewTask("init", func() {
			i = 0
			fmt.Println("i=0")
		}),
		gotaskflow.NewCondition("while i < 5", func() uint {
			time.Sleep(100 * time.Millisecond)
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
		log.Fatal("i < 5")
	}

	if err := gotaskflow.Visualize(tf, os.Stdout); err != nil {
		log.Fatal(err)
	}
	executor.Profile(os.Stdout)
}

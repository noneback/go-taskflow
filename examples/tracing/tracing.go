package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	gotaskflow "github.com/noneback/go-taskflow"
)

// This example demonstrates how to capture Chrome Trace Event data
// using WithTracer(). The output can be visualized in chrome://tracing
// or Perfetto UI (https://ui.perfetto.dev/).
func main() {
	executor := gotaskflow.NewExecutor(uint(runtime.NumCPU()*4), gotaskflow.WithTracer())

	tf := gotaskflow.NewTaskFlow("tracing-demo")

	// Stage 1: parallel fetch and load
	fetch := tf.NewTask("fetch", func() {
		fmt.Println("fetch")
		time.Sleep(10 * time.Millisecond)
	})
	load := tf.NewTask("load", func() {
		fmt.Println("load")
		time.Sleep(8 * time.Millisecond)
	})

	// Stage 2: process depends on both fetch and load
	process := tf.NewSubflow("process", func(sf *gotaskflow.Subflow) {
		transform := sf.NewTask("transform", func() {
			fmt.Println("transform")
			time.Sleep(5 * time.Millisecond)
		})
		enrich := sf.NewTask("enrich", func() {
			fmt.Println("enrich")
			time.Sleep(4 * time.Millisecond)
		})
		transform.Precede(enrich)
	})

	// Stage 3: output waits for process to complete
	output := tf.NewTask("output", func() {
		fmt.Println("output")
		time.Sleep(3 * time.Millisecond)
	})

	fetch.Precede(process)
	load.Precede(process)
	process.Precede(output)

	executor.Run(tf).Wait()

	fmt.Println("--- Chrome Trace JSON (open in chrome://tracing or https://ui.perfetto.dev) ---")
	if err := executor.Trace(os.Stdout); err != nil {
		log.Fatal(err)
	}
}

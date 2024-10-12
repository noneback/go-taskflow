# Go-Taskflow
A static DAG (Directed Acyclic Graph) task computing framework for Go, inspired by taskflow-cpp, with Go's native capabilities and simplicity, suitable for complex dependency management in concurrent tasks.

## Feature
- **High extensibility**: Easily extend the framework to adapt to various specific use cases.

- **Native Go's concurrency model**: Leverages Go's goroutines to manage concurrent task execution effectively.

- **User-friendly programming interface**: Simplify complex task dependency management using Go.

- **Static and subflow tasking**: Define static tasks as well as nested subflows for greater modularity.

- **Built-in visualization & profiling tools**: Generate visual representations of tasks and profile task execution performance using integrated tools, making debugging and optimization easier.

## Use Cases

- **Data Pipeline**: Orchestrate data processing stages that have complex dependencies.

- **Workflow Automation**: Define and run automation workflows where tasks have a clear sequence and dependency structure.

- **Parallel Tasking**: Execute independent tasks concurrently to fully utilize CPU resources.

## Example
```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"

	gotaskflow "github.com/noneback/go-taskflow"
)

func main() {
	executor := gotaskflow.NewExecutor(runtime.NumCPU() - 1)
	tf := gotaskflow.NewTaskFlow("G")
	A, B, C :=
		gotaskflow.NewTask("A", func(ctx *context.Context) {
			fmt.Println("A")
		}),
		gotaskflow.NewTask("B", func(ctx *context.Context) {
			fmt.Println("B")
		}),
		gotaskflow.NewTask("C", func(ctx *context.Context) {
			fmt.Println("C")
		})

	A1, B1, C1 :=
		gotaskflow.NewTask("A1", func(ctx *context.Context) {
			fmt.Println("A1")
		}),
		gotaskflow.NewTask("B1", func(ctx *context.Context) {
			fmt.Println("B1")
		}),
		gotaskflow.NewTask("C1", func(ctx *context.Context) {
			fmt.Println("C1")
		})
	A.Precede(B)
	C.Precede(B)
	A1.Precede(B)
	C.Succeed(A1)
	C.Succeed(B1)

	subflow := gotaskflow.NewSubflow("sub1", func(sf *gotaskflow.Subflow) {
		A2, B2, C2 :=
			gotaskflow.NewTask("A2", func(ctx *context.Context) {
				fmt.Println("A2")
			}),
			gotaskflow.NewTask("B2", func(ctx *context.Context) {
				fmt.Println("B2")
			}),
			gotaskflow.NewTask("C2", func(ctx *context.Context) {
				fmt.Println("C2")
			})
		A2.Precede(B2)
		C2.Precede(B2)
		sf.Push(A2, B2, C2)
	})

	subflow2 := gotaskflow.NewSubflow("sub2", func(sf *gotaskflow.Subflow) {
		A3, B3, C3 :=
			gotaskflow.NewTask("A3", func(ctx *context.Context) {
				fmt.Println("A3")
			}),
			gotaskflow.NewTask("B3", func(ctx *context.Context) {
				fmt.Println("B3")
			}),
			gotaskflow.NewTask("C3", func(ctx *context.Context) {
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
	exector.Run(tf).Wait()
	fmt.Println("Print DOT")
	if err := gotaskflow.Visualizer.Visualize(tf, os.Stdout); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Print Flamegraph")
	if err :=exector.Profile(os.Stdout);err != nil {
		log.Fatal(err)
	}
}
```
### How to use visualize taskflow
```go
if err := gotaskflow.Visualizer.Visualize(tf, os.Stdout); err != nil {
		log.Fatal(err)
}
```
`Visualize` generate raw string in dot format, just use dot to draw a DAG svg.

![dot](https://raw.githubusercontent.com/noneback/images/ae31f3ea57f3f1b8d4cf94300a5ff502b2340214/graphviz.svg)
### How to use profile taskflow
```go
if err :=exector.Profile(os.Stdout);err != nil {
		log.Fatal(err)
}
```

`Profile` alse generate raw string in flamegraph format, just use flamegraph to draw a flamegraph svg.

![flg](https://raw.githubusercontent.com/noneback/images/ae31f3ea57f3f1b8d4cf94300a5ff502b2340214/t.svg)
## What's next
- [ ] Taskflow Composition
- [ ] Conditional Tasking
- [ ] Taskflow Loop Repeatition

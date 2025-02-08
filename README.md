# Go-Taskflow
[![codecov](https://codecov.io/github/noneback/go-taskflow/graph/badge.svg?token=CITXYA10C6)](https://codecov.io/github/noneback/go-taskflow)
[![Go Reference](https://pkg.go.dev/badge/github.com/noneback/go-taskflow.svg)](https://pkg.go.dev/github.com/noneback/go-taskflow)
[![Go Report Card](https://goreportcard.com/badge/github.com/noneback/go-taskflow)](https://goreportcard.com/report/github.com/noneback/go-taskflow)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

![go-taskflow](https://socialify.git.ci/noneback/go-taskflow/image?description=1&language=1&name=1&pattern=Solid&theme=Auto)

A General-purpose Task-parallel Programming Framework for Go, inspired by [taskflow-cpp](https://github.com/taskflow/taskflow), with Go's native capabilities and simplicity, suitable for complex dependency management in concurrent tasks.

## Feature
- **High extensibility**: Easily extend the framework to adapt to various specific use cases.

- **Native Go's concurrency model**: Leverages Go's goroutines to manage concurrent task execution effectively.

- **User-friendly programming interface**: Simplify complex task dependency management using Go.

- **Static\Subflow\Conditional\Cyclic tasking**: Define static tasks, condition nodes, nested subflows and cyclic flow to enhance modularity and programmability.

	| Static | Subflow | Condition | Cyclic |
	|:-----------|:------------:|------------:|------------:|
	| ![](image/simple.svg)     |   ![](image/subflow.svg)   |      ![](image/condition.svg) |      ![](image/loop.svg) |

- **Priority Task Schedule**: Define tasks' priority, higher priority tasks will be scheduled first.

- **Built-in visualization & profiling tools**: Generate visual representations of tasks and profile task execution performance using integrated tools, making debugging and optimization easier.

## Use Cases

- **Data Pipeline**: Orchestrate data processing stages that have complex dependencies.

- **Workflow Automation**: Define and run automation workflows where tasks have a clear sequence and dependency structure.

- **Parallel Tasking**: Execute independent tasks concurrently to fully utilize CPU resources.

## Example
import latest version: `go get -u github.com/noneback/go-taskflow`

```go
package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"sync"

	gtf "github.com/noneback/go-taskflow"
)

// meger sorted src to sorted dest
func mergeInto(dest, src []int) []int {
	size := len(dest) + len(src)
	tmp := make([]int, 0, size)
	i, j := 0, 0
	for i < len(dest) && j < len(src) {
		if dest[i] < src[j] {
			tmp = append(tmp, dest[i])
			i++
		} else {
			tmp = append(tmp, src[j])
			j++
		}
	}

	if i < len(dest) {
		tmp = append(tmp, dest[i:]...)
	} else {
		tmp = append(tmp, src[j:]...)
	}

	return tmp
}
func main() {
	size := 100
	radomArr := make([][]int, 10)
	sortedArr := make([]int, 0, 10*size)
	mutex := &sync.Mutex{}

	for i := 0; i < 10; i++ {
		for j := 0; j < size; j++ {
			radomArr[i] = append(radomArr[i], rand.Int())
		}
	}

	sortTasks := make([]*gtf.Task, 10)
	tf := gtf.NewTaskFlow("merge sort")
	done := tf.NewTask("Done", func() {
		if !slices.IsSorted(sortedArr) {
			log.Fatal("Failed")
		}
		fmt.Println("Sorted")
		fmt.Println(sortedArr[:1000])
	})

	for i := 0; i < 10; i++ {
		sortTasks[i] = tf.NewTask("sort_"+strconv.Itoa(i), func() {
			arr := radomArr[i]
			slices.Sort(arr)
			mutex.Lock()
			defer mutex.Unlock()
			sortedArr = mergeInto(sortedArr, arr)
		})

	}
	done.Succeed(sortTasks...)

	executor := gtf.NewExecutor(1000)

	executor.Run(tf).Wait()

	if err := tf.Dump(os.Stdout); err != nil {
		log.Fatal("V->", err)
	}

	if err := executor.Profile(os.Stdout); err != nil {
		log.Fatal("P->", err)
	}

}
```

[more code examples](https://github.com/noneback/go-taskflow/tree/main/examples)

## Understand Condition Task Correctly
Condition Node is special in [taskflow-cpp](https://github.com/taskflow/taskflow). It not only enrolls in Condition Control but also in Looping.

Our repo keeps almost the same behavior. You should read [ConditionTasking](https://taskflow.github.io/taskflow/ConditionalTasking.html) to avoid common pitfalls.

## Error Handling in go-taskflow

`errors` in golang are values. It is the user's job to handle it correctly. 

Only unrecovered `panic` needs to be addressed by the framework. Now, if it happens, the whole parent graph will be canceled, leaving the rest tasks undone. This behavior may evolve someday. If you have any good thoughts, feel free to let me know.

If you prefer not to interrupt the whole taskflow when panics occur, you can also handle panics manually while registering tasks.
Eg: 
```go
tf.NewTask("not interrupt", func() {
	defer func() {
		if r := recover(); r != nil {
			// deal with it.
		}
	}()
	// user functions.
)
```

## How to use visualize taskflow
```go
if err := tf.Dump(os.Stdout); err != nil {
		log.Fatal(err)
}
```
`tf.Dump` generates raw strings in dot format, use `dot` to draw a Graph svg.

![dot](image/desc.svg)

## How to use profile taskflow
```go
if err :=exector.Profile(os.Stdout);err != nil {
		log.Fatal(err)
}
```

`Profile` generates raw strings in flamegraph format, use `flamegraph` to draw a flamegraph svg.

![flg](image/fl.svg)

## What's more
Any Features Request or Discussions are all welcomed.


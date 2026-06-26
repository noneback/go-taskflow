# go-taskflow

[![codecov](https://codecov.io/github/noneback/go-taskflow/graph/badge.svg?token=CITXYA10C6)](https://codecov.io/github/noneback/go-taskflow)
[![Go Reference](https://pkg.go.dev/badge/github.com/noneback/go-taskflow.svg)](https://pkg.go.dev/github.com/noneback/go-taskflow)
[![Go Report Card](https://goreportcard.com/badge/github.com/noneback/go-taskflow)](https://goreportcard.com/report/github.com/noneback/go-taskflow)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)
[![HN Daily Top #26](https://img.shields.io/badge/HN_Daily_Top-%2326-orange?logo=ycombinator&style=flat-square)](https://news.ycombinator.com/front?day=2024-11-15)

[![DeepWiki][deepwiki-image]][deepwiki-url]

[deepwiki-url]: https://deepwiki.com/noneback/go-taskflow
[deepwiki-image]: https://img.shields.io/badge/Chat%20with-DeepWiki%20🤖-20B2AA

![go-taskflow](https://socialify.git.ci/noneback/go-taskflow/image?description=1&language=1&name=1&pattern=Solid&theme=Auto)

go-taskflow is a general-purpose task-parallel programming framework for Go, inspired by [taskflow-cpp](https://github.com/taskflow/taskflow). It leverages Go's native capabilities and simplicity, making it ideal for managing complex dependencies in concurrent tasks.

## Features

- **High Extensibility**: Easily extend the framework to adapt to various specific use cases.
- **Native Go Concurrency Model**: Leverages Go's goroutines for efficient concurrent task execution.
- **User-Friendly Programming Interface**: Simplifies complex task dependency management in Go.
- **Static, Subflow, Conditional, and Cyclic Tasking**: Define static tasks, condition nodes, nested subflows, and cyclic flows to enhance modularity and programmability.

    | Static | Subflow | Condition | Cyclic |
    |:-----------|:------------:|------------:|------------:|
    | ![](image/simple.svg)     |   ![](image/subflow.svg)   |      ![](image/condition.svg) |      ![](image/loop.svg) |

- **Priority Task Scheduling**: Assign task priorities to ensure higher-priority tasks are executed first.
- **Built-in Visualization, Profiling and Tracing**: Generate visual representations of tasks, profile task execution with flamegraph, and capture Chrome Trace events for timeline analysis.

## Use Cases

- **Data Pipelines**: Orchestrate data processing stages with complex dependencies.
- **AI Agent Workflow Automation**: Define and execute AI agent workflows with clear sequences and dependency structures.
- **Parallel Graph Tasking**: Execute graph-based tasks concurrently to maximize CPU utilization.

## Installation

Import the latest version of go-taskflow using:

```bash
go get -u github.com/noneback/go-taskflow
```

## Documentation

[DeepWiki Page](https://deepwiki.com/noneback/go-taskflow)

## Quick Start

A MapReduce word-count pipeline with parallel mappers, hash-partitioned shuffle, and reducers:

```go
package main

import (
    "fmt"
    "log"
    "os"
    "sort"
    "strings"
    "sync"

    gtf "github.com/noneback/go-taskflow"
)

const (
    numMappers  = 4
    numReducers = 2
)

var (
    mapOutputs [numMappers][numReducers]map[string]int
    mu         sync.Mutex
)

func hashPartition(word string) int {
    h := 0
    for _, c := range word {
        h = 31*h + int(c)
    }
    if h < 0 {
        h = -h
    }
    return h % numReducers
}

func main() {
    input := `the quick brown fox jumps over the lazy dog
the fox and the dog are friends the dog jumps over the fox
brown dog lazy fox the quick brown dog the lazy fox`

    var chunks [numMappers]string
    executor := gtf.NewExecutor(8, gtf.WithProfiler())
    tf := gtf.NewTaskFlow("word-count")

    // Phase 1: Split input into chunks for parallel processing
    splitTask := tf.NewTask("split_input", func() {
        words := strings.Fields(input)
        size := (len(words) + numMappers - 1) / numMappers
        for i := 0; i < numMappers; i++ {
            start, end := i*size, (i+1)*size
            if end > len(words) {
                end = len(words)
            }
            chunks[i] = strings.Join(words[start:end], " ")
        }
    })

    // Phase 2: Map — count words per chunk, partition by hash
    mapTasks := make([]*gtf.Task, numMappers)
    for i := 0; i < numMappers; i++ {
        idx := i
        mapTasks[idx] = tf.NewTask(fmt.Sprintf("map_%d", idx), func() {
            local := [numReducers]map[string]int{}
            for r := 0; r < numReducers; r++ {
                local[r] = make(map[string]int)
            }
            for _, w := range strings.Fields(chunks[idx]) {
                local[hashPartition(w)][w]++
            }
            mu.Lock()
            for r := 0; r < numReducers; r++ {
                mapOutputs[idx][r] = local[r]
            }
            mu.Unlock()
        })
    }
    splitTask.Precede(mapTasks...)

    // Phase 3: Reduce — aggregate each hash partition
    reduceTasks := make([]*gtf.Task, numReducers)
    reduceResults := make([]map[string]int, numReducers)
    for r := 0; r < numReducers; r++ {
        rIdx := r
        reduceResults[rIdx] = make(map[string]int)
        reduceTasks[rIdx] = tf.NewTask(fmt.Sprintf("reduce_%d", rIdx), func() {
            for m := 0; m < numMappers; m++ {
                for word, count := range mapOutputs[m][rIdx] {
                    reduceResults[rIdx][word] += count
                }
            }
        })
    }
    for _, mt := range mapTasks {
        mt.Precede(reduceTasks...)
    }

    // Phase 4: Merge final results
    mergeTask := tf.NewTask("merge_results", func() {
        final := make(map[string]int)
        for r := 0; r < numReducers; r++ {
            for w, c := range reduceResults[r] {
                final[w] += c
            }
        }
        keys := make([]string, 0, len(final))
        for k := range final {
            keys = append(keys, k)
        }
        sort.Strings(keys)
        for _, w := range keys {
            fmt.Printf("%-12s %d\n", w, final[w])
        }
    })
    for _, rt := range reduceTasks {
        rt.Precede(mergeTask)
    }

    executor.Run(tf).Wait()

    if err := tf.Dump(os.Stdout); err != nil {
        log.Fatal(err)
    }
}
```

For more examples, visit the [examples directory](https://github.com/noneback/go-taskflow/tree/main/examples).

## Benchmark

The following benchmarks provide a rough estimate of pure scheduling overhead using empty task functions. Note that most realistic workloads are I/O-bound, and their performance cannot be accurately reflected by these results. For CPU-intensive tasks, consider using [taskflow-cpp](https://github.com/taskflow/taskflow).

```plaintext
$ go test -bench=. -benchmem ./benchmark/
goos: darwin
goarch: arm64
pkg: github.com/noneback/go-taskflow/benchmark
cpu: Apple M4 Pro
BenchmarkConcurrent/N8-12              217042       5349 ns/op      1781 B/op       55 allocs/op
BenchmarkConcurrent/N32-12              47456      24439 ns/op      7566 B/op      213 allocs/op
BenchmarkConcurrent/N128-12             10000     116586 ns/op     32209 B/op      835 allocs/op
BenchmarkConcurrent/N512-12              2839     439930 ns/op    130337 B/op     3353 allocs/op
BenchmarkSerial/N8-12                  126259       9339 ns/op      1905 B/op       63 allocs/op
BenchmarkSerial/N32-12                  30313      39171 ns/op      7669 B/op      255 allocs/op
BenchmarkSerial/N128-12                  7758     156781 ns/op     30725 B/op     1023 allocs/op
BenchmarkSerial/N512-12                  1862     645739 ns/op    122952 B/op     4095 allocs/op
BenchmarkDiamond-12                    181072       6662 ns/op      1441 B/op       47 allocs/op
BenchmarkDenseLayers/L4xW4-12           85122      13461 ns/op      4352 B/op      123 allocs/op
BenchmarkDenseLayers/L4xW8-12           42927      27127 ns/op     11764 B/op      270 allocs/op
BenchmarkDenseLayers/L8xW4-12           44412      27565 ns/op      8963 B/op      251 allocs/op
BenchmarkDenseLayers/L8xW8-12           20775      58088 ns/op     25071 B/op      556 allocs/op
BenchmarkSubflow-12                    170228       6531 ns/op      1409 B/op       45 allocs/op
BenchmarkCondition-12                  507645       2460 ns/op       704 B/op       23 allocs/op
BenchmarkConcurrencyScaling/C1-12       82400      14740 ns/op     15687 B/op      393 allocs/op
BenchmarkConcurrencyScaling/C12-12      23158      52602 ns/op     15700 B/op      422 allocs/op
BenchmarkConcurrencyScaling/C48-12      17500      68434 ns/op     15907 B/op      453 allocs/op
BenchmarkGraphBuild/N32-12             348994       3418 ns/op      6284 B/op      230 allocs/op
BenchmarkGraphBuild/N128-12             92428      13300 ns/op     24856 B/op      904 allocs/op
BenchmarkGraphBuild/N512-12             22143      55325 ns/op    101709 B/op     3850 allocs/op
```

## Understanding Conditional Tasks

Conditional nodes in go-taskflow behave similarly to those in [taskflow-cpp](https://github.com/taskflow/taskflow). They participate in both conditional control and looping. To avoid common pitfalls, refer to the [Conditional Tasking documentation](https://taskflow.github.io/taskflow/ConditionalTasking.html).

## Executor Options

`NewExecutor` accepts functional options to configure behavior:

```go
executor := gtf.NewExecutor(1000,
    gtf.WithProfiler(), // enable flamegraph profiling
    gtf.WithTracer(),   // enable Chrome Trace recording
)
```

| Option | Description |
|:---|:---|
| `WithProfiler()` | Enable flamegraph profiling. Required before calling `executor.Profile()`. |
| `WithTracer()` | Enable Chrome Trace recording. Required before calling `executor.Trace()`. |

## Error Handling in go-taskflow

In Go, `errors` are values, and it is the user's responsibility to handle them appropriately. Only unrecovered `panic` events are managed by the framework. If a `panic` occurs, the entire parent graph is canceled, leaving the remaining tasks incomplete. This behavior may evolve in the future. If you have suggestions, feel free to share them.

To prevent interruptions caused by `panic`, you can handle them manually when registering tasks:

```go
tf.NewTask("not interrupt", func() {
    defer func() {
        if r := recover(); r != nil {
            // Handle the panic.
        }
    }()
    // User-defined logic.
})
```

## Visualizing Taskflows

To generate a visual representation of a taskflow, use the `Dump` method:

```go
if err := tf.Dump(os.Stdout); err != nil {
    log.Fatal(err)
}
```

The `Dump` method generates raw strings in DOT format. Use the `dot` tool to create a graph SVG. 

![dot](image/desc.svg)

## Profiling Taskflows

To profile a taskflow, first enable the profiler with `WithProfiler()`, then call `Profile`:

```go
executor := gtf.NewExecutor(1000, gtf.WithProfiler())
executor.Run(tf).Wait()

if err := executor.Profile(os.Stdout); err != nil {
    log.Fatal(err)
}
```

The `Profile` method generates raw strings in flamegraph format. Use the `flamegraph` tool to create a flamegraph SVG.

![flg](image/fl.svg)

## Tracing Taskflows

To trace a taskflow, first enable the tracer with `WithTracer()`, then call `Trace`:

```go
executor := gtf.NewExecutor(1000, gtf.WithTracer())
executor.Run(tf).Wait()

if err := executor.Trace(os.Stdout); err != nil {
    log.Fatal(err)
}
```

The `Trace` method outputs JSON in [Chrome Trace Event format](https://docs.google.com/document/d/1CvAClvFfyA5R-PhYUmn5OOQtYMH4h6I0nSsKchNAySU). Open it in `chrome://tracing` or [Perfetto UI](https://ui.perfetto.dev/) for visualization.

## Stargazer

[![Star History Chart](https://api.star-history.com/svg?repos=noneback/go-taskflow&type=Date)](https://star-history.com/#noneback/go-taskflow&Date)


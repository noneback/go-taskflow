package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"

	gotaskflow "github.com/noneback/go-taskflow"
)

// A simplified MapReduce word-count pipeline.
//
// DAG topology:
//
//	split_input → [map_0, map_1, map_2, map_3]
//	                     ↓ (shuffle by hash)
//	              [reduce_0, reduce_1]
//	                     ↓
//	                 merge_results
const (
	numMappers  = 4
	numReducers = 2
)

var (
	// intermediate results: mapOutputs[mapper][reducer]
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
the fox and the dog are friends
the dog jumps over the fox
brown dog lazy fox the quick brown dog
the lazy fox jumps over the quick dog
the dog and the fox play together
brown brown brown the fox the fox`

	// Split input into chunks, one per mapper
	var chunks [numMappers]string
	splitChunks := func() {
		words := strings.Fields(input)
		size := (len(words) + numMappers - 1) / numMappers
		for i := 0; i < numMappers; i++ {
			start := i * size
			end := start + size
			if end > len(words) {
				end = len(words)
			}
			chunks[i] = strings.Join(words[start:end], " ")
		}
	}

	executor := gotaskflow.NewExecutor(8, gotaskflow.WithProfiler())
	tf := gotaskflow.NewTaskFlow("word-count")

	// Phase 1: Split input
	splitTask := tf.NewTask("split_input", splitChunks)

	// Phase 2: Map — each mapper counts words and partitions by hash
	mapTasks := make([]*gotaskflow.Task, numMappers)
	for i := 0; i < numMappers; i++ {
		idx := i
		mapTasks[idx] = tf.NewTask(fmt.Sprintf("map_%d", idx), func() {
			localCounts := [numReducers]map[string]int{}
			for r := 0; r < numReducers; r++ {
				localCounts[r] = make(map[string]int)
			}
			for _, w := range strings.Fields(chunks[idx]) {
				localCounts[hashPartition(w)][w]++
			}
			mu.Lock()
			for r := 0; r < numReducers; r++ {
				mapOutputs[idx][r] = localCounts[r]
			}
			mu.Unlock()
		})
	}
	splitTask.Precede(mapTasks...)

	// Phase 3: Reduce — each reducer aggregates a hash partition
	reduceTasks := make([]*gotaskflow.Task, numReducers)
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
	// Every mapper feeds every reducer (shuffle)
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
		fmt.Println("Word Count Results:")
		for _, w := range keys {
			fmt.Printf("  %-12s %d\n", w, final[w])
		}
	})
	for _, rt := range reduceTasks {
		rt.Precede(mergeTask)
	}

	executor.Run(tf).Wait()

	if err := tf.Dump(os.Stdout); err != nil {
		log.Fatal(err)
	}
	if err := executor.Profile(os.Stdout); err != nil {
		log.Fatal(err)
	}
}

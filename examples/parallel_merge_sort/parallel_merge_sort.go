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
	chunks := 100
	chunkSize := 1000
	randomArr := make([][]int, chunks)
	sortedArr := make([]int, 0, chunks*chunkSize)
	mutex := &sync.Mutex{}

	for i := 0; i < chunks; i++ {
		for j := 0; j < chunkSize; j++ {
			randomArr[i] = append(randomArr[i], rand.Int())
		}
	}

	sortTasks := make([]*gtf.Task, chunks)
	tf := gtf.NewTaskFlow("merge sort")
	done := tf.NewTask("done", func() {
		if !slices.IsSorted(sortedArr) {
			log.Fatal("sorting failed")
		}
		fmt.Println("sorted successfully")
		fmt.Println("first 10:", sortedArr[:10])
	})

	for i := 0; i < chunks; i++ {
		idx := i
		sortTasks[idx] = tf.NewTask("sort_"+strconv.Itoa(idx), func() {
			arr := randomArr[idx]
			slices.Sort(arr)
			mutex.Lock()
			defer mutex.Unlock()
			sortedArr = mergeInto(sortedArr, arr)
		})
	}
	done.Succeed(sortTasks...)

	executor := gtf.NewExecutor(1000, gtf.WithProfiler())
	executor.Run(tf).Wait()

	if err := tf.Dump(os.Stdout); err != nil {
		log.Fatal(err)
	}
	if err := executor.Profile(os.Stdout); err != nil {
		log.Fatal(err)
	}
}

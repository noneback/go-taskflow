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
	done := gtf.NewTask("Done", func() {
		if !slices.IsSorted(sortedArr) {
			log.Fatal("Failed")
		}
		fmt.Println("Sorted")
		fmt.Println(sortedArr[:1000])
	})

	for i := 0; i < 10; i++ {
		sortTasks[i] = gtf.NewTask("sort_"+strconv.Itoa(i), func() {
			arr := radomArr[i]
			slices.Sort(arr)
			mutex.Lock()
			defer mutex.Unlock()
			sortedArr = mergeInto(sortedArr, arr)
		})

	}
	done.Succeed(sortTasks...)
	tf.Push(sortTasks...)
	tf.Push(done)

	executor := gtf.NewExecutor(1000)

	executor.Run(tf).Wait()

	if err := gtf.Visualize(tf, os.Stdout); err != nil {
		log.Fatal("V->", err)
	}

	if err := executor.Profile(os.Stdout); err != nil {
		log.Fatal("P->", err)
	}

}

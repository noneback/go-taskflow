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
	"github.com/noneback/go-taskflow/utils"
)

// merge sorted src to sorted dest
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
	pprof := utils.NewPprofUtils(utils.CPU, "./out.prof")
	pprof.StartProfile()
	defer pprof.StopProfile()

	size := 10000
	share := 1000
	randomArr := make([][]int, share)
	sortedArr := make([]int, 0, share*size)
	mutex := &sync.Mutex{}

	for i := 0; i < share; i++ {
		for j := 0; j < size; j++ {
			randomArr[i] = append(randomArr[i], rand.Int())
		}
	}

	sortTasks := make([]*gtf.Task, share)
	tf := gtf.NewTaskFlow("merge sort")
	done := tf.NewTask("Done", func() {
		if !slices.IsSorted(sortedArr) {
			log.Fatal("Failed")
		}
		fmt.Println("Sorted")
		fmt.Println(sortedArr[:1000])
	})

	for i := 0; i < share; i++ {
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

	executor := gtf.NewExecutor(1000000)

	executor.Run(tf).Wait()

	if err := tf.Dump(os.Stdout); err != nil {
		log.Fatal("V->", err)
	}

	if err := executor.Profile(os.Stdout); err != nil {
		log.Fatal("P->", err)
	}

}

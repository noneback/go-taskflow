package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	gotaskflow "github.com/noneback/go-taskflow"
)

// 配置参数结构体
type MRConfig struct {
	NumMappers  int
	NumReducers int
	ChunkSize   int
	TempDir     string
	OutputPath  string
}

type WordCount struct {
	Key   string
	Value int
}

type MapReduce struct {
	cfg        MRConfig
	executor   gotaskflow.Executor
	input      string
	mapOutputs [][]string
}

func NewMapReduce(input string, cfg MRConfig) *MapReduce {
	os.MkdirAll(cfg.TempDir, 0755)
	return &MapReduce{
		cfg:      cfg,
		executor: gotaskflow.NewExecutor(uint(runtime.NumCPU())),
		input:    input,
	}
}
func (mr *MapReduce) Run() {
	tf := gotaskflow.NewTaskFlow("wordcount")
	mapper := make([][]int, mr.cfg.NumMappers)
	// spilt doc
	splitTask := tf.NewTask("split_input", func() {
		chunks := splitString(mr.input, mr.cfg.ChunkSize)
		for i, chunk := range chunks {
			path := filepath.Join(mr.cfg.TempDir, fmt.Sprintf("input_%d.txt", i))
			if err := os.WriteFile(path, []byte(chunk), 0644); err != nil {
				log.Fatal(err)
			}
			mapper[i%mr.cfg.NumMappers] = append(mapper[i%mr.cfg.NumMappers], i)
		}
	})

	mapTasks := make([]*gotaskflow.Task, mr.cfg.NumMappers)
	mr.mapOutputs = make([][]string, mr.cfg.NumMappers)

	for i := 0; i < mr.cfg.NumMappers; i++ {
		// scan input
		idx := i
		mapTasks[idx] = tf.NewTask(fmt.Sprintf("map_%d", idx), func() {
			for _, val := range mapper[idx] {
				mr.processMap(idx, val)
			}
		})
	}
	splitTask.Precede(mapTasks...)

	reduceTasks := make([]*gotaskflow.Task, mr.cfg.NumReducers)
	for r := 0; r < mr.cfg.NumReducers; r++ {
		r := r
		reduceTasks[r] = tf.NewTask(fmt.Sprintf("reduce_%d", r), func() {
			log.Println("reduce:", r)
			mr.processReduce(r)
		})
	}

	for _, mapTask := range mapTasks {
		mapTask.Precede(reduceTasks...)
	}

	mergeTask := tf.NewTask("merge_results", mr.mergeResults)
	for _, rt := range reduceTasks {
		rt.Precede(mergeTask)
	}

	if err := tf.Dump(os.Stdout); err != nil {
		log.Fatal(err)
	}
	mr.executor.Run(tf).Wait()
}

func (mr *MapReduce) processMap(mapID, inputID int) {
	inputPath := filepath.Join(mr.cfg.TempDir, fmt.Sprintf("input_%d.txt", inputID))
	data, err := os.ReadFile(inputPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(mapID, "process", inputPath)

	intermediate := make([]map[string]int, mr.cfg.NumReducers)
	for i := range intermediate {
		intermediate[i] = make(map[string]int)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		word := scanner.Text()
		r := hash(word) % mr.cfg.NumReducers
		intermediate[r][word]++
	}

	outputs := make([]string, mr.cfg.NumReducers)
	for r := 0; r < mr.cfg.NumReducers; r++ {
		fpath := filepath.Join(mr.cfg.TempDir, fmt.Sprintf("map-%d-reduce-%d.json", mapID, r))
		file, err := os.OpenFile(fpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		enc := json.NewEncoder(file)
		for word, count := range intermediate[r] {
			enc.Encode(WordCount{word, count})
		}
		outputs[r] = fpath
	}
	mr.mapOutputs[mapID] = outputs
}

func (mr *MapReduce) processReduce(reduceID int) {
	var intermediate []WordCount

	for m := 0; m < mr.cfg.NumMappers; m++ {
		fpath := filepath.Join(mr.cfg.TempDir, fmt.Sprintf("map-%d-reduce-%d.json", m, reduceID))
		file, err := os.Open(fpath)
		if err != nil {
			continue
		}
		defer file.Close()

		dec := json.NewDecoder(file)
		for dec.More() {
			var wc WordCount
			if err := dec.Decode(&wc); err != nil {
				log.Fatal(err)
			}
			intermediate = append(intermediate, wc)
		}
	}

	results := make(map[string]int)
	for _, item := range intermediate {
		results[item.Key] += item.Value
	}

	outputPath := filepath.Join(mr.cfg.TempDir, fmt.Sprintf("reduce-out-%d.txt", reduceID))
	file, err := os.Create(outputPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	for word, count := range results {
		fmt.Fprintf(file, "%s\t%d\n", word, count)
	}
}

func (mr *MapReduce) mergeResults() {
	var finalResults map[string]int = make(map[string]int)

	for r := 0; r < mr.cfg.NumReducers; r++ {
		path := filepath.Join(mr.cfg.TempDir, fmt.Sprintf("reduce-out-%d.txt", r))
		file, err := os.Open(path)
		if err != nil {
			continue
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			var word string
			var count int
			fmt.Sscanf(scanner.Text(), "%s\t%d", &word, &count)
			finalResults[word] += count
		}
	}

	keys := make([]string, 0, len(finalResults))
	for k := range finalResults {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	outputFile, err := os.Create(mr.cfg.OutputPath)
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	for _, word := range keys {
		fmt.Fprintf(outputFile, "%s\t%d\n", word, finalResults[word])
	}
}

func splitString(s string, chunkSize int) []string {
	var chunks []string
	for len(s) > 0 {
		if len(s) < chunkSize {
			chunks = append(chunks, s)
			break
		}
		chunk := s[:chunkSize]
		if lastSpace := strings.LastIndex(chunk, " "); lastSpace != -1 {
			chunk = chunk[:lastSpace]
			s = s[lastSpace+1:]
		} else {
			s = s[chunkSize:]
		}
		chunks = append(chunks, chunk)
	}
	return chunks
}

func hash(s string) int {
	h := 0
	for _, c := range s {
		h = 31*h + int(c)
	}
	if h < 0 {
		h = -h
	}
	return h
}

func main() {
	log.SetFlags(log.Llongfile)
	input := `Navigating This Book Now that you know who you’ll be hearing from, the next logical step would be to find out what you’ll be hearing about, which brings us to the second thing I wanted to mention. There are conceptually two major parts to this book, each with four chapters, and each followed up by a chapter that stands relatively independently on its own.
The fun begins with Part I, The Beam Model (Chapters 1–4), which focuses on the high-level batch plus streaming data processing model originally developed for Google Cloud Dataflow, later donated to the Apache Software Foundation as Apache Beam, and also now seen in whole or in part across most other systems in the industry. It’s composed of four chapters: Chapter 1, Streaming 101, which covers the basics of stream processing, establishing some terminology, discussing the capabilities of streaming systems, distinguishing between two important domains of time (processing time and event time), and finally looking at some common data processing patterns.
Chapter 2, The What, Where, When, and How of Data Processing, which covers in detail the core concepts of robust stream processing over out-of-order data, each analyzed within the context of a concrete running example and with animated diagrams to highlight the dimension of time.
Chapter 3, Watermarks (written by Slava), which provides a deep survey of temporal progress metrics, how they are created, and how they propagate through pipelines. It ends by examining the details of two real-world watermark implementations.
Chapter 4, Advanced Windowing, which picks up where Chapter 2 left off, diving into some advanced windowing and triggering concepts like processing-time windows, sessions, and continuation triggers.
Between Parts I and II, providing an interlude as timely as the details contained therein are important, stands Chapter 5, Exactly-Once and Side Effects (written by Reuven). In it, he enumerates the challenges of providing end-to-end exactly-once (or effectively-once) processing semantics and walks through the implementation details of three different approaches to exactlyonce processing: Apache Flink, Apache Spark, and Google Cloud Dataflow.
Next begins Part II, Streams and Tables (Chapters 6–9), which dives deeper into the conceptual and investigates the lower-level “streams and tables” way of thinking about stream processing, recently popularized by some upstanding citizens in the Apache Kafka community but, of course, invented decades ago by folks in the database community, because wasn’t everything? It too is composed of four chapters: Chapter 6, Streams and Tables, which introduces the basic idea of streams and tables, analyzes the classic MapReduce approach through a streams-and-tables lens, and then constructs a theory of streams and tables sufficiently general to encompass the full breadth of the Beam Model (and beyond).
Chapter 7, The Practicalities of Persistent State, which considers the motivations for persistent state in streaming pipelines, looks at two common types of implicit state, and then analyzes a practical use case (advertising attribution) to inform the necessary characteristics of a general state management mechanism.
Chapter 8, Streaming SQL, which investigates the meaning of streaming within the context of relational algebra and SQL, contrasts the inherent stream and table biases within the Beam Model and classic SQL as they exist today, and proposes a set of possible paths forward toward incorporating robust streaming semantics in SQL.
Chapter 9, Streaming Joins, which surveys a variety of different join types, analyzes their behavior within the context of streaming, and finally looks in detail at a useful but ill-supported streaming join use case: temporal validity windows.
Finally, closing out the book is Chapter 10, The Evolution of Large-Scale Data Processing, which strolls through a focused history of the MapReduce lineage of data processing systems, examining some of the important contributions that have evolved streaming systems into what they are today.`

	// 配置参数
	cfg := MRConfig{
		NumMappers:  4,
		NumReducers: 2,
		ChunkSize:   300,
		TempDir:     "./mr-tmp",
		OutputPath:  "final-count.txt",
	}

	mr := NewMapReduce(input, cfg)
	mr.Run()

	fmt.Println("Final word count:")
	data, _ := os.ReadFile(cfg.OutputPath)
	fmt.Println(string(data))
}

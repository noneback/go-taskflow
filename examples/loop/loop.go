package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	gotaskflow "github.com/noneback/go-taskflow"
)

// A web crawler with retry and exponential backoff.
//
// DAG topology:
//
//	init → [fetch_pages] → parse → [validate_quality]
//	                                    ├─(pass)→ index → done
//	                                    └─(fail)→ backoff → fetch_pages (loop)
//
// Demonstrates:
//   - Cyclic flow with retry logic
//   - Exponential backoff
//   - Quality validation before committing results
const maxAttempts = 5

var (
	attempt    int
	pageCount  int
	qualityOK  bool
	backoffMs  int
)

func main() {
	executor := gotaskflow.NewExecutor(4, gotaskflow.WithProfiler())
	tf := gotaskflow.NewTaskFlow("crawl-pipeline")

	// Initialize the crawl job
	init := tf.NewTask("init", func() {
		attempt = 0
		pageCount = 0
		qualityOK = false
		backoffMs = 100
		fmt.Println("initializing crawl job...")
	})

	// Fetch pages — simulate fetching with random success rate
	fetchPages := tf.NewTask("fetch_pages", func() {
		attempt++
		// Simulate: each attempt fetches more pages
		pagesFetched := 10 + attempt*5 + rand.Intn(20)
		pageCount = pagesFetched
		fmt.Printf("[attempt %d/%d] fetched %d pages\n", attempt, maxAttempts, pagesFetched)
		time.Sleep(time.Duration(backoffMs) * time.Millisecond)
	})

	// Parse fetched pages
	parse := tf.NewTask("parse", func() {
		parsed := pageCount
		fmt.Printf("parsed %d pages, extracted %d links\n", parsed, parsed*3)
	})

	// Validate quality — pass if we got enough pages
	validateQuality := tf.NewCondition("validate_quality", func() uint {
		threshold := 40 // need at least 40 pages
		if pageCount >= threshold {
			qualityOK = true
			fmt.Printf("quality check PASSED (%d >= %d pages)\n", pageCount, threshold)
			return 0 // pass → go to index
		}
		fmt.Printf("quality check FAILED (%d < %d pages)\n", pageCount, threshold)
		if attempt >= maxAttempts {
			fmt.Println("max attempts reached, giving up")
			return 0 // force pass after max attempts
		}
		return 1 // fail → go to backoff and retry
	})

	// Exponential backoff before retry — must be a condition node to allow cyclic flow
	backoff := tf.NewCondition("backoff", func() uint {
		backoffMs *= 2
		fmt.Printf("backing off %dms before retry...\n", backoffMs)
		return 0 // always loop back to fetch_pages
	})

	// Index the parsed results
	index := tf.NewTask("index", func() {
		if qualityOK {
			fmt.Printf("indexing %d pages into search engine\n", pageCount)
		} else {
			fmt.Printf("indexing %d pages (best effort after %d attempts)\n", pageCount, attempt)
		}
	})

	// Final summary
	done := tf.NewTask("done", func() {
		fmt.Println("--- Crawl Summary ---")
		fmt.Printf("  Total attempts: %d\n", attempt)
		fmt.Printf("  Pages indexed:  %d\n", pageCount)
		fmt.Printf("  Quality:        %v\n", qualityOK)
	})

	// Wire up the DAG
	init.Precede(fetchPages)
	fetchPages.Precede(parse)
	parse.Precede(validateQuality)
	validateQuality.Precede(index, backoff)
	backoff.Precede(fetchPages) // loop back
	index.Precede(done)

	executor.Run(tf).Wait()

	if err := tf.Dump(os.Stdout); err != nil {
		log.Fatal(err)
	}
	if err := executor.Profile(os.Stdout); err != nil {
		log.Fatal(err)
	}
}

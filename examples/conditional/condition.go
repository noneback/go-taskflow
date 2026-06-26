package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	gotaskflow "github.com/noneback/go-taskflow"
)

// An event-driven log processing pipeline with severity-based routing.
//
// DAG topology:
//
//	ingest → classify → [severity_check]
//	                         ├─→ critical_path: alert → escalate → page_oncall
//	                         ├─→ warning_path:  enrich → deduplicate → log_warning
//	                         └─→ info_path:     compress → archive
const eventTypes = 3 // 0=critical, 1=warning, 2=info

var (
	severityLabels = []string{"critical", "warning", "info"}
	eventBatch     []map[string]string
	processedLog   []string
)

func main() {
	executor := gotaskflow.NewExecutor(8, gotaskflow.WithProfiler())
	tf := gotaskflow.NewTaskFlow("event-processor")

	// Phase 1: Ingest — simulate receiving a batch of events
	ingest := tf.NewTask("ingest", func() {
		sources := []string{"webhook", "syslog", "api-gateway", "kafka", "sqs"}
		eventBatch = make([]map[string]string, 0, 100)
		for i := 0; i < 100; i++ {
			severity := severityLabels[rand.Intn(3)]
			eventBatch = append(eventBatch, map[string]string{
				"id":       fmt.Sprintf("evt-%04d", i),
				"source":   sources[rand.Intn(len(sources))],
				"severity": severity,
				"msg":      fmt.Sprintf("event payload %d", i),
				"ts":       time.Now().Format(time.RFC3339),
			})
		}
		fmt.Printf("ingested %d events\n", len(eventBatch))
	})

	// Phase 2: Classify — tag events with metadata
	classify := tf.NewTask("classify", func() {
		for _, evt := range eventBatch {
			evt["region"] = fmt.Sprintf("us-%s", string("eastwest"[rand.Intn(2)*4:rand.Intn(2)*4+4]))
			evt["tagged"] = "true"
		}
		fmt.Println("classified all events")
	})
	ingest.Precede(classify)

	// Phase 3: Severity check — route by event severity
	severityCheck := tf.NewCondition("severity_check", func() uint {
		// Route based on the dominant severity in the batch
		counts := map[string]int{}
		for _, evt := range eventBatch {
			counts[evt["severity"]]++
		}
		if counts["critical"] > 20 {
			fmt.Printf("routing → critical_path (critical=%d)\n", counts["critical"])
			return 0
		} else if counts["warning"] > 20 {
			fmt.Printf("routing → warning_path (warning=%d)\n", counts["warning"])
			return 1
		}
		fmt.Printf("routing → info_path (info=%d)\n", counts["info"])
		return 2
	})
	classify.Precede(severityCheck)

	// Critical path: alert → escalate → page on-call
	criticalPath := tf.NewSubflow("critical_path", func(sf *gotaskflow.Subflow) {
		alert := sf.NewTask("alert", func() {
			critical := filterBySeverity("critical")
			fmt.Printf("  ALERT: %d critical events detected\n", len(critical))
		})
		escalate := sf.NewTask("escalate", func() {
			fmt.Println("  escalating to incident manager...")
		})
		pageOnCall := sf.NewTask("page_oncall", func() {
			fmt.Println("  paging on-call engineer via PagerDuty")
		})
		alert.Precede(escalate)
		escalate.Precede(pageOnCall)
	})

	// Warning path: enrich → deduplicate → log warning
	warningPath := tf.NewSubflow("warning_path", func(sf *gotaskflow.Subflow) {
		enrich := sf.NewTask("enrich", func() {
			warnings := filterBySeverity("warning")
			for _, evt := range warnings {
				evt["enriched_with"] = "historical_context"
			}
			fmt.Printf("  enriched %d warning events\n", len(warnings))
		})
		dedup := sf.NewTask("deduplicate", func() {
			fmt.Println("  deduplicating against known patterns...")
		})
		logWarn := sf.NewTask("log_warning", func() {
			fmt.Println("  logged warnings to monitoring dashboard")
		})
		enrich.Precede(dedup)
		dedup.Precede(logWarn)
	})

	// Info path: compress → archive
	infoPath := tf.NewSubflow("info_path", func(sf *gotaskflow.Subflow) {
		compress := sf.NewTask("compress", func() {
			info := filterBySeverity("info")
			fmt.Printf("  compressing %d info events\n", len(info))
		})
		archive := sf.NewTask("archive", func() {
			fmt.Println("  archived to cold storage (S3 Glacier)")
		})
		compress.Precede(archive)
	})

	severityCheck.Precede(criticalPath, warningPath, infoPath)

	executor.Run(tf).Wait()

	if err := tf.Dump(os.Stdout); err != nil {
		log.Fatal(err)
	}
	if err := executor.Profile(os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func filterBySeverity(severity string) []map[string]string {
	var result []map[string]string
	for _, evt := range eventBatch {
		if evt["severity"] == severity {
			result = append(result, evt)
		}
	}
	return result
}

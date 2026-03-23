package gotaskflow_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	gotaskflow "github.com/noneback/go-taskflow"
)

func TestExecutorWithTracer(t *testing.T) {
	executor := gotaskflow.NewExecutor(4, gotaskflow.WithTracer())
	tf := gotaskflow.NewTaskFlow("G")
	A, B, C :=
		tf.NewTask("A", func() { fmt.Println("A") }),
		tf.NewTask("B", func() { fmt.Println("B") }),
		tf.NewTask("C", func() { fmt.Println("C") })
	A.Precede(B)
	C.Precede(B)
	executor.Run(tf).Wait()

	var buf bytes.Buffer
	if err := executor.Trace(&buf); err != nil {
		t.Fatalf("Trace error: %v", err)
	}

	var events []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &events); err != nil {
		t.Fatalf("Trace output is not valid JSON: %v", err)
	}
	if len(events) != 3 {
		t.Errorf("expected 3 trace events, got %d", len(events))
	}
	t.Logf(buf.String())
}

func TestExecutorWithoutTracer(t *testing.T) {
	executor := gotaskflow.NewExecutor(4)
	tf := gotaskflow.NewTaskFlow("G")
	tf.NewTask("A", func() { fmt.Println("A") })
	executor.Run(tf).Wait()

	// Trace should return nil when tracer is not enabled
	if err := executor.Trace(os.Stdout); err != nil {
		t.Fatalf("Trace should return nil when disabled, got: %v", err)
	}
}

func TestExecutorWithProfiler(t *testing.T) {
	executor := gotaskflow.NewExecutor(4, gotaskflow.WithProfiler())
	tf := gotaskflow.NewTaskFlow("G")
	tf.NewTask("A", func() { fmt.Println("A") })
	tf.NewTask("B", func() { fmt.Println("B") })
	executor.Run(tf).Wait()

	var buf bytes.Buffer
	if err := executor.Profile(&buf); err != nil {
		t.Fatalf("Profile error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected profiler output, got empty")
	}
}

func TestExecutorWithoutProfiler(t *testing.T) {
	executor := gotaskflow.NewExecutor(4)
	tf := gotaskflow.NewTaskFlow("G")
	tf.NewTask("A", func() { fmt.Println("A") })
	executor.Run(tf).Wait()

	// Profile should return nil when profiler is not enabled
	if err := executor.Profile(os.Stdout); err != nil {
		t.Fatalf("Profile should return nil when disabled, got: %v", err)
	}
}

func TestExecutorTracePrint(t *testing.T) {
	executor := gotaskflow.NewExecutor(4, gotaskflow.WithTracer())
	tf := gotaskflow.NewTaskFlow("pipeline")

	A, B, C, D :=
		tf.NewTask("fetch", func() { time.Sleep(10 * time.Millisecond) }),
		tf.NewTask("parse", func() { time.Sleep(5 * time.Millisecond) }),
		tf.NewTask("process", func() { time.Sleep(8 * time.Millisecond) }),
		tf.NewTask("output", func() { time.Sleep(3 * time.Millisecond) })
	A.Precede(B)
	B.Precede(C)
	C.Precede(D)

	executor.Run(tf).Wait()

	t.Log("=== Chrome Trace JSON (paste into chrome://tracing or https://ui.perfetto.dev) ===")
	if err := executor.Trace(os.Stdout); err != nil {
		t.Fatalf("Trace error: %v", err)
	}
	t.Log("=== End of Trace ===")
}

// TestExecutorTraceComplex simulates a data-processing pipeline:
//
//	prepare
//	  ├─ read_config  ─┐
//	  └─ load_data    ─┴─ validate ─ check(cond)
//	                                    └─0─ sub_process (subflow)
//	                                           ├─ transform ─┐
//	                                           └─ enrich    ─┴─ aggregate ─ report
func TestExecutorTraceComplex(t *testing.T) {
	executor := gotaskflow.NewExecutor(8, gotaskflow.WithTracer())
	tf := gotaskflow.NewTaskFlow("data-pipeline")

	prepare := tf.NewTask("prepare", func() { time.Sleep(5 * time.Millisecond) })

	// Parallel stage: read config and load data concurrently.
	readConfig := tf.NewTask("read_config", func() { time.Sleep(8 * time.Millisecond) })
	loadData := tf.NewTask("load_data", func() { time.Sleep(12 * time.Millisecond) })

	// Merge: validate after both parallel tasks complete.
	validate := tf.NewTask("validate", func() { time.Sleep(4 * time.Millisecond) })

	// Condition: always takes the normal processing path (index 0).
	check := tf.NewCondition("check_quality", func() uint { return 0 })

	// Normal path: subflow runs transform and enrich in parallel, then aggregates.
	subProcess := tf.NewSubflow("sub_process", func(sf *gotaskflow.Subflow) {
		transform := sf.NewTask("transform", func() { time.Sleep(10 * time.Millisecond) })
		enrich := sf.NewTask("enrich", func() { time.Sleep(7 * time.Millisecond) })
		aggregate := sf.NewTask("aggregate", func() { time.Sleep(5 * time.Millisecond) })
		transform.Precede(aggregate)
		enrich.Precede(aggregate)
	})

	// Fallback path (not executed in this test).
	fallback := tf.NewTask("fallback", func() { time.Sleep(2 * time.Millisecond) })

	report := tf.NewTask("report", func() { time.Sleep(3 * time.Millisecond) })

	// Build dependencies.
	prepare.Precede(readConfig, loadData)
	readConfig.Precede(validate)
	loadData.Precede(validate)
	validate.Precede(check)
	check.Precede(subProcess, fallback) // 0 → subProcess, 1 → fallback
	subProcess.Precede(report)

	executor.Run(tf).Wait()

	var buf bytes.Buffer
	if err := executor.Trace(&buf); err != nil {
		t.Fatalf("Trace error: %v", err)
	}

	var events []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &events); err != nil {
		t.Fatalf("Trace output is not valid JSON: %v", err)
	}

	// prepare + read_config + load_data + validate + check + sub_process + transform + enrich + aggregate + report = 10
	if len(events) != 10 {
		t.Errorf("expected 10 trace events, got %d", len(events))
	}

	t.Log("=== Complex Pipeline Chrome Trace JSON ===")
	t.Log(buf.String())
	t.Log("=== Paste into https://ui.perfetto.dev to visualize ===")
}

func TestExecutorWithBothProfilerAndTracer(t *testing.T) {
	executor := gotaskflow.NewExecutor(4, gotaskflow.WithProfiler(), gotaskflow.WithTracer())
	tf := gotaskflow.NewTaskFlow("G")
	A, B :=
		tf.NewTask("A", func() { fmt.Println("A") }),
		tf.NewTask("B", func() { fmt.Println("B") })
	A.Precede(B)
	executor.Run(tf).Wait()

	var profBuf bytes.Buffer
	if err := executor.Profile(&profBuf); err != nil {
		t.Fatalf("Profile error: %v", err)
	}
	if profBuf.Len() == 0 {
		t.Error("expected profiler output, got empty")
	}

	var traceBuf bytes.Buffer
	if err := executor.Trace(&traceBuf); err != nil {
		t.Fatalf("Trace error: %v", err)
	}
	var events []map[string]interface{}
	if err := json.Unmarshal(traceBuf.Bytes(), &events); err != nil {
		t.Fatalf("Trace output is not valid JSON: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 trace events, got %d", len(events))
	}
}

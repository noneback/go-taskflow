package gotaskflow

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/noneback/go-taskflow/utils"
)

func TestProfilerAddSpan(t *testing.T) {
	profiler := newProfiler()
	mark := attr{
		typ:  nodeStatic,
		name: "test-span",
	}
	span := &span{
		extra: mark,
		begin: time.Now(),
		cost:  5 * time.Millisecond,
	}
	profiler.AddSpan(span)

	if len(profiler.spans) != 1 {
		t.Errorf("expected 1 span, got %d", len(profiler.spans))
	}

	if profiler.spans[mark] != span {
		t.Errorf("expected span to be added correctly, got %v", profiler.spans[mark])
	}
}

func TestSpanString(t *testing.T) {
	now := time.Now()
	span := &span{
		extra: attr{
			typ:  nodeStatic,
			name: "test-span",
		},
		begin: now,
		cost:  10 * time.Millisecond,
	}

	expected := "static,test-span,cost " + utils.NormalizeDuration(10*time.Millisecond)
	actual := span.String()

	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}

func TestProfilerDraw(t *testing.T) {
	profiler := newProfiler()
	now := time.Now()
	parentSpan := &span{
		extra: attr{
			typ:  nodeStatic,
			name: "parent",
		},
		begin: now,
		cost:  10 * time.Millisecond,
	}

	childSpan := &span{
		extra: attr{
			typ:  nodeStatic,
			name: "child",
		},
		begin:  now,
		cost:   5 * time.Millisecond,
		parent: parentSpan,
	}

	profiler.AddSpan(parentSpan)
	profiler.AddSpan(childSpan)

	var buf bytes.Buffer
	err := profiler.draw(&buf)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Errorf("expected output, got empty string")
	}

	expectedOutput := "static,parent,cost 10ms 10000\nstatic,parent,cost 10ms;static,child,cost 5ms 5000\n"
	expectedOutput2 := "static,parent,cost 10ms;static,child,cost 5ms 5000\nstatic,parent,cost 10ms 10000\n"
	if output != expectedOutput && output != expectedOutput2 {
		t.Errorf("expected output: %v\ngot: %v", expectedOutput, output)
	}
}

// TestProfilerAddSpanMerge tests that spans with the same attr are merged (cost is accumulated).
func TestProfilerAddSpanMerge(t *testing.T) {
	profiler := newProfiler()
	mark := attr{
		typ:  nodeStatic,
		name: "repeated-task",
	}

	// Add first span
	span1 := &span{
		extra: mark,
		begin: time.Now(),
		cost:  5 * time.Millisecond,
	}
	profiler.AddSpan(span1)

	if len(profiler.spans) != 1 {
		t.Fatalf("expected 1 span after first add, got %d", len(profiler.spans))
	}

	// Add second span with same attr - should merge
	span2 := &span{
		extra: mark,
		begin: time.Now(),
		cost:  3 * time.Millisecond,
	}
	profiler.AddSpan(span2)

	if len(profiler.spans) != 1 {
		t.Errorf("expected 1 span after merge, got %d", len(profiler.spans))
	}

	// Check that cost was accumulated
	mergedSpan := profiler.spans[mark]
	if mergedSpan.cost != 8*time.Millisecond {
		t.Errorf("expected merged cost 8ms, got %v", mergedSpan.cost)
	}
}

// TestProfilerDrawSkipsSubflow tests that subflow spans are not included in output.
func TestProfilerDrawSkipsSubflow(t *testing.T) {
	profiler := newProfiler()
	now := time.Now()

	// Add a subflow span - should be skipped in output
	subflowSpan := &span{
		extra: attr{
			typ:  nodeSubflow,
			name: "subflow-task",
		},
		begin: now,
		cost:  10 * time.Millisecond,
	}
	profiler.AddSpan(subflowSpan)

	// Add a static span - should be included
	staticSpan := &span{
		extra: attr{
			typ:  nodeStatic,
			name: "static-task",
		},
		begin: now,
		cost:  5 * time.Millisecond,
	}
	profiler.AddSpan(staticSpan)

	var buf bytes.Buffer
	err := profiler.draw(&buf)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Errorf("expected output, got empty string")
	}

	// Should not contain subflow-task
	if contains := bytes.Contains([]byte(output), []byte("subflow-task")); contains {
		t.Errorf("subflow span should not be in output, but found: %s", output)
	}

	// Should contain static-task
	if contains := bytes.Contains([]byte(output), []byte("static-task")); !contains {
		t.Errorf("static span should be in output, got: %s", output)
	}
}

// TestProfilerDrawWithDependents tests that spans with dependents are correctly formatted.
func TestProfilerDrawWithDependents(t *testing.T) {
	profiler := newProfiler()
	now := time.Now()

	span := &span{
		extra: attr{
			typ:  nodeStatic,
			name: "dependent-task",
		},
		begin:      now,
		cost:       5 * time.Millisecond,
		dependents: []string{"dep1", "dep2"},
	}
	profiler.AddSpan(span)

	var buf bytes.Buffer
	err := profiler.draw(&buf)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Errorf("expected output, got empty string")
	}

	// Output should contain the task name
	if contains := bytes.Contains([]byte(output), []byte("dependent-task")); !contains {
		t.Errorf("expected output to contain 'dependent-task', got: %s", output)
	}
}

// TestProfilerEmpty tests that drawing an empty profiler doesn't error.
func TestProfilerEmpty(t *testing.T) {
	profiler := newProfiler()

	var buf bytes.Buffer
	err := profiler.draw(&buf)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("expected empty output, got: %s", buf.String())
	}
}

// TestProfilerConcurrentAddSpan tests concurrent span addition.
func TestProfilerConcurrentAddSpan(t *testing.T) {
	profiler := newProfiler()
	var wg sync.WaitGroup
	n := 100

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			span := &span{
				extra: attr{
					typ:  nodeStatic,
					name: "concurrent-task",
				},
				begin: time.Now(),
				cost:  time.Duration(i+1) * time.Millisecond,
			}
			profiler.AddSpan(span)
		}(i)
	}
	wg.Wait()

	// All spans have same attr, so should be merged into one
	if len(profiler.spans) != 1 {
		t.Errorf("expected 1 span (merged), got %d", len(profiler.spans))
	}

	// Cost should be accumulated
	mergedSpan := profiler.spans[attr{typ: nodeStatic, name: "concurrent-task"}]
	expectedMin := time.Duration(n * (n + 1) / 2) * time.Millisecond // sum of 1..100
	if mergedSpan.cost < expectedMin {
		t.Errorf("expected merged cost >= %v, got %v", expectedMin, mergedSpan.cost)
	}
}

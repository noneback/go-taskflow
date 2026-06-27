package gotaskflow

import (
	"bytes"
	"sync"
	"testing"
	"time"
)

// TestObserverOpenSpan tests that openSpan creates a span with correct attributes.
func TestObserverOpenSpan(t *testing.T) {
	obs := newObserver()
	obs.withProfiler(newProfiler())

	node := &innerNode{
		name:       "test-task",
		Typ:        nodeStatic,
		dependents: []*innerNode{{name: "dep1"}, {name: "dep2"}},
	}

	s := obs.openSpan(node, nil)
	if s == nil {
		t.Fatal("expected non-nil span")
	}
	if s.extra.name != "test-task" {
		t.Errorf("expected name 'test-task', got %q", s.extra.name)
	}
	if s.extra.typ != nodeStatic {
		t.Errorf("expected type nodeStatic, got %q", s.extra.typ)
	}
	if s.parent != nil {
		t.Errorf("expected nil parent, got %v", s.parent)
	}
	if len(s.dependents) != 2 || s.dependents[0] != "dep1" || s.dependents[1] != "dep2" {
		t.Errorf("expected dependents [dep1, dep2], got %v", s.dependents)
	}
}

// TestObserverOpenSpanNil tests that openSpan returns nil when no profiler or tracer is set.
func TestObserverOpenSpanNil(t *testing.T) {
	obs := newObserver()

	node := &innerNode{
		name: "test-task",
		Typ:  nodeStatic,
	}

	s := obs.openSpan(node, nil)
	if s != nil {
		t.Errorf("expected nil span when no profiler/tracer, got %v", s)
	}
}

// TestObserverCloseSpan tests that closeSpan records the span to profiler.
func TestObserverCloseSpan(t *testing.T) {
	obs := newObserver()
	p := newProfiler()
	obs.withProfiler(p)

	node := &innerNode{
		name: "test-task",
		Typ:  nodeStatic,
	}

	s := obs.openSpan(node, nil)
	time.Sleep(1 * time.Millisecond) // ensure some duration
	obs.closeSpan(s, true)

	if len(p.spans) != 1 {
		t.Errorf("expected 1 span in profiler, got %d", len(p.spans))
	}
	if s.cost <= 0 {
		t.Errorf("expected positive cost, got %v", s.cost)
	}
}

// TestObserverCloseSpanNotOk tests that closeSpan does not record span when ok=false (panic).
func TestObserverCloseSpanNotOk(t *testing.T) {
	obs := newObserver()
	p := newProfiler()
	obs.withProfiler(p)

	node := &innerNode{
		name: "test-task",
		Typ:  nodeStatic,
	}

	s := obs.openSpan(node, nil)
	time.Sleep(1 * time.Millisecond)
	obs.closeSpan(s, false) // ok=false means panic occurred

	if len(p.spans) != 0 {
		t.Errorf("expected 0 spans in profiler (panicked task), got %d", len(p.spans))
	}
}

// TestObserverCloseSpanNil tests that closeSpan handles nil span gracefully.
func TestObserverCloseSpanNil(t *testing.T) {
	obs := newObserver()
	obs.withProfiler(newProfiler())

	// Should not panic
	obs.closeSpan(nil, true)
}

// TestObserverWithProfiler tests that withProfiler sets the profiler correctly.
func TestObserverWithProfiler(t *testing.T) {
	obs := newObserver()
	if obs.profiler != nil {
		t.Error("expected nil profiler initially")
	}

	p := newProfiler()
	obs.withProfiler(p)
	if obs.profiler != p {
		t.Error("profiler not set correctly")
	}
}

// TestObserverWithTracer tests that withTracer sets the tracer correctly.
func TestObserverWithTracer(t *testing.T) {
	obs := newObserver()
	if obs.tracer != nil {
		t.Error("expected nil tracer initially")
	}

	tr := newTracer()
	obs.withTracer(tr)
	if obs.tracer != tr {
		t.Error("tracer not set correctly")
	}
}

// TestObserverSpanParentChain tests that span parent chain is preserved.
func TestObserverSpanParentChain(t *testing.T) {
	obs := newObserver()
	obs.withProfiler(newProfiler())

	parentNode := &innerNode{name: "parent", Typ: nodeStatic}
	childNode := &innerNode{name: "child", Typ: nodeStatic}

	parentSpan := obs.openSpan(parentNode, nil)
	childSpan := obs.openSpan(childNode, parentSpan)

	if childSpan.parent != parentSpan {
		t.Error("child span should have parent span set")
	}
	if parentSpan.parent != nil {
		t.Error("parent span should have nil parent")
	}
}

// TestObserverWithBothProfilerAndTracer tests that both profiler and tracer work together.
func TestObserverWithBothProfilerAndTracer(t *testing.T) {
	obs := newObserver()
	p := newProfiler()
	tr := newTracer()
	obs.withProfiler(p)
	obs.withTracer(tr)

	node := &innerNode{name: "test-task", Typ: nodeStatic}
	s := obs.openSpan(node, nil)
	time.Sleep(1 * time.Millisecond)
	obs.closeSpan(s, true)

	if len(p.spans) != 1 {
		t.Errorf("expected 1 span in profiler, got %d", len(p.spans))
	}
	if len(tr.events) != 1 {
		t.Errorf("expected 1 event in tracer, got %d", len(tr.events))
	}
}

// TestObserverConcurrent tests concurrent use of observer.
func TestObserverConcurrent(t *testing.T) {
	obs := newObserver()
	p := newProfiler()
	tr := newTracer()
	obs.withProfiler(p)
	obs.withTracer(tr)

	var wg sync.WaitGroup
	n := 100
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			node := &innerNode{
				name: "task",
				Typ:  nodeStatic,
			}
			s := obs.openSpan(node, nil)
			time.Sleep(time.Microsecond)
			obs.closeSpan(s, true)
		}(i)
	}
	wg.Wait()

	// Note: profiler merges by attr, so may have fewer than n spans
	// But tracer should have all n events
	if len(tr.events) != n {
		t.Errorf("expected %d tracer events, got %d", n, len(tr.events))
	}
}

// TestObserverIntegrationWithExecutor tests observer works correctly with executor.
func TestObserverIntegrationWithExecutor(t *testing.T) {
	tf := NewTaskFlow("test-flow")
	a := tf.NewTask("A", func() {})
	b := tf.NewTask("B", func() {})
	a.Precede(b)

	exec := NewExecutor(4, WithProfiler())
	exec.Run(tf).Wait()

	var buf bytes.Buffer
	err := exec.Profile(&buf)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected profile output, got empty")
	}
}

// TestObserverIntegrationWithTracer tests observer works correctly with tracer.
func TestObserverIntegrationWithTracer(t *testing.T) {
	tf := NewTaskFlow("test-flow")
	a := tf.NewTask("A", func() {})
	b := tf.NewTask("B", func() {})
	a.Precede(b)

	exec := NewExecutor(4, WithTracer())
	exec.Run(tf).Wait()

	var buf bytes.Buffer
	err := exec.Trace(&buf)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected trace output, got empty")
	}
}

// TestObserverIntegrationWithBoth tests observer with both profiler and tracer.
func TestObserverIntegrationWithBoth(t *testing.T) {
	tf := NewTaskFlow("test-flow")
	a := tf.NewTask("A", func() {})
	b := tf.NewTask("B", func() {})
	a.Precede(b)

	exec := NewExecutor(4, WithProfiler(), WithTracer())
	exec.Run(tf).Wait()

	var profileBuf, traceBuf bytes.Buffer
	if err := exec.Profile(&profileBuf); err != nil {
		t.Errorf("unexpected profile error: %v", err)
	}
	if err := exec.Trace(&traceBuf); err != nil {
		t.Errorf("unexpected trace error: %v", err)
	}
	if profileBuf.Len() == 0 {
		t.Error("expected profile output")
	}
	if traceBuf.Len() == 0 {
		t.Error("expected trace output")
	}
}
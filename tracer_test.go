package gotaskflow

import (
	"bytes"
	"encoding/json"
	"sync"
	"testing"
	"time"
)

func TestTracerAddEvent(t *testing.T) {
	tr := newTracer()
	s := &span{
		extra: attr{typ: nodeStatic, name: "task-a"},
		begin: tr.start.Add(10 * time.Millisecond),
		cost:  5 * time.Millisecond,
	}
	tr.AddEvent(s)

	if len(tr.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(tr.events))
	}
	ev := tr.events[0]
	if ev.Name != "task-a" {
		t.Errorf("expected name 'task-a', got %q", ev.Name)
	}
	if ev.Cat != string(nodeStatic) {
		t.Errorf("expected cat %q, got %q", string(nodeStatic), ev.Cat)
	}
	if ev.Ph != "X" {
		t.Errorf("expected ph 'X', got %q", ev.Ph)
	}
	if ev.Dur != 5000 {
		t.Errorf("expected dur 5000, got %d", ev.Dur)
	}
}

func TestTracerWithParent(t *testing.T) {
	tr := newTracer()
	parent := &span{
		extra: attr{typ: nodeSubflow, name: "parent-flow"},
		begin: tr.start,
		cost:  20 * time.Millisecond,
	}
	child := &span{
		extra:  attr{typ: nodeStatic, name: "child-task"},
		begin:  tr.start.Add(5 * time.Millisecond),
		cost:   10 * time.Millisecond,
		parent: parent,
	}
	tr.AddEvent(child)

	if tr.events[0].Args == nil {
		t.Fatal("expected args with parent info")
	}
	if tr.events[0].Args["parent"] != "parent-flow" {
		t.Errorf("expected parent 'parent-flow', got %q", tr.events[0].Args["parent"])
	}
}

func TestTracerDraw(t *testing.T) {
	tr := newTracer()
	tr.AddEvent(&span{
		extra: attr{typ: nodeStatic, name: "a"},
		begin: tr.start,
		cost:  1 * time.Millisecond,
	})

	var buf bytes.Buffer
	if err := tr.draw(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var events []chromeTraceEvent
	if err := json.Unmarshal(buf.Bytes(), &events); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event in output, got %d", len(events))
	}
}

func TestTracerConcurrentAddEvent(t *testing.T) {
	tr := newTracer()
	var wg sync.WaitGroup
	n := 100
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			tr.AddEvent(&span{
				extra: attr{typ: nodeStatic, name: "task"},
				begin: tr.start.Add(time.Duration(i) * time.Millisecond),
				cost:  1 * time.Millisecond,
			})
		}(i)
	}
	wg.Wait()

	if len(tr.events) != n {
		t.Fatalf("expected %d events, got %d", n, len(tr.events))
	}

	// verify all tids are unique
	tids := make(map[int64]bool)
	for _, ev := range tr.events {
		if tids[ev.Tid] {
			t.Fatalf("duplicate tid: %d", ev.Tid)
		}
		tids[ev.Tid] = true
	}
}

package gotaskflow

import (
	"bytes"
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

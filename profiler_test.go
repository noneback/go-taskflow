package gotaskflow

import (
	"bytes"
	"testing"
	"time"
)

func TestProfilerStartStop(t *testing.T) {
	profiler := NewTracer()
	profiler.Start()
	time.Sleep(10 * time.Millisecond)
	profiler.Stop()

	if profiler.start.After(profiler.end) {
		t.Errorf("expected start time before end time, got start: %v, end: %v", profiler.start, profiler.end)
	}

	if profiler.start.IsZero() || profiler.end.IsZero() {
		t.Errorf("expected start and end times to be set, got start: %v, end: %v", profiler.start, profiler.end)
	}
}

func TestProfilerAddSpan(t *testing.T) {
	profiler := NewTracer()
	span := &span{
		extra: attr{
			typ:     NodeStatic,
			success: true,
			name:    "test-span",
		},
		begin: time.Now(),
		end:   time.Now().Add(5 * time.Millisecond),
	}
	profiler.AddSpan(span)

	if len(profiler.spans) != 1 {
		t.Errorf("expected 1 span, got %d", len(profiler.spans))
	}

	if profiler.spans[0] != span {
		t.Errorf("expected span to be added correctly, got %v", profiler.spans[0])
	}
}

func TestSpanString(t *testing.T) {
	span := &span{
		extra: attr{
			typ:     NodeStatic,
			success: true,
			name:    "test-span",
		},
		begin: time.Now(),
		end:   time.Now().Add(10 * time.Millisecond),
	}

	expected := "static,test-span,cost 10000ns"
	actual := span.String()

	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}

func TestProfilerDraw(t *testing.T) {
	profiler := NewTracer()
	parentSpan := &span{
		extra: attr{
			typ:     NodeStatic,
			success: true,
			name:    "parent",
		},
		begin: time.Now(),
		end:   time.Now().Add(10 * time.Millisecond),
	}

	childSpan := &span{
		extra: attr{
			typ:     NodeStatic,
			success: true,
			name:    "child",
		},
		begin:  time.Now(),
		end:    time.Now().Add(5 * time.Millisecond),
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

	expectedOutput := "static,parent,cost 10000ns 10000\nstatic,parent,cost 10000ns;static,child,cost 5000ns 5000\n"
	if output != expectedOutput {
		t.Errorf("expected output: %v\ngot: %v", []byte(expectedOutput), []byte(output))
	}
}

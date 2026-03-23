package gotaskflow

import (
	"encoding/json"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

// tracer records task execution events and exports them in Chrome Trace Event Format.
// The output can be visualized in Chrome's chrome://tracing or Perfetto UI (https://ui.perfetto.dev).
type tracer struct {
	events []chromeTraceEvent
	mu     sync.Mutex
	start  time.Time
	tidGen atomic.Int64
}

// chromeTraceEvent represents a single trace event following Chrome Trace Event Format.
type chromeTraceEvent struct {
	Name string            `json:"name"`
	Cat  string            `json:"cat"`
	Ph   string            `json:"ph"`
	Ts   int64             `json:"ts"`
	Dur  int64             `json:"dur"`
	Pid  int               `json:"pid"`
	Tid  int64             `json:"tid"`
	Args map[string]string `json:"args,omitempty"`
}

func newTracer() *tracer {
	return &tracer{
		events: make([]chromeTraceEvent, 0, 64),
		start:  time.Now(),
	}
}

// AddEvent records a task execution event from the given span.
func (t *tracer) AddEvent(s *span) {
	t.mu.Lock()
	defer t.mu.Unlock()

	ev := chromeTraceEvent{
		Name: s.extra.name,
		Cat:  string(s.extra.typ),
		Ph:   "X",
		Ts:   s.begin.Sub(t.start).Microseconds(),
		Dur:  s.cost.Microseconds(),
		Pid:  0,
		Tid:  t.tidGen.Add(1),
	}

	// Build args with optional parent and dependents
	args := make(map[string]string)
	if s.parent != nil {
		args["parent"] = s.parent.extra.name
	}
	if len(s.dependents) > 0 {
		// Store as comma-separated string for simplicity
		deps := ""
		for i, d := range s.dependents {
			if i > 0 {
				deps += ","
			}
			deps += d
		}
		args["dependents"] = deps
	}
	if len(args) > 0 {
		ev.Args = args
	}

	t.events = append(t.events, ev)
}

func (t *tracer) draw(w io.Writer) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(t.events)
}

// traceRecord is an immutable snapshot of task execution events produced by a tracer.
// It represents the observed execution result of a TaskFlow run.
type traceRecord []chromeTraceEvent

// snapshot returns an immutable copy of all recorded trace events.
func (t *tracer) snapshot() traceRecord {
	t.mu.Lock()
	defer t.mu.Unlock()
	cp := make(traceRecord, len(t.events))
	copy(cp, t.events)
	return cp
}

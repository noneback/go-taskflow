package gotaskflow

import (
	"fmt"
	"io"
	"sync"
	"time"
)

type Profiler struct {
	start, end time.Time
	spans      []*span
	mu         *sync.Mutex
}

func NewTracer() *Profiler {
	return &Profiler{
		spans: make([]*span, 0),
		mu:    &sync.Mutex{},
	}
}

func (t *Profiler) Start() {
	t.start = time.Now()
}

func (t *Profiler) Stop() {
	t.end = time.Now()
}

func (t *Profiler) AddSpan(s *span) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.spans = append(t.spans, s)
}

type attr struct {
	typ     NodeType
	success bool // 0 for success, 1 for abnormal
	name    string
}

type span struct {
	extra      attr
	begin, end time.Time
	parent     *span
}

func (s *span) String() string {
	return fmt.Sprintf("%s,%s,cost %vns", s.extra.typ, s.extra.name, s.end.Sub(s.begin).Microseconds())
}

func (t *Profiler) draw(w io.Writer) {
	for _, s := range t.spans {
		path := ""
		if s.extra.typ == NodeStatic {
			path = s.String()
			cur := s

			for cur.parent != nil {
				path = cur.parent.String() + ";" + path
				cur = cur.parent
			}
			msg := fmt.Sprintf("%s %v\n", path, s.end.Sub(s.begin).Microseconds())

			if _, err := w.Write([]byte(msg)); err != nil {
				panic("tracer failed:" + err.Error())
			}
		}

	}
}

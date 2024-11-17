package gotaskflow

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/noneback/go-taskflow/utils"
)

type profiler struct {
	spans map[attr]*span

	mu *sync.Mutex
}

func newProfiler() *profiler {
	return &profiler{
		spans: make(map[attr]*span),
		mu:    &sync.Mutex{},
	}
}

func (t *profiler) AddSpan(s *span) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if span, ok := t.spans[s.extra]; ok {
		s.cost += span.cost
	}
	t.spans[s.extra] = s
}

type attr struct {
	typ  nodeType
	name string
}

type span struct {
	extra  attr
	begin  time.Time
	cost   time.Duration
	parent *span
}

func (s *span) String() string {
	return fmt.Sprintf("%s,%s,cost %v", s.extra.typ, s.extra.name, utils.NormalizeDuration(s.cost))
}

func (t *profiler) draw(w io.Writer) error {
	// compact spans base on name
	for _, s := range t.spans {
		path := ""
		if s.extra.typ != nodeSubflow {
			path = s.String()
			cur := s

			for cur.parent != nil {
				path = cur.parent.String() + ";" + path
				cur = cur.parent
			}
			msg := fmt.Sprintf("%s %v\n", path, s.cost.Microseconds())

			if _, err := w.Write([]byte(msg)); err != nil {
				return fmt.Errorf("write profile -> %w", err)
			}
		}

	}
	return nil
}

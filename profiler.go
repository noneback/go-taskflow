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
	extra      attr
	begin      time.Time
	cost       time.Duration
	parent     *span
	dependents []string // names of predecessor tasks
}

func (s *span) String() string {
	return fmt.Sprintf("%s,%s,cost %v", s.extra.typ, s.extra.name, utils.NormalizeDuration(s.cost))
}

func (t *profiler) draw(w io.Writer) error {
	// compact spans base on name
	t.mu.Lock()
	defer t.mu.Unlock()

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

// observer 观察任务执行并记录 span
type observer struct {
	profiler *profiler
	tracer   *tracer
}

func newObserver() *observer {
	return &observer{}
}

// openSpan 创建 span（如果需要观察）
func (o *observer) openSpan(node *innerNode, parent *span) *span {
	if o.profiler == nil && o.tracer == nil {
		return nil
	}
	return &span{
		extra:      attr{typ: node.Typ, name: node.name},
		begin:      time.Now(),
		parent:     parent,
		dependents: getDependentNames(node),
	}
}

// closeSpan 结束 span 并记录
func (o *observer) closeSpan(s *span, ok bool) {
	if s == nil {
		return
	}
	s.cost = time.Since(s.begin)
	if ok && o.profiler != nil {
		o.profiler.AddSpan(s)
	}
	if o.tracer != nil {
		o.tracer.AddEvent(s)
	}
}

func (o *observer) withProfiler(p *profiler) {
	o.profiler = p
}

func (o *observer) withTracer(t *tracer) {
	o.tracer = t
}

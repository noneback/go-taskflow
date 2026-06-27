package gotaskflow

// Option configures executor behavior.
type Option func(*innerExecutorImpl)

// WithProfiler enables flame graph profiling for task execution analysis.
func WithProfiler() Option {
	return func(e *innerExecutorImpl) {
		e.obs.withProfiler(newProfiler())
	}
}

// WithTracer enables Chrome Trace Event recording for task execution analysis.
// The trace output can be visualized in chrome://tracing or Perfetto UI.
func WithTracer() Option {
	return func(e *innerExecutorImpl) {
		e.obs.withTracer(newTracer())
	}
}

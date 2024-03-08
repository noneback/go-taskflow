package gotaskflow

type Executor interface {
	Wait()
	// WaitForAll()
	Run(tf *TaskFlow) Executor
	// Observe()
}

type ExecutorImpl struct {
	concurrency int
}

func NewExecutor(concurrency int) Executor {
	return &ExecutorImpl{
		concurrency: concurrency,
	}
}

func (e *ExecutorImpl) Run(tf *TaskFlow) Executor {
	return e
}
func (e *ExecutorImpl) Wait() {}

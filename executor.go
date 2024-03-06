package gotaskflow

type Executor interface {
	Wait()
	// WaitForAll()
	Run()
	// Observe()
}

type ExecutorImpl struct {
	concurrency int
}

func (e *ExecutorImpl) Run(tf *TaskFlow){}
func (e *ExecutorImpl) Wait(){}
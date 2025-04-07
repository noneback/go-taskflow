package gotaskflow

import "io"

// TaskFlow represents a series of tasks
type TaskFlow struct {
	graph  *eGraph
	frozen bool
}

// Reset resets taskflow
func (tf *TaskFlow) Reset() {
	// tf.graph.reset()
	tf.frozen = false
}

// NewTaskFlow returns a taskflow struct
func NewTaskFlow(name string) *TaskFlow {
	return &TaskFlow{
		graph: newGraph(name),
	}
}

// Push pushs all task into taskflow
func (tf *TaskFlow) push(tasks ...*Task) {
	if tf.frozen {
		panic("Taskflow is frozen, cannot new tasks")
	}

	for _, task := range tasks {
		tf.graph.push(task.node)
	}
}

func (tf *TaskFlow) Name() string {
	return tf.graph.name
}

// NewStaticTask returns a attached static task
func (tf *TaskFlow) NewTask(name string, f func()) *Task {
	task := &Task{
		node: builder.NewStatic(name, f),
	}
	tf.push(task)
	return task
}

// NewSubflow returns a attached subflow task
// NOTICE: instantiate will be invoke only once to instantiate itself
func (tf *TaskFlow) NewSubflow(name string, instantiate func(sf *Subflow)) *Task {
	task := &Task{
		node: builder.NewSubflow(name, instantiate),
	}
	tf.push(task)
	return task
}

// NewCondition returns a attached condition task. NOTICE: The predict func return value determines its successor.
func (tf *TaskFlow) NewCondition(name string, predict func() uint) *Task {
	task := &Task{
		node: builder.NewCondition(name, predict),
	}
	tf.push(task)
	return task
}

// Dump writes graph dot data into writer
func (tf *TaskFlow) Dump(writer io.Writer) error {
	return vizer.Visualize(tf, writer)
}

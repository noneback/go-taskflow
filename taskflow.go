package gotaskflow

// TaskFlow represents a series of tasks organized in DAG.
// Tasks must be pushed via a `Push` api.
type TaskFlow struct {
	name   string
	graph  *eGraph
	forzen bool
}

// Reset resets taskflow
func (tf *TaskFlow) Reset() {
	tf.graph.reset()
	tf.forzen = false
}

// NewTaskFlow returns a taskflow struct
func NewTaskFlow(name string) *TaskFlow {
	return &TaskFlow{
		graph: newGraph(name),
	}
}

// Push pushs all task into taskflow
func (tf *TaskFlow) push(tasks ...*Task) {
	if tf.forzen {
		panic("Taskflow is frozen, cannot new tasks")
	}

	for _, task := range tasks {
		tf.graph.push(task.node)
	}
}

func (tf *TaskFlow) Name() string {
	return tf.name
}

// NewStaticTask returns a static task
func (tf *TaskFlow) NewTask(name string, f func()) *Task {
	task := &Task{
		node: builder.NewStatic(name, f),
	}
	tf.push(task)
	return task
}

// NewSubflow returns a subflow task
func (tf *TaskFlow) NewSubflow(name string, f func(sf *Subflow)) *Task {
	task := &Task{
		node: builder.NewSubflow(name, f),
	}
	tf.push(task)
	return task
}

// NewCondition returns a condition task. The predict func return value determines its successor.
func (tf *TaskFlow) NewCondition(name string, predict func() uint) *Task {
	task := &Task{
		node: builder.NewCondition(name, predict),
	}
	tf.push(task)
	return task
}

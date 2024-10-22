package gotaskflow

// TaskFlow represents a series of tasks organized in DAG.
// Tasks must be pushed via a `Push` api.
type TaskFlow struct {
	name  string
	graph *eGraph
}

// Reset resets taskflow
func (tf *TaskFlow) Reset() {
	tf.graph.reset()
}

// NewTaskFlow returns a taskflow struct
func NewTaskFlow(name string) *TaskFlow {
	return &TaskFlow{
		graph: newGraph(name),
	}
}

// Push pushs all task into taskflow
func (tf *TaskFlow) Push(tasks ...*Task) {
	for _, task := range tasks {
		tf.graph.push(task.node)
	}
}

func (tf *TaskFlow) Name() string {
	return tf.name
}

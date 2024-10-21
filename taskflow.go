package gotaskflow

type TaskFlow struct {
	name  string
	graph *eGraph
}

func (tf *TaskFlow) Reset() {
	tf.graph.reset()
}

func NewTaskFlow(name string) *TaskFlow {
	return &TaskFlow{
		graph: newGraph(name),
	}
}

func (tf *TaskFlow) Push(tasks ...*Task) {
	for _, task := range tasks {
		tf.graph.push(task.node)
	}
}

func (tf *TaskFlow) Name() string {
	return tf.name
}

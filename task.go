package gotaskflow

type TaskInterface interface {
	Name()
	Precede(task TaskInterface)
	Succeed(task TaskInterface)
}

type Task struct {
	node *Node
}

func NewTask(name string, f func()) *Task {
	return &Task{
		node: FlowBuilder.NewStatic(name, f),
	}
}

func NewSubflow(name string, f func(sf *Subflow)) *Task {
	return &Task{
		node: FlowBuilder.NewSubflow(name, f),
	}
}

// task deps on T
func (t *Task) Precede(task *Task) {
	t.node.precede(task.node) // TODO: 如何去重
}

// T deps on task
func (t *Task) Succeed(task *Task) {
	task.node.precede(t.node)
}

func (t *Task) Name() string {
	return t.node.name
}

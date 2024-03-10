package gotaskflow

type TaskInterface interface {
	Name() string
	Node() kNode
	Precede(task TaskInterface)
	Succeed(task TaskInterface)
}

type Task struct {
	node *RawNode
}

func NewTask(name string, f TaskHandle) *Task {
	return &Task{
		node: newNodeWithHandle(name, f),
	}
}

// task deps on T
func (t *Task) Precede(task TaskInterface) {
	t.node.Precede(task.Node()) // TODO: 如何去重
}

func (t *Task) Node() kNode {
	return t.node
}

// T deps on task
func (t *Task) Succeed(task TaskInterface) {
	task.Node().Precede(t.node)
}

func (t *Task) Name() string {
	return t.node.name
}

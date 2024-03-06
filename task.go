package gotaskflow

type TaskInterface interface {
	Name()
	Precede(task TaskInterface)
	Succeed(task TaskInterface)
}

type Task struct {
	name string
	node *Node
}

func NewTask(name string, f TaskHandle) *Task {
	return &Task{
		name: name,
		node: newNodeWithHandle(f),
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
	return t.name
}

package gotaskflow

type Task struct {
	node *innerNode
}

func NewTask(name string, f func()) *Task {
	return &Task{
		node: builder.NewStatic(name, f),
	}
}

func NewSubflow(name string, f func(sf *Subflow)) *Task {
	return &Task{
		node: builder.NewSubflow(name, f),
	}
}

func NewCondition(name string, f func() uint) *Task {
	return &Task{
		node: builder.NewCondition("cond", f),
	}
}

// task deps on T
func (t *Task) Precede(tasks ...*Task) {
	if cond, ok := t.node.ptr.(*Condition); ok {
		for i, task := range tasks {
			cond.mapper[uint(i)] = task.node
		}
	}

	for _, task := range tasks {
		t.node.precede(task.node) // TODO: 如何去重
	}
}

// T deps on task
func (t *Task) Succeed(tasks ...*Task) {
	for _, task := range tasks {
		task.node.precede(t.node)
	}
}

func (t *Task) Name() string {
	return t.node.name
}

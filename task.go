package gotaskflow

// Basic component of Taskflow
type Task struct {
	node *innerNode
}

// NewStaticTask
func NewTask(name string, f func()) *Task {
	return &Task{
		node: builder.NewStatic(name, f),
	}
}

// NewSubflow
func NewSubflow(name string, f func(sf *Subflow)) *Task {
	return &Task{
		node: builder.NewSubflow(name, f),
	}
}

// NewCondition
func NewCondition(name string, f func() uint) *Task {
	return &Task{
		node: builder.NewCondition("cond", f),
	}
}

// Precede: Tasks all depend on *this*
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

// Succeed: *this* deps on tasks
func (t *Task) Succeed(tasks ...*Task) {
	for _, task := range tasks {
		task.node.precede(t.node)
	}
}

func (t *Task) Name() string {
	return t.node.name
}

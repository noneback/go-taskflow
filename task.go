package gotaskflow

// Basic component of Taskflow
type Task struct {
	node *innerNode
}

// Precede: Tasks all depend on *this*.
// In Addition, order of tasks is correspond to predict result, ranging from 0...len(tasks)
func (t *Task) Precede(tasks ...*Task) {
	if cond, ok := t.node.ptr.(*Condition); ok {
		for _, task := range tasks {
			index := len(cond.mapper)
			cond.mapper[uint(index)] = task.node
		}
	}

	for _, task := range tasks {
		t.node.precede(task.node)
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

// Priority sets task's sche priority. Noted that due to goroutine concurrent mode, it can only assure task schedule priority, rather than its execution.
func (t *Task) Priority(p TaskPriority) *Task {
	t.node.priority = p
	return t
}

// Task sche priority
type TaskPriority uint

const (
	HIGH = TaskPriority(iota + 1)
	NORMAL
	LOW
)

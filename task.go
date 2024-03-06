package gotaskflow

type TaskInterface interface {
	Name()
}

type Task struct {
}

func (t *Task) Precede(task *Task) {}
func (t *Task) Succeed(task *Task) {}
func (t *Task) Priority()          {}
func (t *Task) Name()              {}

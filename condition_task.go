package gotaskflow

type ConditionTask struct {
	node *ConditionNode
}

func NewConditionTask(name string, f ConditionTaskHandle) *ConditionTask {
	return &ConditionTask{
		node: newCondNodeWithHandle(name, f),
	}
}

// task deps on T
func (t *ConditionTask) Precede(task TaskInterface) {
	t.node.Precede(task.Node()) // TODO: 如何去重
}

// T deps on task
func (t *ConditionTask) Succeed(task TaskInterface) {
	task.Node().Precede(t.node)
}

func (t *ConditionTask) Name() string {
	return t.node.name
}

func (t *ConditionTask) Node() kNode {
	return t.node
}

func (t *ConditionTask) SetMapper(mapper map[int]TaskInterface) {
	for k, v := range mapper {
		t.node.mapper[k] = v.Node()
	}
}

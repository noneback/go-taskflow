package gotaskflow

import "fmt"

var builder = flowBuilder{}

type flowBuilder struct{}

// Condition Wrapper
type Condition struct {
	handle func() uint
	mapper map[uint]*innerNode
}

// Static Wrapper
type Static struct {
	handle func()
}

// Subflow Wrapper
type Subflow struct {
	handle func(sf *Subflow)
	g      *eGraph
}

// only for visualizer
func (sf *Subflow) instancelize() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("instancelize may failed or paniced")
		}
	}()

	if sf.g.instancelized {
		return nil
	}
	sf.g.instancelized = true
	sf.handle(sf)
	return nil
}

// Push pushs all tasks into subflow
func (sf *Subflow) push(tasks ...*Task) {
	for _, task := range tasks {
		sf.g.push(task.node)
	}
}

func (tf *flowBuilder) NewStatic(name string, f func()) *innerNode {
	node := newNode(name)
	node.ptr = &Static{
		handle: f,
	}
	node.Typ = nodeStatic
	return node
}

func (fb *flowBuilder) NewSubflow(name string, f func(sf *Subflow)) *innerNode {
	node := newNode(name)
	node.ptr = &Subflow{
		handle: f,
		g:      newGraph(name),
	}
	node.Typ = nodeSubflow
	return node
}

func (fb *flowBuilder) NewCondition(name string, f func() uint) *innerNode {
	node := newNode(name)
	node.ptr = &Condition{
		handle: f,
		mapper: make(map[uint]*innerNode),
	}
	node.Typ = nodeCondition
	return node
}

// NewStaticTask returns a static task
func (sf *Subflow) NewTask(name string, f func()) *Task {
	task := &Task{
		node: builder.NewStatic(name, f),
	}
	sf.push(task)
	return task
}

// NewSubflow returns a subflow task
func (sf *Subflow) NewSubflow(name string, f func(sf *Subflow)) *Task {
	task := &Task{
		node: builder.NewSubflow(name, f),
	}
	sf.push(task)
	return task
}

// NewCondition returns a condition task. The predict func return value determines its successor.
func (sf *Subflow) NewCondition(name string, predict func() uint) *Task {
	task := &Task{
		node: builder.NewCondition(name, predict),
	}
	sf.push(task)
	return task
}

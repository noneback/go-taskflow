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
func (sf *Subflow) Push(tasks ...*Task) {
	for _, task := range tasks {
		sf.g.push(task.node)
	}
}

func (fb *flowBuilder) NewStatic(name string, f func()) *innerNode {
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

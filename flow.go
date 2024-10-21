package gotaskflow

import "fmt"

var builder = flowBuilder{}

type flowBuilder struct{}

type Condition struct {
	handle func() int
	mapper map[int]*Node
}

type Static struct {
	handle func()
}

type Subflow struct {
	handle func(sf *Subflow)
	g      *Graph
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

func (sf *Subflow) Push(tasks ...*Task) {
	for _, task := range tasks {
		sf.g.push(task.node)
	}
}

func (fb *flowBuilder) NewStatic(name string, f func()) *Node {
	node := newNode(name)
	node.ptr = &Static{
		handle: f,
	}
	node.Typ = NodeStatic
	return node
}

func (fb *flowBuilder) NewSubflow(name string, f func(sf *Subflow)) *Node {
	node := newNode(name)
	node.ptr = &Subflow{
		handle: f,
		g:      newGraph(name),
	}
	node.Typ = NodeSubflow
	return node
}

func (fb *flowBuilder) NewCondition(name string, f func() int) *Node {
	node := newNode(name)
	node.ptr = &Condition{
		handle: f,
		mapper: make(map[int]*Node),
	}
	node.Typ = NodeCondition
	return node
}

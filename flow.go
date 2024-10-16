package gotaskflow

import "fmt"

var FlowBuilder = flowBuilder{}

type flowBuilder struct{}

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
			fmt.Println("instancelize may failed or paniced")
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

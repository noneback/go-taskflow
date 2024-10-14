package gotaskflow

var FlowBuilder = flowBuilder{}

type flowBuilder struct{}

type Static struct {
	handle func()
}

type Subflow struct {
	handle func(sf *Subflow)
	g      *Graph
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

package gotaskflow

import (
	"errors"
	"fmt"
	"io"

	"github.com/awalterschulze/gographviz"
)

var (
	ErrTaskFlowIsCyclic = errors.New("task flow is cyclic, not support")
)

type TaskFlow struct {
	name  string
	graph *Graph
}

func NewTaskFlow(name string) *TaskFlow {
	return &TaskFlow{
		graph: newGraph(name),
	}
}

func (tf *TaskFlow) Push(tasks ...*Task) {
	for _, task := range tasks {
		tf.graph.push(task.node)
	}
}

func (tf *TaskFlow) Name() string {
	return tf.name
}

// TODO: some other suger to set graph dependency, current not importent



func (tf *TaskFlow) Visualize(writer io.Writer) error {
	nodes, ok := tf.graph.TopologicalSort()
	if !ok {
		return ErrTaskFlowIsCyclic
	}
	vGraph := gographviz.NewGraph()
	vGraph.Directed = true

	for _, node := range nodes {
		if err := vGraph.AddNode(tf.graph.name, node.name, nil); err != nil {
			return fmt.Errorf("add node %v -> %w", node.name, err)
		}
	}

	for _, node := range nodes {
		for _, deps := range node.dependents {
			if err := vGraph.AddEdge(deps.name, node.name, true, nil); err != nil {
				return fmt.Errorf("add edge %v - %v -> %w", deps.name, node.name, err)
			}
		}
	}

	if n, err := writer.Write(unsafeToBytes(vGraph.String())); err != nil {
		return fmt.Errorf("write at %v -> %w", n, err)
	}

	return nil
}

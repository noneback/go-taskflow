package gotaskflow

import (
	"errors"
	"fmt"
	"io"

	"github.com/awalterschulze/gographviz"
	"github.com/noneback/go-taskflow/utils"
)

var (
	ErrGraphIsCyclic = errors.New("graph is cyclic, not support")
)

type TaskFlow struct {
	name  string
	graph *Graph
}

func (tf *TaskFlow) JoinCounter() int {
	return tf.graph.JoinCounter()
}

func NewTaskFlow(name string) *TaskFlow {
	return &TaskFlow{
		graph: newGraph(name),
	}
}

func (tf *TaskFlow) Push(tasks ...*Task) {
	for _, task := range tasks {
		tf.graph.Push(task.node)
	}
}

func (tf *TaskFlow) Name() string {
	return tf.name
}

func visualizeG(g *Graph) (*gographviz.Graph, error) {
	nodes, ok := g.topologicalSort()
	if !ok {
		return nil, fmt.Errorf("graph %v topological sort -> %w", g.name, ErrGraphIsCyclic)
	}
	vGraph := gographviz.NewGraph()
	vGraph.Directed = true
	vGraph.Name = g.name

	for _, node := range g.nodes {
		switch p := node.ptr.(type) {
		case *Static:
			if err := vGraph.AddNode(g.name, node.name, nil); err != nil {
				return nil, fmt.Errorf("add node %v -> %w", node.name, err)
			}
		case *Subflow:
			subG, err := visualizeG(p.g)
			if err != nil {
				return nil, fmt.Errorf("graph %v visualize -> %w", g.name, ErrGraphIsCyclic)
			}
			if err := vGraph.AddSubGraph(g.name, subG.Name, nil); err != nil {
				return nil, fmt.Errorf("add SubGraph %v -> %w", node.name, err)
			}

		}

		if err := vGraph.AddNode(g.name, node.name, nil); err != nil {
			return nil, fmt.Errorf("add node %v -> %w", node.name, err)
		}
	}

	for _, node := range nodes {
		for _, deps := range node.dependents {
			if err := vGraph.AddEdge(deps.name, node.name, true, nil); err != nil {
				return nil, fmt.Errorf("add edge %v - %v -> %w", deps.name, node.name, err)
			}
		}
	}
	return vGraph, nil
}

// // TODO: some other suger to set graph dependency, current not importent
func (tf *TaskFlow) Visualize(writer io.Writer) error {
	vGraph, err := visualizeG(tf.graph)
	if err != nil {
		return fmt.Errorf("graph %v topological sort -> %w", tf.graph.name, ErrGraphIsCyclic)
	}

	if n, err := writer.Write(utils.UnsafeToBytes(vGraph.String())); err != nil {
		return fmt.Errorf("write at %v -> %w", n, err)
	}

	return nil
}

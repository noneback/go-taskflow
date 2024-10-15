package gotaskflow

import (
	"fmt"
	"io"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

type visualizer struct {
	root *cgraph.Graph
}

var Visualizer = visualizer{}

func (v *visualizer) visualizeG(gv *graphviz.Graphviz, g *Graph, parentG *cgraph.Graph) error {
	nodes, err := g.topologicalSort()
	if err != nil {
		return fmt.Errorf("graph %v topological sort -> %w", g.name, err)
	}
	vGraph := parentG
	if vGraph == nil {
		var err error
		vGraph, err = gv.Graph(graphviz.Directed, graphviz.Name(g.name))
		if err != nil {
			return fmt.Errorf("make graph -> %w", err)
		}
		v.root = vGraph
	}
	// defer vGraph.Close()

	nodeMap := make(map[string]*cgraph.Node)

	for _, node := range g.nodes {
		switch p := node.ptr.(type) {
		case *Static:
			vNode, err := vGraph.CreateNode(node.name)
			if err != nil {
				return fmt.Errorf("add node %v -> %w", node.name, err)
			}
			nodeMap[node.name] = vNode
		case *Subflow:
			vSubGraph := vGraph.SubGraph("cluster_"+node.name, 1)
			err := v.visualizeG(gv, p.g, vSubGraph)
			if err != nil {
				fmt.Printf("graph %v visualize -> %s\n", g.name, err)
				// return fmt.Errorf()
				vNode, err := vSubGraph.CreateNode("unvisualized_" + p.g.name)
				if err != nil {
					return fmt.Errorf("add node %v -> %w", node.name, err)
				}
				nodeMap[node.name] = vNode
			} else {
				nodeMap[node.name] = vSubGraph.FirstNode()
			}
		}
	}

	for _, node := range nodes {
		for _, deps := range node.dependents {
			if _, err := vGraph.CreateEdge("", nodeMap[deps.name], nodeMap[node.name]); err != nil {
				return fmt.Errorf("add edge %v - %v -> %w", deps.name, node.name, err)
			}
		}
	}

	return nil
}

func (v *visualizer) Visualize(tf *TaskFlow, writer io.Writer) error {
	gv := graphviz.New()
	defer gv.Close()

	err := v.visualizeG(gv, tf.graph, nil)
	if err != nil {
		return fmt.Errorf("graph %v topological sort -> %w", tf.graph.name, err)
	}

	if err := gv.Render(v.root, graphviz.XDOT, writer); err != nil {
		return fmt.Errorf("render -> %w", err)
	}

	v.root.Close()
	return nil
}

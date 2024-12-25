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

func (v *visualizer) visualizeG(gv *graphviz.Graphviz, g *eGraph, parentG *cgraph.Graph) error {
	vGraph := parentG
	if vGraph == nil {
		var err error
		vGraph, err = gv.Graph(graphviz.Directed, graphviz.Name(g.name))
		if err != nil {
			return fmt.Errorf("make graph -> %w", err)
		}
		vGraph.SetRankDir(cgraph.LRRank)
		v.root = vGraph
	}

	nodeMap := make(map[string]*cgraph.Node)

	for _, node := range g.nodes {
		switch p := node.ptr.(type) {
		case *Static:
			vNode, err := vGraph.CreateNode(node.name)
			if err != nil {
				return fmt.Errorf("add node %v -> %w", node.name, err)
			}
			nodeMap[node.name] = vNode
		case *Condition:
			vNode, err := vGraph.CreateNode(node.name)
			if err != nil {
				return fmt.Errorf("add node %v -> %w", node.name, err)
			}
			vNode.SetShape(cgraph.DiamondShape)
			vNode.SetColor("green")
			nodeMap[node.name] = vNode
		case *Subflow:
			vSubGraph := vGraph.SubGraph("cluster_"+node.name, 1)
			vSubGraph.SetLabel(node.name)
			vSubGraph.SetStyle(cgraph.DashedGraphStyle)
			vSubGraph.SetBackgroundColor("#F5F5F5")
			vSubGraph.SetRankDir(cgraph.LRRank)
			if p.instantiate() != nil || v.visualizeG(gv, p.g, vSubGraph) != nil {
				vNode, err := vGraph.CreateNode("unvisualized_subflow_" + p.g.name)
				if err != nil {
					return fmt.Errorf("add node %v -> %w", node.name, err)
				}
				vNode.SetColor("red")
				vNode.SetComment("cannot visualize due to instantiate panic or failed")
				nodeMap[node.name] = vNode
			} else {
				dummy, _ := vSubGraph.CreateNode(p.g.name)
				dummy.SetShape(cgraph.PointShape)
				nodeMap[node.name] = dummy
			}
		}
	}

	for _, node := range g.nodes {
		for idx, deps := range node.successors {
			label := ""
			style := cgraph.SolidEdgeStyle
			if _, ok := node.ptr.(*Condition); ok {
				label = fmt.Sprintf("%d", idx)
				style = cgraph.DashedEdgeStyle
			}
			edge, err := vGraph.CreateEdge(label, nodeMap[node.name], nodeMap[deps.name])
			if err != nil {
				return fmt.Errorf("add edge %v - %v -> %w", deps.name, node.name, err)
			}
			edge.SetLabel(label)
			edge.SetStyle(style)
		}
	}

	return nil
}

// visualize generate raw dag text in dot format and write to writer
func visualize(tf *TaskFlow, writer io.Writer) error {
	gv := graphviz.New()
	defer gv.Close()
	v := visualizer{}
	err := v.visualizeG(gv, tf.graph, nil)
	if err != nil {
		return fmt.Errorf("visualize %v -> %w", tf.graph.name, err)
	}

	if err := gv.Render(v.root, graphviz.XDOT, writer); err != nil {
		return fmt.Errorf("render -> %w", err)
	}

	v.root.Close()
	return nil
}

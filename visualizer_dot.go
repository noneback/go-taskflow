//go:build !gotaskflow_novis

package gotaskflow

import (
	"fmt"
	"io"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

func init() {
	vizer = &dotVizer{}
}

type dotVizer struct {
	root *cgraph.Graph
}

func (v *dotVizer) visualizeG(gv *graphviz.Graphviz, g *eGraph, parentG *cgraph.Graph) error {
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
		color := "black"
		if node.priority == HIGH {
			color = "#f5427b"
		} else if node.priority == LOW {
			color = "purple"
		}

		switch p := node.ptr.(type) {
		case *Static:
			vNode, err := vGraph.CreateNode(node.name)
			if err != nil {
				return fmt.Errorf("add node %v -> %w", node.name, err)
			}
			vNode.SetColor(color)
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
			vSubGraph.SetRankDir(cgraph.LRRank)
			vSubGraph.SetBackgroundColor("#F5F5F5")
			vSubGraph.SetFontColor(color)
			if p.instantiate() != nil || v.visualizeG(gv, p.g, vSubGraph) != nil {
				vNode, err := vGraph.CreateNode("unvisualized_subflow_" + p.g.name)
				if err != nil {
					return fmt.Errorf("add node %v -> %w", node.name, err)
				}
				vNode.SetColor("#a10212")
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
func (v *dotVizer) Visualize(tf *TaskFlow, writer io.Writer) error {
	gv := graphviz.New()
	defer gv.Close()
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

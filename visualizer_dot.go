package gotaskflow

import (
	"fmt"
	"io"
)

func init() {
	vizer = &dotVizer{}
}

type dotVizer struct{}

// visualizeG recursively visualizes the graph and its subgraphs in DOT format
func (v *dotVizer) visualizeG(g *eGraph, parentGraph *DotGraph) error {
	graph := parentGraph
	if graph == nil {
		graph = NewDotGraph(g.name)
		graph.attributes["rankdir"] = "LR"
	}
	
	nodeMap := make(map[string]*DotNode)
	
	for _, node := range g.nodes {
		color := "black"
		if node.priority == HIGH {
			color = "#f5427b"
		} else if node.priority == LOW {
			color = "purple"
		}
		
		switch p := node.ptr.(type) {
		case *Static:
			dotNode := graph.CreateNode(node.name)
			dotNode.attributes["color"] = color
			nodeMap[node.name] = dotNode
			
		case *Condition:
			dotNode := graph.CreateNode(node.name)
			dotNode.attributes["shape"] = "diamond"
			dotNode.attributes["color"] = "green"
			nodeMap[node.name] = dotNode
			
		case *Subflow:
			subgraph := graph.SubGraph(node.name)
			subgraph.attributes["label"] = node.name
			subgraph.attributes["style"] = "dashed"
			subgraph.attributes["rankdir"] = "LR"
			subgraph.attributes["bgcolor"] = "#F5F5F5"
			
			err := v.visualizeG(p.g, subgraph)
			if err != nil {
				errorNodeName := "unvisualized_subflow_" + p.g.name
				dotNode := graph.CreateNode(errorNodeName)
				dotNode.attributes["color"] = "#a10212"
				dotNode.attributes["comment"] = "cannot visualize due to instantiate panic or failed"
				nodeMap[node.name] = dotNode
			} else {
				dummyNode := graph.CreateNode(p.g.name)
				dummyNode.attributes["shape"] = "point"
				nodeMap[node.name] = dummyNode
			}
		}
	}
	
	for _, node := range g.nodes {
		for idx, deps := range node.successors {
			label := ""
			style := "solid"
			if _, ok := node.ptr.(*Condition); ok {
				label = fmt.Sprintf("%d", idx)
				style = "dashed"
			}
			
			edge := graph.CreateEdge(nodeMap[node.name], nodeMap[deps.name], label)
			if style != "solid" {
				edge.attributes["style"] = style
			}
		}
	}
	
	return nil
}

// visualize generate raw dag text in dot format and write to writer
func (v *dotVizer) Visualize(tf *TaskFlow, writer io.Writer) error {
	graph := NewDotGraph(tf.graph.name)
	err := v.visualizeG(tf.graph, graph)
	if err != nil {
		return fmt.Errorf("visualize %v -> %w", tf.graph.name, err)
	}
	
	_, err = writer.Write([]byte(graph.String()))
	if err != nil {
		return fmt.Errorf("write dot output -> %w", err)
	}
	
	return nil
}

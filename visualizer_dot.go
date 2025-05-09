package gotaskflow

import (
	"fmt"
	"io"
	"strings"
)

func init() {
	vizer = &dotVizer{}
}

type dotVizer struct {
	sb strings.Builder
}

// visualizeG recursively visualizes the graph and its subgraphs in DOT format
func (v *dotVizer) visualizeG(g *eGraph, parent *strings.Builder, isSubgraph bool) error {
	sb := parent
	if parent == nil {
		sb = &v.sb
		sb.WriteString(fmt.Sprintf("digraph %q {\n", g.name))
		sb.WriteString("  rankdir=LR;\n")
	} else if isSubgraph {
		sb.WriteString(fmt.Sprintf("  subgraph %q {\n", "cluster_"+g.name))
		sb.WriteString(fmt.Sprintf("    label=%q;\n", g.name))
		sb.WriteString("    style=dashed;\n")
		sb.WriteString("    rankdir=LR;\n")
		sb.WriteString("    bgcolor=\"#F5F5F5\";\n")
	}

	nodeMap := make(map[string]string)

	for _, node := range g.nodes {
		color := "black"
		if node.priority == HIGH {
			color = "#f5427b"
		} else if node.priority == LOW {
			color = "purple"
		}

		switch p := node.ptr.(type) {
		case *Static:
			nodeName := node.name
			nodeMap[node.name] = nodeName
			sb.WriteString(fmt.Sprintf("    %q [color=%q];\n", nodeName, color))
			
		case *Condition:
			nodeName := node.name
			nodeMap[node.name] = nodeName
			sb.WriteString(fmt.Sprintf("    %q [shape=diamond, color=%q];\n", nodeName, "green"))
			
		case *Subflow:
			subSb := &strings.Builder{}
			err := v.visualizeG(p.g, subSb, true)
			
			if err != nil {
				errorNodeName := "unvisualized_subflow_" + p.g.name
				nodeMap[node.name] = errorNodeName
				sb.WriteString(fmt.Sprintf("    %q [color=%q, comment=%q];\n", 
					errorNodeName, "#a10212", "cannot visualize due to instantiate panic or failed"))
			} else {
				sb.WriteString(subSb.String())
				
				dummyName := p.g.name
				sb.WriteString(fmt.Sprintf("    %q [shape=point];\n", dummyName))
				nodeMap[node.name] = dummyName
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
			
			sb.WriteString(fmt.Sprintf("    %q -> %q", 
				nodeMap[node.name], nodeMap[deps.name]))
				
			if label != "" || style != "solid" {
				attrs := []string{}
				if label != "" {
					attrs = append(attrs, fmt.Sprintf("label=%q", label))
				}
				if style != "solid" {
					attrs = append(attrs, fmt.Sprintf("style=%q", style))
				}
				sb.WriteString(" [" + strings.Join(attrs, ", ") + "]")
			}
			
			sb.WriteString(";\n")
		}
	}

	if parent == nil {
		sb.WriteString("}\n")
	} else if isSubgraph {
		sb.WriteString("  }\n")
	}

	return nil
}

// visualize generate raw dag text in dot format and write to writer
func (v *dotVizer) Visualize(tf *TaskFlow, writer io.Writer) error {
	v.sb.Reset()
	err := v.visualizeG(tf.graph, nil, false)
	if err != nil {
		return fmt.Errorf("visualize %v -> %w", tf.graph.name, err)
	}

	_, err = writer.Write([]byte(v.sb.String()))
	if err != nil {
		return fmt.Errorf("write dot output -> %w", err)
	}

	return nil
}

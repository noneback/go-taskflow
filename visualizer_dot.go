package gotaskflow

import (
	"fmt"
	"io"
	"strings"
)

type dotVizer struct{}

// dotGraph represents a graph in DOT format
type dotGraph struct {
	name       string
	isSubgraph bool
	nodes      map[string]*dotNode
	edges      []*dotEdge
	attributes map[string]string
	subgraphs  []*dotGraph
	indent     string
}

// dotNode represents a node in DOT format
type dotNode struct {
	id         string
	attributes map[string]string
}

// dotEdge represents an edge in DOT format
type dotEdge struct {
	from       *dotNode
	to         *dotNode
	attributes map[string]string
}

func newDotGraph(name string) *dotGraph {
	return &dotGraph{
		name:       name,
		isSubgraph: false,
		nodes:      make(map[string]*dotNode),
		edges:      make([]*dotEdge, 0),
		attributes: make(map[string]string),
		subgraphs:  make([]*dotGraph, 0),
		indent:     "",
	}
}

func (g *dotGraph) CreateNode(name string) *dotNode {
	if node, exists := g.nodes[name]; exists {
		return node
	}

	node := &dotNode{
		id:         name,
		attributes: make(map[string]string),
	}
	g.nodes[name] = node
	return node
}

func (g *dotGraph) CreateEdge(from, to *dotNode, label string) *dotEdge {
	edge := &dotEdge{
		from:       from,
		to:         to,
		attributes: make(map[string]string),
	}
	if label != "" {
		edge.attributes["label"] = label
	}
	g.edges = append(g.edges, edge)
	return edge
}

func (g *dotGraph) SubGraph(name string) *dotGraph {
	subgraph := &dotGraph{
		name:       name,
		isSubgraph: true,
		nodes:      make(map[string]*dotNode),
		edges:      make([]*dotEdge, 0),
		attributes: make(map[string]string),
		subgraphs:  make([]*dotGraph, 0),
		indent:     g.indent + "  ",
	}
	g.subgraphs = append(g.subgraphs, subgraph)
	return subgraph
}

func (g *dotGraph) String() string {
	var sb strings.Builder

	if g.isSubgraph {
		sb.WriteString(g.indent + "subgraph " + quote("cluster_"+g.name) + " {\n")
	} else {
		sb.WriteString(g.indent + "digraph " + quote(g.name) + " {\n")
	}

	for k, v := range g.attributes {
		sb.WriteString(g.indent + "  " + k + "=" + quote(v) + ";\n")
	}

	for _, node := range g.nodes {
		sb.WriteString(node.Format(g.indent + "  "))
	}

	for _, edge := range g.edges {
		sb.WriteString(edge.Format(g.indent + "  "))
	}

	for _, subgraph := range g.subgraphs {
		sb.WriteString(subgraph.String())
	}

	sb.WriteString(g.indent + "}\n")
	return sb.String()
}

func (node *dotNode) Format(indent string) string {
	attrs := formatAttributes(node.attributes)

	if attrs == "" {
		return indent + quote(node.id) + ";\n"
	}

	return indent + quote(node.id) + " [" + attrs + "];\n"
}

func (edge *dotEdge) Format(indent string) string {
	from := edge.from.id
	to := edge.to.id

	attrs := formatAttributes(edge.attributes)

	if attrs == "" {
		return indent + quote(from) + " -> " + quote(to) + ";\n"
	}

	return indent + quote(from) + " -> " + quote(to) + " [" + attrs + "];\n"
}

func quote(s string) string {
	return "\"" + s + "\""
}

func formatAttributes(attrs map[string]string) string {
	if len(attrs) == 0 {
		return ""
	}

	result := make([]string, 0, len(attrs))
	for k, v := range attrs {
		result = append(result, k+"="+quote(v))
	}
	return strings.Join(result, ", ")
}

// visualizeG recursively visualizes the graph and its subgraphs in DOT format
func (v *dotVizer) visualizeG(g *eGraph, parentGraph *dotGraph) error {
	graph := parentGraph
	graph.attributes["rankdir"] = "LR"

	nodeMap := make(map[string]*dotNode)

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
			subgraph.attributes["fontcolor"] = color

			subgraphDot := subgraph.CreateNode(node.name)
			subgraphDot.attributes["shape"] = "point"
			subgraphDot.attributes["height"] = "0.05"
			subgraphDot.attributes["width"] = "0.05"

			nodeMap[node.name] = subgraphDot

			err := v.visualizeG(p.g, subgraph)
			if err != nil {
				errorNodeName := "unvisualized_subflow_" + p.g.name
				dotNode := graph.CreateNode(errorNodeName)
				dotNode.attributes["color"] = "#a10212"
				dotNode.attributes["comment"] = "cannot visualize due to instantiate panic or failed"
				nodeMap[node.name] = dotNode
			}
		}
	}

	for _, node := range g.nodes {
		for idx, deps := range node.successors {
			if from, ok := nodeMap[node.name]; ok {
				if to, ok := nodeMap[deps.name]; ok {
					label := ""
					style := "solid"
					if _, ok := node.ptr.(*Condition); ok {
						label = fmt.Sprintf("%d", idx)
						style = "dashed"
					}

					edge := graph.CreateEdge(from, to, label)
					if style != "solid" {
						edge.attributes["style"] = style
					}
				}
			}
		}
	}

	return nil
}

// Visualize generates raw dag text in dot format and writes to writer
func (v *dotVizer) Visualize(tf *TaskFlow, writer io.Writer) error {
	graph := newDotGraph(tf.graph.name)
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

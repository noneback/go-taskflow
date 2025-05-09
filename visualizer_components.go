package gotaskflow

import (
	"strings"
)

type DotGraph interface {
	CreateNode(name string) DotNode
	CreateEdge(from, to DotNode, label string) DotEdge
	SubGraph(name string) DotGraph
	SetAttribute(key, value string)
	String() string
}

type DotNode interface {
	SetShape(shape string)
	SetColor(color string)
	SetAttribute(key, value string)
	ID() string
}

type DotEdge interface {
	SetStyle(style string)
	SetLabel(label string)
	SetAttribute(key, value string)
}

type dotGraphImpl struct {
	name       string
	isSubgraph bool
	nodes      map[string]DotNode
	edges      []DotEdge
	attributes map[string]string
	subgraphs  []DotGraph
	indent     string
}

type dotNodeImpl struct {
	id         string
	attributes map[string]string
}

type dotEdgeImpl struct {
	from       DotNode
	to         DotNode
	attributes map[string]string
}

func NewDotGraph(name string) DotGraph {
	return &dotGraphImpl{
		name:       name,
		isSubgraph: false,
		nodes:      make(map[string]DotNode),
		edges:      make([]DotEdge, 0),
		attributes: make(map[string]string),
		subgraphs:  make([]DotGraph, 0),
		indent:     "",
	}
}

// CreateNode creates a new node in the graph
func (g *dotGraphImpl) CreateNode(name string) DotNode {
	if node, exists := g.nodes[name]; exists {
		return node
	}
	
	node := &dotNodeImpl{
		id:         name,
		attributes: make(map[string]string),
	}
	g.nodes[name] = node
	return node
}

// CreateEdge creates a new edge in the graph
func (g *dotGraphImpl) CreateEdge(from, to DotNode, label string) DotEdge {
	edge := &dotEdgeImpl{
		from:       from,
		to:         to,
		attributes: make(map[string]string),
	}
	if label != "" {
		edge.SetLabel(label)
	}
	g.edges = append(g.edges, edge)
	return edge
}

func (g *dotGraphImpl) SubGraph(name string) DotGraph {
	subgraph := &dotGraphImpl{
		name:       name,
		isSubgraph: true,
		nodes:      make(map[string]DotNode),
		edges:      make([]DotEdge, 0),
		attributes: make(map[string]string),
		subgraphs:  make([]DotGraph, 0),
		indent:     g.indent + "  ",
	}
	g.subgraphs = append(g.subgraphs, subgraph)
	return subgraph
}

// SetAttribute sets a graph attribute
func (g *dotGraphImpl) SetAttribute(key, value string) {
	g.attributes[key] = value
}

func (n *dotNodeImpl) SetShape(shape string) {
	n.attributes["shape"] = shape
}

func (n *dotNodeImpl) SetColor(color string) {
	n.attributes["color"] = color
}

// SetAttribute sets a node attribute
func (n *dotNodeImpl) SetAttribute(key, value string) {
	n.attributes[key] = value
}

func (n *dotNodeImpl) ID() string {
	return n.id
}

func (e *dotEdgeImpl) SetStyle(style string) {
	e.attributes["style"] = style
}

func (e *dotEdgeImpl) SetLabel(label string) {
	e.attributes["label"] = label
}

// SetAttribute sets an edge attribute
func (e *dotEdgeImpl) SetAttribute(key, value string) {
	e.attributes[key] = value
}

// String generates the DOT format string for the graph
func (g *dotGraphImpl) String() string {
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
		sb.WriteString(formatNode(node, g.indent+"  "))
	}

	for _, edge := range g.edges {
		sb.WriteString(formatEdge(edge, g.indent+"  "))
	}

	for _, subgraph := range g.subgraphs {
		sb.WriteString(subgraph.String())
	}

	sb.WriteString(g.indent + "}\n")
	return sb.String()
}

func formatNode(node DotNode, indent string) string {
	n := node.(*dotNodeImpl)
	if len(n.attributes) == 0 {
		return indent + quote(n.id) + ";\n"
	}

	attrs := make([]string, 0, len(n.attributes))
	for k, v := range n.attributes {
		attrs = append(attrs, k + "=" + quote(v))
	}
	
	return indent + quote(n.id) + " [" + strings.Join(attrs, ", ") + "];\n"
}

func formatEdge(edge DotEdge, indent string) string {
	e := edge.(*dotEdgeImpl)
	from := e.from.ID()
	to := e.to.ID()
	
	if len(e.attributes) == 0 {
		return indent + quote(from) + " -> " + quote(to) + ";\n"
	}

	attrs := make([]string, 0, len(e.attributes))
	for k, v := range e.attributes {
		attrs = append(attrs, k + "=" + quote(v))
	}
	
	return indent + quote(from) + " -> " + quote(to) + " [" + strings.Join(attrs, ", ") + "];\n"
}

func quote(s string) string {
	return "\"" + s + "\""
}

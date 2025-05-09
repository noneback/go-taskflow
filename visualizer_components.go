package gotaskflow

import (
	"strings"
)

type DotGraph struct {
	name       string
	isSubgraph bool
	nodes      map[string]*DotNode
	edges      []*DotEdge
	attributes map[string]string
	subgraphs  []*DotGraph
	indent     string
}

type DotNode struct {
	id         string
	attributes map[string]string
}

type DotEdge struct {
	from       *DotNode
	to         *DotNode
	attributes map[string]string
}

func NewDotGraph(name string) *DotGraph {
	return &DotGraph{
		name:       name,
		isSubgraph: false,
		nodes:      make(map[string]*DotNode),
		edges:      make([]*DotEdge, 0),
		attributes: make(map[string]string),
		subgraphs:  make([]*DotGraph, 0),
		indent:     "",
	}
}

// CreateNode creates a new node in the graph
func (g *DotGraph) CreateNode(name string) *DotNode {
	if node, exists := g.nodes[name]; exists {
		return node
	}
	
	node := &DotNode{
		id:         name,
		attributes: make(map[string]string),
	}
	g.nodes[name] = node
	return node
}

// CreateEdge creates a new edge in the graph
func (g *DotGraph) CreateEdge(from, to *DotNode, label string) *DotEdge {
	edge := &DotEdge{
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

func (g *DotGraph) SubGraph(name string) *DotGraph {
	subgraph := &DotGraph{
		name:       name,
		isSubgraph: true,
		nodes:      make(map[string]*DotNode),
		edges:      make([]*DotEdge, 0),
		attributes: make(map[string]string),
		subgraphs:  make([]*DotGraph, 0),
		indent:     g.indent + "  ",
	}
	g.subgraphs = append(g.subgraphs, subgraph)
	return subgraph
}

func (n *DotNode) ID() string {
	return n.id
}

// String generates the DOT format string for the graph
func (g *DotGraph) String() string {
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

func formatNode(node *DotNode, indent string) string {
	if len(node.attributes) == 0 {
		return indent + quote(node.id) + ";\n"
	}

	attrs := make([]string, 0, len(node.attributes))
	for k, v := range node.attributes {
		attrs = append(attrs, k + "=" + quote(v))
	}
	
	return indent + quote(node.id) + " [" + strings.Join(attrs, ", ") + "];\n"
}

func formatEdge(edge *DotEdge, indent string) string {
	from := edge.from.id
	to := edge.to.id
	
	if len(edge.attributes) == 0 {
		return indent + quote(from) + " -> " + quote(to) + ";\n"
	}

	attrs := make([]string, 0, len(edge.attributes))
	for k, v := range edge.attributes {
		attrs = append(attrs, k + "=" + quote(v))
	}
	
	return indent + quote(from) + " -> " + quote(to) + " [" + strings.Join(attrs, ", ") + "];\n"
}

func quote(s string) string {
	return "\"" + s + "\""
}

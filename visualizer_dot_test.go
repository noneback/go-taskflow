package gotaskflow

import (
	"bytes"
	"strings"
	"testing"
)

func TestDotGraph_String(t *testing.T) {
	graph := NewDotGraph("test_graph")
	graph.attributes["rankdir"] = "LR"

	nodeA := graph.CreateNode("A")
	nodeA.attributes["color"] = "black"

	nodeB := graph.CreateNode("B")
	nodeB.attributes["shape"] = "diamond"

	edge := graph.CreateEdge(nodeA, nodeB, "edge_label")
	edge.attributes["style"] = "dashed"

	result := graph.String()

	expectedParts := []string{
		`digraph "test_graph" {`,
		`rankdir="LR";`,
		`"A" [color="black"];`,
		`"B" [shape="diamond"];`,
		`"A" -> "B"`,
		`}`,
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Expected DOT output to contain %q, but it didn't.\nActual output:\n%s", part, result)
		}
	}
}

func TestDotGraph_SubGraph(t *testing.T) {
	graph := NewDotGraph("main_graph")

	nodeA := graph.CreateNode("A")
	nodeB := graph.CreateNode("B")
	graph.CreateEdge(nodeA, nodeB, "")

	subgraph := graph.SubGraph("sub_graph")
	subgraph.attributes["style"] = "dashed"

	nodeC := subgraph.CreateNode("C")
	nodeD := subgraph.CreateNode("D")
	subgraph.CreateEdge(nodeC, nodeD, "")

	result := graph.String()

	expectedParts := []string{
		`digraph "main_graph" {`,
		`"A" -> "B";`,
		`subgraph "cluster_sub_graph" {`,
		`style="dashed";`,
		`"C";`,
		`"D";`,
		`"C" -> "D";`,
		`}`,
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Expected DOT output to contain %q, but it didn't.\nActual output:\n%s", part, result)
		}
	}
}

func TestDotVizer_Visualize(t *testing.T) {
	tf := NewTaskFlow("test_flow")

	taskA := tf.NewTask("A", func() {})
	taskB := tf.NewTask("B", func() {})
	taskC := tf.NewTask("C", func() {})

	taskA.Precede(taskB)
	taskC.Precede(taskB)

	var buf bytes.Buffer

	vizer := &dotVizer{}
	err := vizer.Visualize(tf, &buf)

	if err != nil {
		t.Fatalf("Visualize returned an error: %v", err)
	}

	result := buf.String()

	expectedParts := []string{
		`digraph "test_flow" {`,
		`rankdir="LR";`,
		`"A"`,
		`"B"`,
		`"C"`,
		`"A" -> "B"`,
		`"C" -> "B"`,
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Expected DOT output to contain %q, but it didn't.\nActual output:\n%s", part, result)
		}
	}
}

func TestDotVizer_VisualizeComplex(t *testing.T) {
	tf := NewTaskFlow("complex_flow")

	taskA := tf.NewTask("A", func() {})
	taskB := tf.NewTask("B", func() {})

	condTask := tf.NewCondition("cond", func() uint { return 0 })

	subTask := tf.NewSubflow("sub", func(sf *Subflow) {
		subA := sf.NewTask("subA", func() {})
		subB := sf.NewTask("subB", func() {})
		subA.Precede(subB)
	})

	taskA.Precede(taskB)
	taskB.Precede(condTask)
	condTask.Precede(subTask)

	var buf bytes.Buffer

	vizer := &dotVizer{}
	err := vizer.Visualize(tf, &buf)

	if err != nil {
		t.Fatalf("Visualize returned an error: %v", err)
	}

	result := buf.String()

	expectedParts := []string{
		`digraph "complex_flow" {`,
		`rankdir="LR";`,
		`"A"`,
		`"B"`,
		`"cond" `,
		`subgraph "cluster_sub" {`,
		`"sub"`,
		`"A" -> "B"`,
		`"B" -> "cond"`,
		`"cond" -> "sub"`,
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Expected DOT output to contain %q, but it didn't.\nActual output:\n%s", part, result)
		}
	}
}

func TestDotNode_Format(t *testing.T) {
	node := &DotNode{
		id:         "test_node",
		attributes: make(map[string]string),
	}

	result := node.Format("  ")
	expected := `  "test_node";` + "\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}

	node.attributes["color"] = "red"
	result = node.Format("  ")
	expected = `  "test_node" [color="red"];` + "\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestDotEdge_Format(t *testing.T) {
	from := &DotNode{id: "from"}
	to := &DotNode{id: "to"}

	edge := &DotEdge{
		from:       from,
		to:         to,
		attributes: make(map[string]string),
	}

	result := edge.Format("  ")
	expected := `  "from" -> "to";` + "\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}

	edge.attributes["style"] = "dashed"
	result = edge.Format("  ")
	expected = `  "from" -> "to" [style="dashed"];` + "\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestFormatAttributes(t *testing.T) {
	attrs := make(map[string]string)
	result := formatAttributes(attrs)
	if result != "" {
		t.Errorf("Expected empty string for empty attributes, got %q", result)
	}

	attrs["color"] = "red"
	result = formatAttributes(attrs)
	expected := `color="red"`
	if result != expected {
		t.Errorf("Expected %q for single attribute, got %q", expected, result)
	}

	attrs["shape"] = "box"
	result = formatAttributes(attrs)

	option1 := `color="red", shape="box"`
	option2 := `shape="box", color="red"`
	if result != option1 && result != option2 {
		t.Errorf("Expected either %q or %q for multiple attributes, got %q", option1, option2, result)
	}
}

func TestQuote(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"test", `"test"`},
		{"", `""`},
		{`"quoted"`, `""quoted""`},
	}

	for _, tc := range testCases {
		result := quote(tc.input)
		if result != tc.expected {
			t.Errorf("quote(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

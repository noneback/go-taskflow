package gotaskflow

import "testing"

func TestTopologicalSort(t *testing.T) {
	t.Run("TestEmptyGraph", func(t *testing.T) {
		graph := newGraph("empty")
		sorted, ok := graph.topologicalSort()
		if !ok || len(sorted) != 0 {
			t.Errorf("expected true and an empty slice, got %v and %v", ok, sorted)
		}
	})

	t.Run("TestSingleNodeGraph", func(t *testing.T) {
		graph := newGraph("single node")
		nodeA := newNode("A")
		graph.push(nodeA)
		sorted, ok := graph.topologicalSort()
		if !ok || len(sorted) != 1 || sorted[0] != nodeA {
			t.Errorf("expected true and the single node, got %v and %v", ok, sorted)
		}
	})

	t.Run("TestSimpleDAG", func(t *testing.T) {
		graph := newGraph("simple DAG")
		nodeA := newNode("A")
		nodeB := newNode("B")
		nodeC := newNode("C")
		nodeA.precede(nodeB)
		nodeB.precede(nodeC)
		graph.push(nodeA, nodeB, nodeC)
		sorted, ok := graph.topologicalSort()
		if !ok || len(sorted) != 3 || sorted[0] != nodeA || sorted[1] != nodeB || sorted[2] != nodeC {
			t.Errorf("expected true and a correct sorted order, got %v and %v", ok, sorted)
		}
	})

	t.Run("TestComplexDAG", func(t *testing.T) {
		graph := newGraph("complex DAG")
		nodeA := newNode("A")
		nodeB := newNode("B")
		nodeC := newNode("C")
		nodeD := newNode("D")
		nodeE := newNode("E")
		nodeA.precede(nodeB)
		nodeA.precede(nodeC)
		nodeB.precede(nodeD)
		nodeC.precede(nodeD)
		nodeD.precede(nodeE)
		graph.push(nodeA, nodeB, nodeC, nodeD, nodeE)
		sorted, ok := graph.topologicalSort()
		if !ok || len(sorted) != 5 {
			t.Errorf("expected true and a correct sorted order, got %v and %v", ok, sorted)
		}
		// Further check the ordering
		nodeIndex := make(map[*Node]int)
		for i, node := range sorted {
			nodeIndex[node] = i
		}
		if nodeIndex[nodeA] > nodeIndex[nodeB] || nodeIndex[nodeC] > nodeIndex[nodeD] {
			t.Errorf("unexpected sort order for complex DAG")
		}
	})

	t.Run("TestGraphWithCycle", func(t *testing.T) {
		graph := newGraph("graph with cycle")
		nodeA := newNode("A")
		nodeB := newNode("B")
		nodeC := newNode("C")
		nodeA.precede(nodeB)
		nodeB.precede(nodeC)
		nodeC.precede(nodeA) // Creates a cycle
		graph.push(nodeA, nodeB, nodeC)
		_, ok := graph.topologicalSort()
		if ok {
			t.Errorf("expected false due to cycle, got %v", ok)
		}
	})
}

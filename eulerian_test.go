package igraph

import (
	"errors"
	"testing"
)

func TestEulerianPathAndCycle(t *testing.T) {
	pathGraph, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pathGraph.Close() })
	status, err := pathGraph.EulerianStatus()
	if err != nil || !status.HasPath || status.HasCycle {
		t.Errorf("path status = %#v, %v", status, err)
	}
	path, err := pathGraph.EulerianPath()
	if err != nil || !path.Found || len(path.Edges) != 2 || len(path.Vertices) != 3 {
		t.Errorf("Eulerian path = %#v, %v", path, err)
	}
	cycle, err := pathGraph.EulerianCycle()
	if err != nil || cycle.Found || cycle.Vertices == nil || cycle.Edges == nil {
		t.Errorf("missing cycle = %#v, %v", cycle, err)
	}

	cycleGraph, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}, {2, 0}}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cycleGraph.Close() })
	status, err = cycleGraph.EulerianStatus()
	if err != nil || !status.HasPath || !status.HasCycle {
		t.Errorf("cycle status = %#v, %v", status, err)
	}
	cycle, err = cycleGraph.EulerianCycle()
	if err != nil || !cycle.Found || len(cycle.Edges) != 3 || len(cycle.Vertices) != 4 || cycle.Vertices[0] != cycle.Vertices[3] {
		t.Errorf("Eulerian cycle = %#v, %v", cycle, err)
	}
}

func TestEulerianMissingEmptyAndClosed(t *testing.T) {
	missing, err := NewGraphFromEdges(4, []Edge{{0, 1}, {0, 2}, {0, 3}}, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = missing.Close() })
	status, err := missing.EulerianStatus()
	if err != nil || status.HasPath || status.HasCycle {
		t.Errorf("missing status = %#v, %v", status, err)
	}
	path, err := missing.EulerianPath()
	if err != nil || path.Found || path.Vertices == nil || path.Edges == nil {
		t.Errorf("missing path = %#v, %v", path, err)
	}

	empty, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	status, err = empty.EulerianStatus()
	if err != nil || !status.HasPath || !status.HasCycle {
		t.Errorf("empty status = %#v, %v", status, err)
	}
	if err := empty.Close(); err != nil {
		t.Fatal(err)
	}
	var nilGraph *Graph
	for _, graph := range []*Graph{empty, nilGraph} {
		_, statusErr := graph.EulerianStatus()
		_, pathErr := graph.EulerianPath()
		_, cycleErr := graph.EulerianCycle()
		for _, err := range []error{statusErr, pathErr, cycleErr} {
			if !errors.Is(err, ErrClosed) {
				t.Errorf("closed Eulerian error = %v", err)
			}
		}
	}
}

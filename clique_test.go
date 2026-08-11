package igraph

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestCliqueScalarAndMembershipQueries(t *testing.T) {
	tests := []struct {
		name         string
		vertices     int
		edges        []Edge
		directed     bool
		complete     bool
		cliqueNumber int
		independence int
	}{
		{name: "null", complete: true},
		{name: "singleton", vertices: 1, complete: true, cliqueNumber: 1, independence: 1},
		{
			name: "complete", vertices: 4, complete: true, cliqueNumber: 4, independence: 1,
			edges: []Edge{{0, 1}, {0, 2}, {0, 3}, {1, 2}, {1, 3}, {2, 3}},
		},
		{name: "edgeless", vertices: 4, cliqueNumber: 1, independence: 4},
		{
			name: "disconnected", vertices: 5, cliqueNumber: 3, independence: 2,
			edges: []Edge{{0, 1}, {0, 2}, {1, 2}, {3, 4}},
		},
		{
			name: "loops and parallel edges ignored", vertices: 3, cliqueNumber: 2, independence: 2,
			edges: []Edge{{0, 0}, {0, 1}, {0, 1}, {2, 2}},
		},
		{
			name: "directed scalar directions ignored", vertices: 3, directed: true,
			cliqueNumber: 3, independence: 1,
			edges: []Edge{{0, 1}, {1, 2}, {2, 0}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph, err := NewGraphFromEdges(tt.vertices, tt.edges, tt.directed)
			if err != nil {
				t.Fatalf("NewGraphFromEdges failed: %v", err)
			}
			defer graph.Close()
			complete, err := graph.IsComplete()
			if err != nil || complete != tt.complete {
				t.Errorf("IsComplete = %t, %v; want %t", complete, err, tt.complete)
			}
			cliqueNumber, err := graph.CliqueNumber()
			if err != nil || cliqueNumber != tt.cliqueNumber {
				t.Errorf("CliqueNumber = %d, %v; want %d", cliqueNumber, err, tt.cliqueNumber)
			}
			independence, err := graph.IndependenceNumber()
			if err != nil || independence != tt.independence {
				t.Errorf("IndependenceNumber = %d, %v; want %d", independence, err, tt.independence)
			}
		})
	}
}

func TestCliqueCandidateQueries(t *testing.T) {
	graph, err := NewGraphFromEdges(4, []Edge{{0, 1}, {1, 0}, {0, 2}, {1, 2}}, true)
	if err != nil {
		t.Fatalf("NewGraphFromEdges failed: %v", err)
	}
	defer graph.Close()

	ids01, _ := VertexIDs(0, 1)
	ids02, _ := VertexIDs(0, 2)
	ids03, _ := VertexIDs(0, 3)
	for _, selector := range []VertexSelector{NoVertices(), mustCliqueVertexIDs(t, 0)} {
		clique, err := graph.IsClique(selector, true)
		if err != nil || !clique {
			t.Errorf("IsClique(empty/singleton) = %t, %v", clique, err)
		}
		independent, err := graph.IsIndependentVertexSet(selector)
		if err != nil || !independent {
			t.Errorf("IsIndependentVertexSet(empty/singleton) = %t, %v", independent, err)
		}
	}
	if got, err := graph.IsClique(ids01, true); err != nil || !got {
		t.Errorf("reciprocal directed clique = %t, %v", got, err)
	}
	if got, err := graph.IsClique(ids02, true); err != nil || got {
		t.Errorf("one-way directed clique = %t, %v", got, err)
	}
	if got, err := graph.IsClique(ids02, false); err != nil || !got {
		t.Errorf("undirected interpretation clique = %t, %v", got, err)
	}
	if got, err := graph.IsIndependentVertexSet(ids03); err != nil || !got {
		t.Errorf("independent set = %t, %v", got, err)
	}
	if got, err := graph.IsIndependentVertexSet(ids02); err != nil || got {
		t.Errorf("adjacent independent set = %t, %v", got, err)
	}
}

func TestCliqueCandidateValidationAndClosure(t *testing.T) {
	var nilGraph *Graph
	nilQueries := []func() error{
		func() error { _, err := nilGraph.IsComplete(); return err },
		func() error { _, err := nilGraph.IsClique(NoVertices(), false); return err },
		func() error { _, err := nilGraph.IsIndependentVertexSet(NoVertices()); return err },
		func() error { _, err := nilGraph.CliqueNumber(); return err },
		func() error { _, err := nilGraph.IndependenceNumber(); return err },
	}
	for i, query := range nilQueries {
		if err := query(); !errors.Is(err, ErrClosed) {
			t.Errorf("nil query %d error = %v", i, err)
		}
	}

	graph, err := NewGraphFromEdges(2, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	duplicate := mustCliqueVertexIDs(t, 0, 0)
	if _, err := graph.IsClique(duplicate, false); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("duplicate IsClique error = %v", err)
	}
	invalid := mustCliqueVertexIDs(t, 2)
	if _, err := graph.IsIndependentVertexSet(invalid); err == nil {
		t.Error("expected invalid vertex error")
	}
	if _, err := graph.IsClique(VertexSelector{kind: 255}, false); err == nil {
		t.Error("expected invalid selector kind error")
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	queries := []func() error{
		func() error { _, err := graph.IsComplete(); return err },
		func() error { _, err := graph.IsClique(NoVertices(), false); return err },
		func() error { _, err := graph.IsIndependentVertexSet(NoVertices()); return err },
		func() error { _, err := graph.CliqueNumber(); return err },
		func() error { _, err := graph.IndependenceNumber(); return err },
	}
	for i, query := range queries {
		if err := query(); !errors.Is(err, ErrClosed) {
			t.Errorf("closed query %d error = %v", i, err)
		}
	}
}

func TestVertexSetContracts(t *testing.T) {
	min, max := 2, 4
	valid := VertexSetEnumerationOptions{Range: VertexSetRange{Minimum: &min, Maximum: &max}, MaxResults: 3}
	if err := valid.validate(); err != nil {
		t.Fatalf("valid options failed: %v", err)
	}
	tests := []VertexSetEnumerationOptions{
		{},
		{MaxResults: -1},
		{MaxResults: int(^uint(0) >> 1)},
		{Range: VertexSetRange{Minimum: intPointer(0)}, MaxResults: 1},
		{Range: VertexSetRange{Maximum: intPointer(-1)}, MaxResults: 1},
		{Range: VertexSetRange{Minimum: intPointer(3), Maximum: intPointer(2)}, MaxResults: 1},
	}
	for _, options := range tests {
		if err := options.validate(); err == nil {
			t.Errorf("options %#v unexpectedly valid", options)
		}
	}
	empty := VertexSetEnumeration{Sets: make([][]int, 0)}
	if empty.Sets == nil || !reflect.DeepEqual(empty.Sets, [][]int{}) {
		t.Errorf("empty result = %#v", empty)
	}
}

func mustCliqueVertexIDs(t *testing.T, ids ...int) VertexSelector {
	t.Helper()
	selector, err := VertexIDs(ids...)
	if err != nil {
		t.Fatal(err)
	}
	return selector
}

func intPointer(value int) *int { return &value }

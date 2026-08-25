package igraph

import (
	"errors"
	"sync"
	"testing"
)

func TestSeparatorPredicates(t *testing.T) {
	starEdges := []Edge{{From: 0, To: 1}, {From: 0, To: 2}, {From: 0, To: 3}, {From: 0, To: 4}}
	for _, directed := range []bool{false, true} {
		g, err := NewGraphFromEdges(5, starEdges, directed)
		if err != nil {
			t.Fatal(err)
		}
		center, _ := VertexIDs(0)
		centerAndLeaf, _ := VertexIDs(0, 1)
		leaf, _ := VertexIDs(1)
		tests := []struct {
			name      string
			candidate VertexSelector
			separator bool
			minimal   bool
		}{
			{name: "center", candidate: center, separator: true, minimal: true},
			{name: "center and leaf", candidate: centerAndLeaf, separator: true},
			{name: "leaf", candidate: leaf},
			{name: "empty", candidate: NoVertices()},
			{name: "all", candidate: AllVertices()},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				separator, err := g.IsSeparator(tt.candidate)
				if err != nil || separator != tt.separator {
					t.Errorf("IsSeparator() = %t, %v, want %t, nil", separator, err, tt.separator)
				}
				minimal, err := g.IsMinimalSeparator(tt.candidate)
				if err != nil || minimal != tt.minimal {
					t.Errorf("IsMinimalSeparator() = %t, %v, want %t, nil", minimal, err, tt.minimal)
				}
			})
		}
		_ = g.Close()
	}
}

func TestSeparatorPredicatesOnDisconnectedAndDegenerateGraphs(t *testing.T) {
	tests := []struct {
		name        string
		vertexCount int
		edges       []Edge
	}{
		{name: "empty"},
		{name: "singleton", vertexCount: 1},
		{name: "disconnected", vertexCount: 3, edges: []Edge{{From: 0, To: 1}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := NewGraphFromEdges(tt.vertexCount, tt.edges, false)
			if err != nil {
				t.Fatal(err)
			}
			defer g.Close()
			for _, predicate := range []func(VertexSelector) (bool, error){g.IsSeparator, g.IsMinimalSeparator} {
				got, err := predicate(NoVertices())
				if err != nil || got {
					t.Errorf("empty candidate = %t, %v, want false, nil", got, err)
				}
			}
		})
	}
}

func TestSeparatorPredicatesRejectInvalidCandidatesAndClosedGraphs(t *testing.T) {
	g, err := NewGraphFromEdges(3, []Edge{{From: 0, To: 1}, {From: 1, To: 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	duplicate, _ := VertexIDs(1, 1)
	outOfRange, _ := VertexIDs(3)
	for _, candidate := range []VertexSelector{duplicate, outOfRange, {kind: vertexSelectorKind(99)}} {
		if _, err := g.IsSeparator(candidate); err == nil {
			t.Errorf("IsSeparator(%#v) error = nil", candidate)
		}
		if _, err := g.IsMinimalSeparator(candidate); err == nil {
			t.Errorf("IsMinimalSeparator(%#v) error = nil", candidate)
		}
	}
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	for _, graph := range []*Graph{g, nil} {
		if _, err := graph.IsSeparator(NoVertices()); !errors.Is(err, ErrClosed) {
			t.Errorf("IsSeparator() error = %v, want %v", err, ErrClosed)
		}
		if _, err := graph.IsMinimalSeparator(NoVertices()); !errors.Is(err, ErrClosed) {
			t.Errorf("IsMinimalSeparator() error = %v, want %v", err, ErrClosed)
		}
	}
}

func TestSeparatorPredicatesAllowConcurrentReads(t *testing.T) {
	g, err := NewGraphFromEdges(4, []Edge{{From: 0, To: 1}, {From: 0, To: 2}, {From: 0, To: 3}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	center, _ := VertexIDs(0)
	var wait sync.WaitGroup
	for range 8 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			if got, err := g.IsSeparator(center); err != nil || !got {
				t.Errorf("IsSeparator() = %t, %v", got, err)
			}
			if got, err := g.IsMinimalSeparator(center); err != nil || !got {
				t.Errorf("IsMinimalSeparator() = %t, %v", got, err)
			}
		}()
	}
	wait.Wait()
}

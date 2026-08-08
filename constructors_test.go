package igraph

import (
	"errors"
	"testing"
)

func TestNewFull(t *testing.T) {
	tests := []struct {
		name      string
		directed  bool
		loops     bool
		wantEdges int
	}{
		{name: "undirected", wantEdges: 6},
		{name: "undirected loops", loops: true, wantEdges: 10},
		{name: "directed", directed: true, wantEdges: 12},
		{name: "directed loops", directed: true, loops: true, wantEdges: 16},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := NewFull(4, tt.directed, tt.loops)
			g = cleanupConstructedGraph(t, g, err)
			assertGraphShape(t, g, 4, tt.wantEdges, tt.directed)
		})
	}
}

func TestNewRing(t *testing.T) {
	tests := []struct {
		name      string
		directed  bool
		mutual    bool
		wantEdges int
	}{
		{name: "undirected", wantEdges: 5},
		{name: "directed", directed: true, wantEdges: 5},
		{name: "directed mutual", directed: true, mutual: true, wantEdges: 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := NewRing(5, tt.directed, tt.mutual)
			g = cleanupConstructedGraph(t, g, err)
			assertGraphShape(t, g, 5, tt.wantEdges, tt.directed)
			for vertex := 0; vertex < 5; vertex++ {
				neighbors, err := g.Neighbors(vertex, DirectionAll)
				if err != nil || len(neighbors) != 2*(1+btoi(tt.directed && tt.mutual)) {
					t.Errorf("Neighbors(%d) = %v, %v", vertex, neighbors, err)
				}
			}
		})
	}
}

func TestNewPath(t *testing.T) {
	for _, directed := range []bool{false, true} {
		g, err := NewPath(4, directed, false)
		g = cleanupConstructedGraph(t, g, err)
		assertGraphShape(t, g, 4, 3, directed)
		adjacent, err := g.AreAdjacent(0, 3)
		if err != nil || adjacent {
			t.Errorf("AreAdjacent(0, 3) = %t, %v, want false, nil", adjacent, err)
		}
	}
}

func TestNewStar(t *testing.T) {
	tests := []struct {
		mode     StarMode
		directed bool
		from     int
		to       int
	}{
		{mode: StarOut, directed: true, from: 2, to: 0},
		{mode: StarIn, directed: true, from: 0, to: 2},
		{mode: StarUndirected, from: 2, to: 0},
		{mode: StarMutual, directed: true, from: 2, to: 0},
	}
	for _, tt := range tests {
		g, err := NewStar(5, 2, tt.mode)
		g = cleanupConstructedGraph(t, g, err)
		wantEdges := 4
		if tt.mode == StarMutual {
			wantEdges = 8
		}
		assertGraphShape(t, g, 5, wantEdges, tt.directed)
		adjacent, err := g.AreAdjacent(tt.from, tt.to)
		if err != nil || !adjacent {
			t.Errorf("AreAdjacent(%d, %d) = %t, %v, want true, nil", tt.from, tt.to, adjacent, err)
		}
	}
}

func TestNewKaryTree(t *testing.T) {
	tests := []struct {
		mode     TreeMode
		directed bool
		from     int
		to       int
	}{
		{mode: TreeOut, directed: true, from: 0, to: 1},
		{mode: TreeIn, directed: true, from: 1, to: 0},
		{mode: TreeUndirected, from: 0, to: 1},
	}
	for _, tt := range tests {
		g, err := NewKaryTree(7, 2, tt.mode)
		g = cleanupConstructedGraph(t, g, err)
		assertGraphShape(t, g, 7, 6, tt.directed)
		adjacent, err := g.AreAdjacent(tt.from, tt.to)
		if err != nil || !adjacent {
			t.Errorf("AreAdjacent(%d, %d) = %t, %v, want true, nil", tt.from, tt.to, adjacent, err)
		}
	}
}

func TestNewHypercube(t *testing.T) {
	for _, directed := range []bool{false, true} {
		g, err := NewHypercube(3, directed)
		g = cleanupConstructedGraph(t, g, err)
		assertGraphShape(t, g, 8, 12, directed)
		for vertex := 0; vertex < 8; vertex++ {
			neighbors, err := g.Neighbors(vertex, DirectionAll)
			if err != nil || len(neighbors) != 3 {
				t.Errorf("Neighbors(%d) = %v, %v, want 3 neighbors", vertex, neighbors, err)
			}
		}
	}
}

func TestDeterministicConstructorsRejectInvalidInput(t *testing.T) {
	tests := []struct {
		name      string
		construct func() (*Graph, error)
	}{
		{name: "negative full size", construct: func() (*Graph, error) { return NewFull(-1, false, false) }},
		{name: "negative ring size", construct: func() (*Graph, error) { return NewRing(-1, false, false) }},
		{name: "negative path size", construct: func() (*Graph, error) { return NewPath(-1, false, false) }},
		{name: "empty star", construct: func() (*Graph, error) { return NewStar(0, 0, StarUndirected) }},
		{name: "invalid star center", construct: func() (*Graph, error) { return NewStar(3, 3, StarUndirected) }},
		{name: "invalid star mode", construct: func() (*Graph, error) { return NewStar(3, 0, StarMode(99)) }},
		{name: "negative tree size", construct: func() (*Graph, error) { return NewKaryTree(-1, 2, TreeOut) }},
		{name: "invalid child count", construct: func() (*Graph, error) { return NewKaryTree(3, 0, TreeOut) }},
		{name: "invalid tree mode", construct: func() (*Graph, error) { return NewKaryTree(3, 2, TreeMode(99)) }},
		{name: "negative hypercube dimension", construct: func() (*Graph, error) { return NewHypercube(-1, false) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if graph, err := tt.construct(); err == nil || graph != nil {
				t.Errorf("constructor = %v, %v, want nil, error", graph, err)
			}
		})
	}
}

func TestFinishGraphConstructionPropagatesErrors(t *testing.T) {
	graph, err := finishGraphConstruction(&Graph{}, "test construction", 1)
	if graph != nil || err == nil {
		t.Fatalf("finishGraphConstruction() = %v, %v, want nil, error", graph, err)
	}
}

func TestConstructedGraphOwnership(t *testing.T) {
	g, err := NewRing(4, false, false)
	g = cleanupConstructedGraph(t, g, err)
	if err := g.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if _, err := g.VertexCount(); !errors.Is(err, ErrClosed) {
		t.Errorf("VertexCount() error = %v, want %v", err, ErrClosed)
	}
}

func cleanupConstructedGraph(t *testing.T, graph *Graph, err error) *Graph {
	t.Helper()
	if err != nil {
		t.Fatalf("constructor error = %v", err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	return graph
}

func assertGraphShape(t *testing.T, graph *Graph, wantVertices, wantEdges int, wantDirected bool) {
	t.Helper()
	if got, err := graph.VertexCount(); err != nil || got != wantVertices {
		t.Errorf("VertexCount() = %d, %v, want %d, nil", got, err, wantVertices)
	}
	if got, err := graph.EdgeCount(); err != nil || got != wantEdges {
		t.Errorf("EdgeCount() = %d, %v, want %d, nil", got, err, wantEdges)
	}
	if got, err := graph.IsDirected(); err != nil || got != wantDirected {
		t.Errorf("IsDirected() = %t, %v, want %t, nil", got, err, wantDirected)
	}
}

func btoi(value bool) int {
	if value {
		return 1
	}
	return 0
}

package igraph_test

import (
	"testing"

	"github.com/h8gi/go-igraph"
)

func TestDyadCensus(t *testing.T) {
	tests := []struct {
		name string
		fn   func() (*igraph.Graph, error)
		want igraph.DyadCensusResult
	}{
		{
			name: "empty graph",
			fn: func() (*igraph.Graph, error) {
				return igraph.NewGraphFromEdges(0, nil, true)
			},
			want: igraph.DyadCensusResult{Mutual: 0, Asymmetric: 0, Null: 0},
		},
		{
			name: "two vertices directed no edge",
			fn: func() (*igraph.Graph, error) {
				return igraph.NewGraphFromEdges(2, nil, true)
			},
			want: igraph.DyadCensusResult{Mutual: 0, Asymmetric: 0, Null: 1},
		},
		{
			name: "two vertices directed one edge",
			fn: func() (*igraph.Graph, error) {
				return igraph.NewGraphFromEdges(2, []igraph.Edge{{0, 1}}, true)
			},
			want: igraph.DyadCensusResult{Mutual: 0, Asymmetric: 1, Null: 0},
		},
		{
			name: "two vertices directed both edges",
			fn: func() (*igraph.Graph, error) {
				return igraph.NewGraphFromEdges(2, []igraph.Edge{{0, 1}, {1, 0}}, true)
			},
			want: igraph.DyadCensusResult{Mutual: 1, Asymmetric: 0, Null: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := tt.fn()
			if err != nil {
				t.Fatalf("failed to create graph: %v", err)
			}
			defer g.Close()

			got, err := g.DyadCensus()
			if err != nil {
				t.Fatalf("DyadCensus() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("DyadCensus() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestTriadCensus(t *testing.T) {
	g, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{0, 1}, {1, 2}, {2, 0}}, false)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}
	defer g.Close()

	result, err := g.TriadCensus()
	if err != nil {
		t.Fatalf("TriadCensus() error = %v", err)
	}

	if len(result) != 16 {
		t.Errorf("result length = %d, want 16", len(result))
	}
}

func TestTrianglesCount(t *testing.T) {
	tests := []struct {
		name string
		fn   func() (*igraph.Graph, error)
		want int64
	}{
		{
			name: "empty graph",
			fn: func() (*igraph.Graph, error) {
				return igraph.NewGraphFromEdges(0, nil, false)
			},
			want: 0,
		},
		{
			name: "single triangle",
			fn: func() (*igraph.Graph, error) {
				return igraph.NewGraphFromEdges(3, []igraph.Edge{{0, 1}, {1, 2}, {2, 0}}, false)
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := tt.fn()
			if err != nil {
				t.Fatalf("failed to create graph: %v", err)
			}
			defer g.Close()

			got, err := g.TrianglesCount()
			if err != nil {
				t.Fatalf("TrianglesCount() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("TrianglesCount() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestTrianglesList(t *testing.T) {
	g, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{0, 1}, {1, 2}, {2, 0}}, false)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}
	defer g.Close()

	triangles, err := g.TrianglesList()
	if err != nil {
		t.Fatalf("TrianglesList() error = %v", err)
	}

	if len(triangles) != 1 {
		t.Errorf("len(triangles) = %d, want 1", len(triangles))
	} else if triangles[0] != [3]int{0, 1, 2} {
		t.Errorf("triangle = %v, want [0,1,2]", triangles[0])
	}
}

func TestDyadCensusErrors(t *testing.T) {
	g, err := igraph.NewGraphFromEdges(2, nil, true)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}
	g.Close()

	_, err = g.DyadCensus()
	if err != igraph.ErrClosed {
		t.Errorf("DyadCensus() on closed graph: expected ErrClosed, got %v", err)
	}
}

func TestTriadCensusErrors(t *testing.T) {
	g, err := igraph.NewGraphFromEdges(2, nil, true)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}
	g.Close()

	_, err = g.TriadCensus()
	if err != igraph.ErrClosed {
		t.Errorf("TriadCensus() on closed graph: expected ErrClosed, got %v", err)
	}
}

func TestTrianglesCountErrors(t *testing.T) {
	g, err := igraph.NewGraphFromEdges(2, nil, true)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}
	g.Close()

	_, err = g.TrianglesCount()
	if err != igraph.ErrClosed {
		t.Errorf("TrianglesCount() on closed graph: expected ErrClosed, got %v", err)
	}
}

func TestTrianglesListErrors(t *testing.T) {
	g, err := igraph.NewGraphFromEdges(2, nil, true)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}
	g.Close()

	_, err = g.TrianglesList()
	if err != igraph.ErrClosed {
		t.Errorf("TrianglesList() on closed graph: expected ErrClosed, got %v", err)
	}
}

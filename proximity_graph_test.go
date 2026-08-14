package igraph_test

import (
	"errors"
	"math"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestProximityGraphsPreservePointIdentity(t *testing.T) {
	points, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}, {2, 0}, {0, 2}})
	constructors := map[string]struct {
		construct func(igraph.Matrix) (*igraph.Graph, error)
		edges     int
	}{
		"Delaunay":              {igraph.NewDelaunayGraph, 3},
		"Gabriel":               {igraph.NewGabrielGraph, 2},
		"relative neighborhood": {igraph.NewRelativeNeighborhoodGraph, 2},
	}
	for name, test := range constructors {
		t.Run(name, func(t *testing.T) {
			graph, err := test.construct(points)
			if err != nil {
				t.Fatal(err)
			}
			defer graph.Close()
			if vertices, err := graph.VertexCount(); err != nil || vertices != 3 {
				t.Fatalf("VertexCount = %d, %v, want 3, nil", vertices, err)
			}
			if edges, err := graph.EdgeCount(); err != nil || edges != test.edges {
				t.Fatalf("EdgeCount = %d, %v, want %d, nil", edges, err, test.edges)
			}
			if directed, err := graph.IsDirected(); err != nil || directed {
				t.Fatalf("IsDirected = %t, %v, want false, nil", directed, err)
			}
		})
	}
}

func TestOneDimensionalProximityGraphs(t *testing.T) {
	points, _ := igraph.NewMatrixFromRows([][]float64{{5}, {0}, {9}, {2}})
	constructors := []func(igraph.Matrix) (*igraph.Graph, error){
		igraph.NewDelaunayGraph,
		igraph.NewGabrielGraph,
		igraph.NewRelativeNeighborhoodGraph,
	}
	for _, construct := range constructors {
		graph, err := construct(points)
		if err != nil {
			t.Fatal(err)
		}
		defer graph.Close()
		if edges, err := graph.EdgeCount(); err != nil || edges != 3 {
			t.Fatalf("EdgeCount = %d, %v, want 3, nil", edges, err)
		}
		for _, edge := range [][2]int{{1, 3}, {3, 0}, {0, 2}} {
			adjacent, err := graph.AreAdjacent(edge[0], edge[1])
			if err != nil || !adjacent {
				t.Fatalf("AreAdjacent(%d, %d) = %t, %v", edge[0], edge[1], adjacent, err)
			}
		}
	}
}

func TestProximityGraphInclusion(t *testing.T) {
	points, _ := igraph.NewMatrixFromRows([][]float64{
		{0, 0}, {3, 0}, {4, 2}, {2, 4}, {0, 3}, {1.4, 1.1},
	})
	delaunay, err := igraph.NewDelaunayGraph(points)
	if err != nil {
		t.Fatal(err)
	}
	defer delaunay.Close()
	gabriel, err := igraph.NewGabrielGraph(points)
	if err != nil {
		t.Fatal(err)
	}
	defer gabriel.Close()
	relative, err := igraph.NewRelativeNeighborhoodGraph(points)
	if err != nil {
		t.Fatal(err)
	}
	defer relative.Close()

	assertEdgeSubset(t, relative, gabriel)
	assertEdgeSubset(t, gabriel, delaunay)
}

func TestProximityGraphsEmptySingleAndOwnership(t *testing.T) {
	constructors := []func(igraph.Matrix) (*igraph.Graph, error){
		igraph.NewDelaunayGraph,
		igraph.NewGabrielGraph,
		igraph.NewRelativeNeighborhoodGraph,
	}
	for _, construct := range constructors {
		empty, err := construct(igraph.Matrix{})
		if err != nil {
			t.Fatal(err)
		}
		if vertices, _ := empty.VertexCount(); vertices != 0 {
			t.Fatalf("empty vertex count = %d", vertices)
		}
		if err := empty.Close(); err != nil {
			t.Fatal(err)
		}
		if err := empty.Close(); err != nil {
			t.Fatalf("repeated Close error = %v", err)
		}
		if _, err := empty.VertexCount(); !errors.Is(err, igraph.ErrClosed) {
			t.Fatalf("VertexCount after Close error = %v, want ErrClosed", err)
		}

		singlePoint, _ := igraph.NewMatrixFromRows([][]float64{{4, 5}})
		single, err := construct(singlePoint)
		if err != nil {
			t.Fatal(err)
		}
		defer single.Close()
		if vertices, _ := single.VertexCount(); vertices != 1 {
			t.Fatalf("single vertex count = %d", vertices)
		}
		if edges, _ := single.EdgeCount(); edges != 0 {
			t.Fatalf("single edge count = %d", edges)
		}
	}
}

func TestProximityGraphsRejectInvalidInputs(t *testing.T) {
	duplicate, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}, {0, 0}})
	zeroDimensional, _ := igraph.NewMatrix(2, 0)
	nonFinite, _ := igraph.NewMatrixFromRows([][]float64{{0}, {math.NaN()}})
	constructors := []func(igraph.Matrix) (*igraph.Graph, error){
		igraph.NewDelaunayGraph,
		igraph.NewGabrielGraph,
		igraph.NewRelativeNeighborhoodGraph,
	}
	for _, construct := range constructors {
		for _, points := range []igraph.Matrix{duplicate, zeroDimensional, nonFinite} {
			if graph, err := construct(points); err == nil {
				_ = graph.Close()
				t.Fatal("expected validation error")
			}
		}
	}

	collinear, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}, {1, 1}, {2, 2}})
	if graph, err := igraph.NewDelaunayGraph(collinear); err == nil {
		_ = graph.Close()
		t.Fatal("expected degenerate Delaunay error")
	}
}

func assertEdgeSubset(t *testing.T, subset, superset *igraph.Graph) {
	t.Helper()
	edges, err := subset.Edges()
	if err != nil {
		t.Fatal(err)
	}
	for _, edge := range edges {
		adjacent, err := superset.AreAdjacent(edge.From, edge.To)
		if err != nil || !adjacent {
			t.Fatalf("edge (%d, %d) missing from superset: %v", edge.From, edge.To, err)
		}
	}
}

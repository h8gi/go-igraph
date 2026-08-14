package igraph_test

import (
	"math"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestNearestNeighborDirectedAndUndirected(t *testing.T) {
	points, _ := igraph.NewMatrixFromRows([][]float64{{0}, {3}, {10}})
	maximum := 1
	directed, err := igraph.NewNearestNeighborGraph(points, igraph.NearestNeighborOptions{MaxNeighbors: &maximum, Directed: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = directed.Close() })
	if value, _ := directed.IsDirected(); !value {
		t.Fatal("directed nearest-neighbor graph is undirected")
	}
	assertOutgoingNeighbors(t, directed, 0, []int{1})
	assertOutgoingNeighbors(t, directed, 1, []int{0})
	assertOutgoingNeighbors(t, directed, 2, []int{1})

	undirected, err := igraph.NewNearestNeighborGraph(points, igraph.NearestNeighborOptions{MaxNeighbors: &maximum})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = undirected.Close() })
	if value, _ := undirected.IsDirected(); value {
		t.Fatal("undirected nearest-neighbor graph is directed")
	}
	if count, _ := undirected.EdgeCount(); count != 2 {
		t.Fatalf("undirected edge count = %d, want 2", count)
	}
	for _, edge := range [][2]int{{0, 1}, {1, 2}} {
		adjacent, err := undirected.AreAdjacent(edge[0], edge[1])
		if err != nil || !adjacent {
			t.Fatalf("AreAdjacent(%d, %d) = %t, %v", edge[0], edge[1], adjacent, err)
		}
	}
}

func TestNearestNeighborMetricAndCutoff(t *testing.T) {
	points, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}, {3, 0}, {2, 2}})
	maximum := 1
	for name, test := range map[string]struct {
		metric igraph.SpatialMetric
		want   int
	}{
		"Euclidean": {metric: igraph.SpatialEuclidean, want: 2},
		"Manhattan": {metric: igraph.SpatialManhattan, want: 1},
	} {
		t.Run(name, func(t *testing.T) {
			graph, err := igraph.NewNearestNeighborGraph(points, igraph.NearestNeighborOptions{Metric: test.metric, MaxNeighbors: &maximum, Directed: true})
			if err != nil {
				t.Fatal(err)
			}
			defer graph.Close()
			assertOutgoingNeighbors(t, graph, 0, []int{test.want})
		})
	}

	line, _ := igraph.NewMatrixFromRows([][]float64{{0}, {2}, {5}})
	cutoff := 2.1
	graph, err := igraph.NewNearestNeighborGraph(line, igraph.NearestNeighborOptions{Cutoff: &cutoff, Directed: true})
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	if count, _ := graph.EdgeCount(); count != 2 {
		t.Fatalf("cutoff edge count = %d, want 2", count)
	}
	assertOutgoingNeighbors(t, graph, 0, []int{1})
	assertOutgoingNeighbors(t, graph, 1, []int{0})
	assertOutgoingNeighbors(t, graph, 2, []int{})

	exactCutoff := 2.0
	exact, err := igraph.NewNearestNeighborGraph(line, igraph.NearestNeighborOptions{Cutoff: &exactCutoff, Directed: true})
	if err != nil {
		t.Fatal(err)
	}
	defer exact.Close()
	if count, _ := exact.EdgeCount(); count != 0 {
		t.Fatalf("exact-cutoff edge count = %d, want 0", count)
	}
}

func TestNearestNeighborTieAndZeroBounds(t *testing.T) {
	star, _ := igraph.NewMatrixFromRows([][]float64{{0, 0}, {1, 0}, {-1, 0}, {0, 1}, {0, -1}})
	maximum := 1
	graph, err := igraph.NewNearestNeighborGraph(star, igraph.NearestNeighborOptions{MaxNeighbors: &maximum, Directed: true})
	if err != nil {
		t.Fatal(err)
	}
	neighbors, err := graph.Neighbors(0, igraph.DirectionOut)
	_ = graph.Close()
	if err != nil || len(neighbors) != 1 || neighbors[0] < 1 || neighbors[0] > 4 {
		t.Fatalf("tied nearest neighbor = %v, %v", neighbors, err)
	}

	zero := 0
	zeroGraph, err := igraph.NewNearestNeighborGraph(star, igraph.NearestNeighborOptions{MaxNeighbors: &zero})
	if err != nil {
		t.Fatal(err)
	}
	defer zeroGraph.Close()
	if count, _ := zeroGraph.EdgeCount(); count != 0 {
		t.Fatalf("zero-neighbor edge count = %d, want 0", count)
	}

	zeroCutoff := 0.0
	distinct, _ := igraph.NewMatrixFromRows([][]float64{{0}, {1}})
	cutoffGraph, err := igraph.NewNearestNeighborGraph(distinct, igraph.NearestNeighborOptions{Cutoff: &zeroCutoff})
	if err != nil {
		t.Fatal(err)
	}
	defer cutoffGraph.Close()
	if count, _ := cutoffGraph.EdgeCount(); count != 0 {
		t.Fatalf("zero-cutoff edge count = %d, want 0", count)
	}
}

func TestNearestNeighborEmptySingleAndOwnership(t *testing.T) {
	empty, err := igraph.NewNearestNeighborGraph(igraph.Matrix{}, igraph.NearestNeighborOptions{Directed: true})
	if err != nil {
		t.Fatal(err)
	}
	if vertices, _ := empty.VertexCount(); vertices != 0 {
		t.Fatalf("empty vertex count = %d", vertices)
	}
	if directed, _ := empty.IsDirected(); !directed {
		t.Fatal("empty directed result lost directedness")
	}
	if err := empty.Close(); err != nil {
		t.Fatal(err)
	}
	if err := empty.Close(); err != nil {
		t.Fatalf("repeated Close error = %v", err)
	}

	singlePoint, _ := igraph.NewMatrixFromRows([][]float64{{4, 5, 6}})
	single, err := igraph.NewNearestNeighborGraph(singlePoint, igraph.NearestNeighborOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer single.Close()
	if vertices, _ := single.VertexCount(); vertices != 1 {
		t.Fatalf("single-point vertex count = %d", vertices)
	}
	if edges, _ := single.EdgeCount(); edges != 0 {
		t.Fatalf("single-point edge count = %d", edges)
	}
}

func TestNearestNeighborRejectsInvalidInputs(t *testing.T) {
	valid, _ := igraph.NewMatrixFromRows([][]float64{{0}, {1}})
	zeroDimensional, _ := igraph.NewMatrix(2, 0)
	nonFinite, _ := igraph.NewMatrixFromRows([][]float64{{0}, {math.Inf(1)}})
	negativeMaximum := -1
	negativeCutoff := -1.0
	for name, test := range map[string]struct {
		points  igraph.Matrix
		options igraph.NearestNeighborOptions
	}{
		"metric":          {valid, igraph.NearestNeighborOptions{Metric: igraph.SpatialMetric(99)}},
		"neighbor count":  {valid, igraph.NearestNeighborOptions{MaxNeighbors: &negativeMaximum}},
		"cutoff":          {valid, igraph.NearestNeighborOptions{Cutoff: &negativeCutoff}},
		"zero dimensions": {zeroDimensional, igraph.NearestNeighborOptions{}},
		"non-finite":      {nonFinite, igraph.NearestNeighborOptions{}},
	} {
		t.Run(name, func(t *testing.T) {
			if graph, err := igraph.NewNearestNeighborGraph(test.points, test.options); err == nil {
				_ = graph.Close()
				t.Fatal("expected validation error")
			}
		})
	}
}

func assertOutgoingNeighbors(t *testing.T, graph *igraph.Graph, vertex int, want []int) {
	t.Helper()
	got, err := graph.Neighbors(vertex, igraph.DirectionOut)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("vertex %d outgoing neighbors = %v, want %v", vertex, got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("vertex %d outgoing neighbors = %v, want %v", vertex, got, want)
		}
	}
}

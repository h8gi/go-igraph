package igraph_test

import (
	"math"
	"reflect"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestAdjacencyDirectedMultiplicityAndOwnership(t *testing.T) {
	matrix, _ := igraph.NewMatrixFromRows([][]float64{{0, 2, 0}, {0, 0, 1}, {1, 0, 0}})
	graph, err := igraph.NewAdjacency(matrix, igraph.AdjacencyOptions{})
	if err != nil {
		t.Fatal(err)
	}
	edges, err := graph.Edges()
	if err != nil {
		t.Fatal(err)
	}
	want := []igraph.Edge{{From: 2, To: 0}, {From: 0, To: 1}, {From: 0, To: 1}, {From: 1, To: 2}}
	if !reflect.DeepEqual(edges, want) {
		t.Fatalf("edges = %v, want %v", edges, want)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestAdjacencyUndirectedCombinationModes(t *testing.T) {
	matrix, _ := igraph.NewMatrixFromRows([][]float64{{0, 2}, {3, 0}})
	for _, test := range []struct {
		mode igraph.AdjacencyMode
		want int
	}{
		{igraph.AdjacencyUpper, 2},
		{igraph.AdjacencyLower, 3},
		{igraph.AdjacencyMin, 2},
		{igraph.AdjacencyPlus, 5},
		{igraph.AdjacencyMax, 3},
	} {
		graph, err := igraph.NewAdjacency(matrix, igraph.AdjacencyOptions{Mode: test.mode})
		if err != nil {
			t.Fatalf("mode %d: %v", test.mode, err)
		}
		count, err := graph.EdgeCount()
		_ = graph.Close()
		if err != nil || count != test.want {
			t.Errorf("mode %d count = %d, %v; want %d", test.mode, count, err, test.want)
		}
	}

	symmetric, _ := igraph.NewMatrixFromRows([][]float64{{0, 2}, {2, 0}})
	graph, err := igraph.NewAdjacency(symmetric, igraph.AdjacencyOptions{Mode: igraph.AdjacencyUndirected})
	if err != nil {
		t.Fatal(err)
	}
	count, _ := graph.EdgeCount()
	_ = graph.Close()
	if count != 2 {
		t.Fatalf("symmetric count = %d", count)
	}
	if _, err := igraph.NewAdjacency(matrix, igraph.AdjacencyOptions{Mode: igraph.AdjacencyUndirected}); err == nil {
		t.Fatal("asymmetric undirected error nil")
	}
}

func TestAdjacencyLoopModes(t *testing.T) {
	matrix, _ := igraph.NewMatrixFromRows([][]float64{{2}})
	for _, test := range []struct {
		loops igraph.AdjacencyLoops
		want  int
	}{
		{igraph.AdjacencyNoLoops, 0},
		{igraph.AdjacencyLoopsOnce, 2},
		{igraph.AdjacencyLoopsTwice, 1},
	} {
		graph, err := igraph.NewAdjacency(matrix, igraph.AdjacencyOptions{Mode: igraph.AdjacencyUndirected, Loops: test.loops})
		if err != nil {
			t.Fatalf("loops %d: %v", test.loops, err)
		}
		count, _ := graph.EdgeCount()
		_ = graph.Close()
		if count != test.want {
			t.Errorf("loops %d count = %d, want %d", test.loops, count, test.want)
		}
	}
}

func TestWeightedAdjacencyAlignmentAndLoops(t *testing.T) {
	matrix, _ := igraph.NewMatrixFromRows([][]float64{{4, 2}, {-3, 0}})
	result, err := igraph.NewWeightedAdjacency(matrix, igraph.AdjacencyOptions{Loops: igraph.AdjacencyLoopsOnce})
	if err != nil {
		t.Fatal(err)
	}
	defer result.Graph.Close()
	edges, err := result.Graph.Edges()
	if err != nil {
		t.Fatal(err)
	}
	got := make(map[igraph.Edge]float64, len(edges))
	for index, edge := range edges {
		got[edge] = result.Weights[index]
	}
	want := map[igraph.Edge]float64{{From: 0, To: 0}: 4, {From: 0, To: 1}: 2, {From: 1, To: 0}: -3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("weighted edges = %v, want %v", got, want)
	}
	if err := result.Graph.Close(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Weights, []float64{4, -3, 2}) {
		t.Fatalf("weights after close = %v", result.Weights)
	}

	loop, _ := igraph.NewMatrixFromRows([][]float64{{4}})
	twice, err := igraph.NewWeightedAdjacency(loop, igraph.AdjacencyOptions{Mode: igraph.AdjacencyUndirected, Loops: igraph.AdjacencyLoopsTwice})
	if err != nil {
		t.Fatal(err)
	}
	defer twice.Graph.Close()
	if !reflect.DeepEqual(twice.Weights, []float64{2}) {
		t.Fatalf("twice loop weights = %v", twice.Weights)
	}
}

func TestAdjacencyEmptyAndValidation(t *testing.T) {
	empty, _ := igraph.NewMatrix(0, 0)
	graph, err := igraph.NewAdjacency(empty, igraph.AdjacencyOptions{})
	if err != nil {
		t.Fatal(err)
	}
	count, _ := graph.VertexCount()
	_ = graph.Close()
	if count != 0 {
		t.Fatalf("empty vertices = %d", count)
	}
	weighted, err := igraph.NewWeightedAdjacency(empty, igraph.AdjacencyOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer weighted.Graph.Close()
	if weighted.Weights == nil || len(weighted.Weights) != 0 {
		t.Fatalf("empty weights = %#v", weighted.Weights)
	}

	rectangular, _ := igraph.NewMatrix(1, 2)
	if _, err := igraph.NewAdjacency(rectangular, igraph.AdjacencyOptions{}); err == nil {
		t.Fatal("rectangular error nil")
	}
	for _, value := range []float64{-1, 1.5, math.NaN(), math.Inf(1)} {
		matrix, _ := igraph.NewMatrixFromRows([][]float64{{value}})
		if _, err := igraph.NewAdjacency(matrix, igraph.AdjacencyOptions{}); err == nil {
			t.Errorf("unweighted value %v error nil", value)
		}
	}
	for _, value := range []float64{math.NaN(), math.Inf(-1)} {
		matrix, _ := igraph.NewMatrixFromRows([][]float64{{value}})
		if _, err := igraph.NewWeightedAdjacency(matrix, igraph.AdjacencyOptions{}); err == nil {
			t.Errorf("weighted value %v error nil", value)
		}
	}
	valid, _ := igraph.NewMatrixFromRows([][]float64{{0}})
	if _, err := igraph.NewAdjacency(valid, igraph.AdjacencyOptions{Mode: igraph.AdjacencyMode(99)}); err == nil {
		t.Fatal("invalid mode error nil")
	}
	if _, err := igraph.NewAdjacency(valid, igraph.AdjacencyOptions{Loops: igraph.AdjacencyLoops(99)}); err == nil {
		t.Fatal("invalid loops error nil")
	}
}

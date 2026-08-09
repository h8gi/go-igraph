package igraph

import (
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestDistancesPreserveSelectorOrderAndDuplicates(t *testing.T) {
	graph := newPathTestGraph(t)
	sources, err := VertexIDs(2, 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	targets, err := VertexIDs(3, 1, 3, 4, 1)
	if err != nil {
		t.Fatal(err)
	}

	distances, err := graph.Distances(sources, targets, PathOptions{Direction: DirectionOut})
	if err != nil {
		t.Fatal(err)
	}
	assertMatrixRows(t, distances, [][]float64{
		{1, math.Inf(1), 1, math.Inf(1), math.Inf(1)},
		{2, 1, 2, math.Inf(1), 1},
		{1, math.Inf(1), 1, math.Inf(1), math.Inf(1)},
	})

	weighted, err := graph.Distances(sources, targets, PathOptions{
		Direction: DirectionOut,
		Weights:   []float64{2, 3, 10, 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertMatrixRows(t, weighted, [][]float64{
		{1, math.Inf(1), 1, math.Inf(1), math.Inf(1)},
		{6, 2, 6, math.Inf(1), 2},
		{1, math.Inf(1), 1, math.Inf(1), math.Inf(1)},
	})
}

func TestDistancesDirectionAndEmptySelections(t *testing.T) {
	graph := newPathTestGraph(t)
	source, _ := VertexIDs(3)
	target, _ := VertexIDs(0)

	incoming, err := graph.Distances(source, target, PathOptions{Direction: DirectionIn})
	if err != nil {
		t.Fatal(err)
	}
	assertMatrixRows(t, incoming, [][]float64{{2}})
	outgoing, err := graph.Distances(source, target, PathOptions{Direction: DirectionOut})
	if err != nil {
		t.Fatal(err)
	}
	assertMatrixRows(t, outgoing, [][]float64{{math.Inf(1)}})
	all, err := graph.Distances(source, target, PathOptions{Direction: DirectionAll})
	if err != nil {
		t.Fatal(err)
	}
	assertMatrixRows(t, all, [][]float64{{2}})

	emptyRows, err := graph.Distances(NoVertices(), AllVertices(), PathOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if rows, columns := emptyRows.Dims(); rows != 0 || columns != 5 {
		t.Errorf("empty-row dimensions = (%d, %d), want (0, 5)", rows, columns)
	}
	duplicateTargets, _ := VertexIDs(1, 1, 3)
	emptyRowsWithDuplicates, err := graph.Distances(NoVertices(), duplicateTargets, PathOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if rows, columns := emptyRowsWithDuplicates.Dims(); rows != 0 || columns != 3 {
		t.Errorf("duplicate-target empty-row dimensions = (%d, %d), want (0, 3)", rows, columns)
	}
	emptyColumns, err := graph.Distances(AllVertices(), NoVertices(), PathOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if rows, columns := emptyColumns.Dims(); rows != 5 || columns != 0 {
		t.Errorf("empty-column dimensions = (%d, %d), want (5, 0)", rows, columns)
	}

	emptyGraph, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = emptyGraph.Close() })
	empty, err := emptyGraph.Distances(AllVertices(), AllVertices(), PathOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if rows, columns := empty.Dims(); rows != 0 || columns != 0 {
		t.Errorf("empty-graph dimensions = (%d, %d), want (0, 0)", rows, columns)
	}
}

func TestShortestPathUnweightedWeightedAndUnreachable(t *testing.T) {
	graph := newPathTestGraph(t)

	unweighted, err := graph.ShortestPath(0, 3, PathOptions{Direction: DirectionOut})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(unweighted, Path{
		Vertices: []int{0, 2, 3},
		Edges:    []int{2, 3},
		Found:    true,
	}) {
		t.Errorf("unweighted path = %#v", unweighted)
	}
	assertPathEdges(t, graph, unweighted, DirectionOut)

	weighted, err := graph.ShortestPath(0, 3, PathOptions{
		Direction: DirectionOut,
		Weights:   []float64{2, 3, 10, 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(weighted, Path{
		Vertices: []int{0, 1, 2, 3},
		Edges:    []int{0, 1, 3},
		Found:    true,
	}) {
		t.Errorf("weighted path = %#v", weighted)
	}
	assertPathEdges(t, graph, weighted, DirectionOut)

	zero, err := graph.ShortestPath(2, 2, PathOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(zero, Path{Vertices: []int{2}, Edges: []int{}, Found: true}) {
		t.Errorf("zero-length path = %#v", zero)
	}
	unreachable, err := graph.ShortestPath(0, 4, PathOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if unreachable.Found || unreachable.Vertices == nil || unreachable.Edges == nil || len(unreachable.Vertices) != 0 || len(unreachable.Edges) != 0 {
		t.Errorf("unreachable path = %#v, want explicit non-nil empty result", unreachable)
	}
}

func TestShortestPathDirectionAndGoOwnership(t *testing.T) {
	graph := newPathTestGraph(t)
	path, err := graph.ShortestPath(3, 0, PathOptions{Direction: DirectionIn})
	if err != nil {
		t.Fatal(err)
	}
	if !path.Found || !reflect.DeepEqual(path.Vertices, []int{3, 2, 0}) {
		t.Errorf("incoming path = %#v", path)
	}
	assertPathEdges(t, graph, path, DirectionIn)
	distances, err := graph.Distances(AllVertices(), AllVertices(), PathOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	path.Vertices[0] = 99
	if got, _ := distances.At(0, 3); got != 2 {
		t.Errorf("Go-owned distance after graph close = %v, want 2", got)
	}
}

func TestPathsRejectInvalidInputsAndClosedGraphs(t *testing.T) {
	graph := newPathTestGraph(t)
	invalidSelector, _ := VertexIDs(5)
	if _, err := graph.Distances(invalidSelector, AllVertices(), PathOptions{}); err == nil || !strings.Contains(err.Error(), "source selector") {
		t.Errorf("invalid source selector error = %v", err)
	}
	if _, err := graph.Distances(AllVertices(), invalidSelector, PathOptions{}); err == nil || !strings.Contains(err.Error(), "target selector") {
		t.Errorf("invalid target selector error = %v", err)
	}
	for _, options := range []PathOptions{
		{Direction: DirectionMode(99)},
		{Weights: []float64{}},
		{Weights: []float64{1, 2, 3}},
		{Weights: []float64{1, 2, 3, math.NaN()}},
		{Weights: []float64{1, 2, 3, math.Inf(1)}},
	} {
		if _, err := graph.Distances(AllVertices(), AllVertices(), options); err == nil {
			t.Errorf("Distances(%#v) error = nil", options)
		}
		if _, err := graph.ShortestPath(0, 3, options); err == nil {
			t.Errorf("ShortestPath(%#v) error = nil", options)
		}
	}
	for _, endpoint := range [][2]int{{-1, 0}, {0, -1}, {5, 0}, {0, 5}} {
		if _, err := graph.ShortestPath(endpoint[0], endpoint[1], PathOptions{}); err == nil {
			t.Errorf("ShortestPath(%d, %d) error = nil", endpoint[0], endpoint[1])
		}
	}

	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	assertPathsClosed(t, graph)
	var nilGraph *Graph
	assertPathsClosed(t, nilGraph)
}

func TestPathsUseNegativeWeightsAndPropagateNegativeCycleErrors(t *testing.T) {
	weighted, err := NewGraphFromEdges(4, []Edge{{0, 1}, {0, 2}, {2, 1}, {1, 3}}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = weighted.Close() })
	path, err := weighted.ShortestPath(0, 3, PathOptions{
		Direction: DirectionOut,
		Weights:   []float64{2, 5, -4, 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(path, Path{
		Vertices: []int{0, 2, 1, 3},
		Edges:    []int{1, 2, 3},
		Found:    true,
	}) {
		t.Errorf("negative-weight path = %#v", path)
	}
	assertPathEdges(t, weighted, path, DirectionOut)

	graph, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}, {2, 1}}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	_, err = graph.Distances(AllVertices(), AllVertices(), PathOptions{
		Direction: DirectionOut,
		Weights:   []float64{1, -3, 1},
	})
	if err == nil {
		t.Fatal("negative-cycle Distances error = nil")
	}
	if _, err := graph.ShortestPath(0, 2, PathOptions{
		Direction: DirectionOut,
		Weights:   []float64{1, -3, 1},
	}); err == nil {
		t.Fatal("negative-cycle ShortestPath error = nil")
	}
}

func TestPathsOnUndirectedGraph(t *testing.T) {
	graph, err := NewGraphFromEdges(4, []Edge{{0, 1}, {1, 2}, {2, 3}}, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	source, _ := VertexIDs(3)
	target, _ := VertexIDs(0)

	distances, err := graph.Distances(source, target, PathOptions{
		Direction: DirectionIn,
		Weights:   []float64{2, 3, 4},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertMatrixRows(t, distances, [][]float64{{9}})

	path, err := graph.ShortestPath(3, 0, PathOptions{Direction: DirectionOut})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(path, Path{
		Vertices: []int{3, 2, 1, 0},
		Edges:    []int{2, 1, 0},
		Found:    true,
	}) {
		t.Errorf("undirected path = %#v", path)
	}
	assertPathEdges(t, graph, path, DirectionOut)
}

func newPathTestGraph(t *testing.T) *Graph {
	t.Helper()
	graph, err := NewGraphFromEdges(5, []Edge{{0, 1}, {1, 2}, {0, 2}, {2, 3}}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	return graph
}

func assertPathsClosed(t *testing.T, graph *Graph) {
	t.Helper()
	if _, err := graph.Distances(AllVertices(), AllVertices(), PathOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("Distances closed error = %v, want %v", err, ErrClosed)
	}
	if _, err := graph.ShortestPath(0, 0, PathOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("ShortestPath closed error = %v, want %v", err, ErrClosed)
	}
}

func assertMatrixRows(t *testing.T, matrix Matrix, want [][]float64) {
	t.Helper()
	got := matrix.Rows()
	if len(got) != len(want) {
		t.Fatalf("matrix row count = %d, want %d", len(got), len(want))
	}
	for row := range want {
		if len(got[row]) != len(want[row]) {
			t.Fatalf("matrix row %d length = %d, want %d", row, len(got[row]), len(want[row]))
		}
		for column := range want[row] {
			if got[row][column] != want[row][column] {
				t.Errorf("matrix[%d][%d] = %v, want %v", row, column, got[row][column], want[row][column])
			}
		}
	}
}

func assertPathEdges(t *testing.T, graph *Graph, path Path, direction DirectionMode) {
	t.Helper()
	if len(path.Edges)+1 != len(path.Vertices) {
		t.Fatalf("path has %d edges and %d vertices", len(path.Edges), len(path.Vertices))
	}
	directed, err := graph.IsDirected()
	if err != nil {
		t.Fatal(err)
	}
	for index, edgeID := range path.Edges {
		from, to, err := graph.EdgeEndpoints(edgeID)
		if err != nil {
			t.Fatal(err)
		}
		stepFrom, stepTo := path.Vertices[index], path.Vertices[index+1]
		matchesForward := from == stepFrom && to == stepTo
		matchesReverse := from == stepTo && to == stepFrom
		matches := matchesForward
		if !directed || direction == DirectionAll {
			matches = matchesForward || matchesReverse
		} else if direction == DirectionIn {
			matches = matchesReverse
		}
		if !matches {
			t.Errorf("edge %d endpoints = (%d, %d), path step = (%d, %d)", edgeID, from, to, path.Vertices[index], path.Vertices[index+1])
		}
	}
}

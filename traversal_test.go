package igraph

import (
	"errors"
	"reflect"
	"testing"
)

func TestBreadthFirstSearchDirected(t *testing.T) {
	graph := newTraversalTestGraph(t, true)

	result, err := graph.BreadthFirstSearch(BFSOptions{
		Roots:     []int{0},
		Direction: DirectionOut,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertBFSResult(t, result, BFSResult{
		Order:     []int{0, 1, 2, 3, 4},
		Parents:   []int{-1, 0, 0, 1, 2, -2, -2},
		Distances: []int{0, 1, 1, 2, 2, -1, -1},
	})

	incoming, err := graph.BreadthFirstSearch(BFSOptions{
		Roots:     []int{3},
		Direction: DirectionIn,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertBFSResult(t, incoming, BFSResult{
		Order:     []int{3, 1, 0},
		Parents:   []int{1, 3, -2, -1, -2, -2, -2},
		Distances: []int{2, 1, -1, 0, -1, -1, -1},
	})
}

func TestBreadthFirstSearchMultipleRootsAndUnreachable(t *testing.T) {
	graph := newTraversalTestGraph(t, true)

	multiple, err := graph.BreadthFirstSearch(BFSOptions{
		Roots:     []int{5, 0},
		Direction: DirectionOut,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertBFSResult(t, multiple, BFSResult{
		Order:     []int{5, 6, 0, 1, 2, 3, 4},
		Parents:   []int{-1, 0, 0, 1, 2, -1, 5},
		Distances: []int{0, 1, 1, 2, 2, 0, 1},
	})

	all, err := graph.BreadthFirstSearch(BFSOptions{
		Roots:               []int{0},
		Direction:           DirectionOut,
		TraverseUnreachable: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertBFSResult(t, all, BFSResult{
		Order:     []int{0, 1, 2, 3, 4, 5, 6},
		Parents:   []int{-1, 0, 0, 1, 2, -1, 5},
		Distances: []int{0, 1, 1, 2, 2, 0, 1},
	})
}

func TestBreadthFirstSearchRestriction(t *testing.T) {
	graph := newTraversalTestGraph(t, true)
	restriction, err := VertexIDs(0, 2, 4)
	if err != nil {
		t.Fatal(err)
	}
	result, err := graph.BreadthFirstSearch(BFSOptions{
		Roots:       []int{0},
		Direction:   DirectionOut,
		Restriction: restriction,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertBFSResult(t, result, BFSResult{
		Order:     []int{0, 2, 4},
		Parents:   []int{-1, -2, 0, -2, 2, -2, -2},
		Distances: []int{0, -1, 1, -1, 2, -1, -1},
	})
}

func TestTraversalsUndirectedGraphWithLoop(t *testing.T) {
	graph, err := NewGraphFromEdges(4, []Edge{{0, 1}, {1, 1}, {1, 2}, {2, 3}}, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })

	for _, direction := range []DirectionMode{DirectionOut, DirectionIn, DirectionAll} {
		bfs, err := graph.BreadthFirstSearch(BFSOptions{Roots: []int{0}, Direction: direction})
		if err != nil {
			t.Fatal(err)
		}
		assertBFSResult(t, bfs, BFSResult{
			Order:     []int{0, 1, 2, 3},
			Parents:   []int{-1, 0, 1, 2},
			Distances: []int{0, 1, 2, 3},
		})

		dfs, err := graph.DepthFirstSearch(DFSOptions{Root: 0, Direction: direction})
		if err != nil {
			t.Fatal(err)
		}
		assertDFSResult(t, dfs, DFSResult{
			Order:       []int{0, 1, 2, 3},
			FinishOrder: []int{3, 2, 1, 0},
			Parents:     []int{-1, 0, 1, 2},
			Distances:   []int{0, 1, 2, 3},
		})
	}
}

func TestDepthFirstSearchDirected(t *testing.T) {
	graph := newTraversalTestGraph(t, true)

	result, err := graph.DepthFirstSearch(DFSOptions{Root: 0, Direction: DirectionOut})
	if err != nil {
		t.Fatal(err)
	}
	assertDFSResult(t, result, DFSResult{
		Order:       []int{0, 1, 3, 2, 4},
		FinishOrder: []int{3, 1, 4, 2, 0},
		Parents:     []int{-1, 0, 0, 1, 2, -2, -2},
		Distances:   []int{0, 1, 1, 2, 2, -1, -1},
	})

	incoming, err := graph.DepthFirstSearch(DFSOptions{Root: 3, Direction: DirectionIn})
	if err != nil {
		t.Fatal(err)
	}
	assertDFSResult(t, incoming, DFSResult{
		Order:       []int{3, 1, 0},
		FinishOrder: []int{0, 1, 3},
		Parents:     []int{1, 3, -2, -1, -2, -2, -2},
		Distances:   []int{2, 1, -1, 0, -1, -1, -1},
	})
}

func TestDepthFirstSearchUnreachable(t *testing.T) {
	graph := newTraversalTestGraph(t, true)
	result, err := graph.DepthFirstSearch(DFSOptions{
		Root:                0,
		Direction:           DirectionOut,
		TraverseUnreachable: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertDFSResult(t, result, DFSResult{
		Order:       []int{0, 1, 3, 2, 4, 5, 6},
		FinishOrder: []int{3, 1, 4, 2, 0, 6, 5},
		Parents:     []int{-1, 0, 0, 1, 2, -1, 5},
		Distances:   []int{0, 1, 1, 2, 2, 0, 1},
	})
}

func TestTraversalsReturnGoOwnedResults(t *testing.T) {
	graph := newTraversalTestGraph(t, true)
	bfs, err := graph.BreadthFirstSearch(BFSOptions{Roots: []int{0}, Direction: DirectionOut})
	if err != nil {
		t.Fatal(err)
	}
	dfs, err := graph.DepthFirstSearch(DFSOptions{Root: 0, Direction: DirectionOut})
	if err != nil {
		t.Fatal(err)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(bfs.Order, []int{0, 1, 2, 3, 4}) {
		t.Errorf("BFS order after Close = %v", bfs.Order)
	}
	if !reflect.DeepEqual(dfs.FinishOrder, []int{3, 1, 4, 2, 0}) {
		t.Errorf("DFS finish order after Close = %v", dfs.FinishOrder)
	}
}

func TestTraversalsRejectEmptyGraphs(t *testing.T) {
	graph, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	if _, err := graph.BreadthFirstSearch(BFSOptions{Roots: []int{0}}); err == nil {
		t.Error("BreadthFirstSearch on empty graph error = nil")
	}
	if _, err := graph.BreadthFirstSearch(BFSOptions{}); err == nil {
		t.Error("BreadthFirstSearch without roots error = nil")
	}
	if _, err := graph.DepthFirstSearch(DFSOptions{}); err == nil {
		t.Error("DepthFirstSearch on empty graph error = nil")
	}
}

func TestTraversalsRejectInvalidOptions(t *testing.T) {
	graph := newTraversalTestGraph(t, true)

	bfsCases := []BFSOptions{
		{},
		{Roots: []int{-1}},
		{Roots: []int{7}},
		{Roots: []int{0}, Direction: DirectionMode(99)},
		{Roots: []int{0}, Restriction: NoVertices()},
		{Roots: []int{0}, Restriction: VertexSelector{kind: vertexSelectorKind(99)}},
		{Roots: []int{0}, Restriction: VertexSelector{kind: vertexSelectorIDs, ids: []int{-1, 0}}},
		{Roots: []int{0}, Restriction: VertexSelector{kind: vertexSelectorRange, start: -1, end: 0}},
		{Roots: []int{0}, Restriction: VertexSelector{kind: vertexSelectorRange, start: 2, end: 1}},
	}
	for index, options := range bfsCases {
		if _, err := graph.BreadthFirstSearch(options); err == nil {
			t.Errorf("BreadthFirstSearch invalid case %d error = nil", index)
		}
	}
	for _, options := range []DFSOptions{
		{Root: -1},
		{Root: 7},
		{Root: 0, Direction: DirectionMode(99)},
	} {
		if _, err := graph.DepthFirstSearch(options); err == nil {
			t.Errorf("DepthFirstSearch(%+v) error = nil", options)
		}
	}
}

func TestTraversalsRejectClosedAndNilGraphs(t *testing.T) {
	graph := newTraversalTestGraph(t, true)
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	for _, graph := range []*Graph{graph, nil} {
		if _, err := graph.BreadthFirstSearch(BFSOptions{Roots: []int{0}}); !errors.Is(err, ErrClosed) {
			t.Errorf("BreadthFirstSearch error = %v, want %v", err, ErrClosed)
		}
		if _, err := graph.DepthFirstSearch(DFSOptions{}); !errors.Is(err, ErrClosed) {
			t.Errorf("DepthFirstSearch error = %v, want %v", err, ErrClosed)
		}
	}
}

func newTraversalTestGraph(t *testing.T, directed bool) *Graph {
	t.Helper()
	graph, err := NewGraphFromEdges(7, []Edge{
		{From: 0, To: 1},
		{From: 0, To: 2},
		{From: 1, To: 1},
		{From: 1, To: 3},
		{From: 2, To: 4},
		{From: 5, To: 6},
	}, directed)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	return graph
}

func assertBFSResult(t *testing.T, got, want BFSResult) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BFS result = %+v, want %+v", got, want)
	}
}

func assertDFSResult(t *testing.T, got, want DFSResult) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("DFS result = %+v, want %+v", got, want)
	}
}

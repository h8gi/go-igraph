package igraph

import (
	"errors"
	"reflect"
	"testing"
)

func TestReachabilityAndCounts(t *testing.T) {
	graph := newReachabilityTestGraph(t)
	result, err := graph.Reachability(DirectionOut)
	if err != nil {
		t.Fatal(err)
	}
	wantSets := [][]int{{0, 1, 2}, {1, 2}, {2}, {2, 3}}
	if !reflect.DeepEqual(result.Reachable, wantSets) {
		t.Errorf("reachable = %v, want %v", result.Reachable, wantSets)
	}
	if result.ComponentCount != 4 || !reflect.DeepEqual(result.Sizes, []int{1, 1, 1, 1}) || len(result.Membership) != 4 {
		t.Errorf("component metadata = %#v", result)
	}
	counts, err := graph.ReachableCounts(DirectionOut)
	if err != nil || !reflect.DeepEqual(counts, []int{3, 2, 1, 2}) {
		t.Errorf("out counts = %v, %v", counts, err)
	}
	inCounts, err := graph.ReachableCounts(DirectionIn)
	if err != nil || !reflect.DeepEqual(inCounts, []int{1, 2, 4, 1}) {
		t.Errorf("in counts = %v, %v", inCounts, err)
	}
}

func TestNeighborhoodGraphsPreserveRootsProvenanceAndOwnership(t *testing.T) {
	graph := newReachabilityTestGraph(t)
	roots, _ := VertexIDs(2, 0, 2)
	results, err := graph.NeighborhoodGraphs(roots, NeighborhoodOptions{
		Order: 1, Direction: DirectionIn,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatalf("result count = %d", len(results))
	}
	wantVertices := [][]int{{2, 1, 3}, {0}, {2, 1, 3}}
	for i, result := range results {
		if result.Root != []int{2, 0, 2}[i] || !reflect.DeepEqual(result.SourceVertices, wantVertices[i]) {
			t.Errorf("result %d = root %d provenance %v", i, result.Root, result.SourceVertices)
		}
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	for i, result := range results {
		got, err := result.Graph.VertexCount()
		if err != nil {
			t.Fatal(err)
		}
		if got != len(wantVertices[i]) {
			t.Errorf("owned graph %d vertex count = %d", i, got)
		}
		if err := result.Graph.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestTransitiveClosureIsIndependent(t *testing.T) {
	graph := newReachabilityTestGraph(t)
	closure, err := graph.TransitiveClosure()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = closure.Close() })
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	vertices, vertexErr := closure.VertexCount()
	edges, edgeErr := closure.EdgeCount()
	if vertexErr != nil || edgeErr != nil || vertices != 4 || edges != 4 {
		t.Errorf("closure counts = (%d, %d), errors (%v, %v), want (4, 4)", vertices, edges, vertexErr, edgeErr)
	}
	distances, err := closure.Distances(AllVertices(), AllVertices(), PathOptions{Direction: DirectionOut})
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := distances.At(0, 2); got != 1 {
		t.Errorf("closure distance 0->2 = %v, want 1", got)
	}
}

func TestReachabilityEmptyInvalidAndClosed(t *testing.T) {
	empty, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	result, err := empty.Reachability(DirectionOut)
	if err != nil || result.ComponentCount != 0 || result.Membership == nil || result.Sizes == nil || result.Reachable == nil {
		t.Errorf("empty reachability = %#v, %v", result, err)
	}
	graphs, err := empty.NeighborhoodGraphs(AllVertices(), NeighborhoodOptions{})
	if err != nil || graphs == nil || len(graphs) != 0 {
		t.Errorf("empty neighborhood graphs = %v, %v", graphs, err)
	}
	if _, err := empty.ReachableCounts(DirectionMode(99)); err == nil {
		t.Error("invalid direction accepted")
	}
	if err := empty.Close(); err != nil {
		t.Fatal(err)
	}
	assertReachabilityClosed(t, empty)
	var nilGraph *Graph
	assertReachabilityClosed(t, nilGraph)
}

func assertReachabilityClosed(t *testing.T, graph *Graph) {
	t.Helper()
	_, reachErr := graph.Reachability(DirectionOut)
	_, countErr := graph.ReachableCounts(DirectionOut)
	_, graphErr := graph.NeighborhoodGraphs(AllVertices(), NeighborhoodOptions{})
	_, closureErr := graph.TransitiveClosure()
	for i, err := range []error{reachErr, countErr, graphErr, closureErr} {
		if !errors.Is(err, ErrClosed) {
			t.Errorf("closed check %d error = %v", i, err)
		}
	}
}

func newReachabilityTestGraph(t *testing.T) *Graph {
	t.Helper()
	graph, err := NewGraphFromEdges(4, []Edge{{0, 1}, {1, 2}, {3, 2}}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	return graph
}

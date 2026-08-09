package igraph

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestMilestone3DirectedWeightedSelectorIntegration(t *testing.T) {
	graph, err := NewGraphFromEdges(6, []Edge{
		{0, 1}, {0, 2}, {1, 2}, {2, 0}, {2, 3}, {3, 3},
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	weights := []float64{2, 10, 1, 4, 3, 5}
	sources, _ := VertexIDs(0, 2, 0)
	targets, _ := VertexIDs(3, 4, 3)

	unweighted, err := graph.Distances(sources, targets, PathOptions{Direction: DirectionOut})
	if err != nil {
		t.Fatal(err)
	}
	assertMatrixRows(t, unweighted, [][]float64{
		{2, math.Inf(1), 2},
		{1, math.Inf(1), 1},
		{2, math.Inf(1), 2},
	})
	weighted, err := graph.Distances(sources, targets, PathOptions{
		Direction: DirectionOut,
		Weights:   weights,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertMatrixRows(t, weighted, [][]float64{
		{6, math.Inf(1), 6},
		{3, math.Inf(1), 3},
		{6, math.Inf(1), 6},
	})

	path, err := graph.ShortestPath(0, 3, PathOptions{
		Direction: DirectionOut,
		Weights:   weights,
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := (Path{Vertices: []int{0, 1, 2, 3}, Edges: []int{0, 2, 4}, Found: true}); !reflect.DeepEqual(path, want) {
		t.Errorf("ShortestPath = %#v, want %#v", path, want)
	}

	vertices, _ := VertexIDs(2, 0, 2, 3)
	degrees, err := graph.Degree(vertices, DegreeOptions{Direction: DirectionAll, CountLoops: true})
	if err != nil {
		t.Fatal(err)
	}
	if want := []int{4, 3, 4, 3}; !reflect.DeepEqual(degrees, want) {
		t.Errorf("Degree = %v, want %v", degrees, want)
	}
	restriction, _ := VertexIDs(0, 1, 2, 2, 3)
	bfs, err := graph.BreadthFirstSearch(BFSOptions{
		Roots:       []int{0},
		Direction:   DirectionOut,
		Restriction: restriction,
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := []int{0, 1, 2, 3}; !reflect.DeepEqual(bfs.Order, want) {
		t.Errorf("BFS order = %v, want %v", bfs.Order, want)
	}

	components, err := graph.ConnectedComponents(ConnectednessStrong)
	if err != nil {
		t.Fatal(err)
	}
	assertComponents(t, components, [][]int{{0, 1, 2}, {3}, {4}, {5}})
	if connected, err := graph.IsConnected(ConnectednessWeak); err != nil || connected {
		t.Errorf("IsConnected(weak) = %t, %v, want false, nil", connected, err)
	}

	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	path.Vertices[0] = 99
	bfs.Order[0] = 99
	components.Membership[0] = 99
	if got, _ := weighted.At(0, 0); got != 6 {
		t.Errorf("weighted distance after Close = %v, want 6", got)
	}
	if _, err := graph.Distances(AllVertices(), AllVertices(), PathOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("Distances after Close error = %v, want %v", err, ErrClosed)
	}
}

func TestMilestone3UndirectedLoopDisconnectedIntegration(t *testing.T) {
	graph, err := NewGraphFromEdges(4, []Edge{{0, 0}, {0, 1}, {1, 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })

	for _, direction := range []DirectionMode{DirectionOut, DirectionIn, DirectionAll} {
		path, err := graph.ShortestPath(2, 0, PathOptions{Direction: direction})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(path.Vertices, []int{2, 1, 0}) {
			t.Errorf("ShortestPath direction %d = %v", direction, path.Vertices)
		}
	}
	degrees, err := graph.Degree(AllVertices(), DegreeOptions{CountLoops: true})
	if err != nil {
		t.Fatal(err)
	}
	if want := []int{3, 2, 1, 0}; !reflect.DeepEqual(degrees, want) {
		t.Errorf("Degree = %v, want %v", degrees, want)
	}
	diameter, err := graph.Diameter(DistanceSummaryOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsInf(diameter.Length, 1) || diameter.Path.Found {
		t.Errorf("disconnected Diameter = %#v", diameter)
	}
	local, err := graph.LocalTransitivity(NoVertices(), TransitivityNaN)
	if err != nil || local == nil || len(local) != 0 {
		t.Errorf("empty LocalTransitivity = %#v, %v", local, err)
	}
}

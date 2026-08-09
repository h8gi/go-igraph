package igraph

import (
	"errors"
	"reflect"
	"testing"
)

func TestDegreeUndirectedLoopsParallelEdgesAndSelectionOrder(t *testing.T) {
	graph, err := NewGraphFromEdges(5, []Edge{
		{0, 1}, {0, 1}, {0, 0}, {1, 2}, {2, 3},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	vertices, err := VertexIDs(2, 0, 2, 4)
	if err != nil {
		t.Fatal(err)
	}

	for _, direction := range []DirectionMode{DirectionOut, DirectionIn, DirectionAll} {
		withoutLoops, err := graph.Degree(vertices, DegreeOptions{Direction: direction})
		if err != nil {
			t.Fatal(err)
		}
		if want := []int{2, 2, 2, 0}; !reflect.DeepEqual(withoutLoops, want) {
			t.Errorf("Degree without loops = %v, want %v", withoutLoops, want)
		}
		withLoops, err := graph.Degree(vertices, DegreeOptions{Direction: direction, CountLoops: true})
		if err != nil {
			t.Fatal(err)
		}
		if want := []int{2, 4, 2, 0}; !reflect.DeepEqual(withLoops, want) {
			t.Errorf("Degree with loops = %v, want %v", withLoops, want)
		}
	}
}

func TestDegreeDirectedModesAndLoops(t *testing.T) {
	graph, err := NewGraphFromEdges(3, []Edge{
		{0, 1}, {0, 1}, {2, 0}, {0, 0}, {1, 2},
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	vertices, err := VertexIDs(0, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		options DegreeOptions
		want    []int
	}{
		{name: "out without loops", options: DegreeOptions{Direction: DirectionOut}, want: []int{2, 1, 2}},
		{name: "in without loops", options: DegreeOptions{Direction: DirectionIn}, want: []int{1, 2, 1}},
		{name: "all without loops", options: DegreeOptions{Direction: DirectionAll}, want: []int{3, 3, 3}},
		{name: "out with loops", options: DegreeOptions{Direction: DirectionOut, CountLoops: true}, want: []int{3, 1, 3}},
		{name: "in with loops", options: DegreeOptions{Direction: DirectionIn, CountLoops: true}, want: []int{2, 2, 2}},
		{name: "all with loops", options: DegreeOptions{Direction: DirectionAll, CountLoops: true}, want: []int{5, 3, 5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := graph.Degree(vertices, tt.options)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Degree = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUndirectedNeighborhoodQueriesPreserveSelection(t *testing.T) {
	graph, err := NewGraphFromEdges(5, []Edge{
		{0, 1}, {0, 1}, {0, 0}, {1, 2}, {2, 3},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	vertices, err := VertexIDs(2, 0, 2, 4)
	if err != nil {
		t.Fatal(err)
	}
	options := NeighborhoodOptions{Order: 1, Direction: DirectionAll}

	sizes, err := graph.NeighborhoodSizes(vertices, options)
	if err != nil {
		t.Fatal(err)
	}
	if want := []int{3, 2, 3, 1}; !reflect.DeepEqual(sizes, want) {
		t.Errorf("NeighborhoodSizes = %v, want %v", sizes, want)
	}
	neighborhoods, err := graph.Neighborhoods(vertices, options)
	if err != nil {
		t.Fatal(err)
	}
	want := [][]int{{2, 1, 3}, {0, 1}, {2, 1, 3}, {4}}
	if !reflect.DeepEqual(neighborhoods, want) {
		t.Errorf("Neighborhoods = %v, want %v", neighborhoods, want)
	}

	// Each result is copied out of the C list into independent Go storage.
	neighborhoods[0][0] = 99
	if !reflect.DeepEqual(neighborhoods[2], want[2]) {
		t.Errorf("duplicate selector results share storage: %v", neighborhoods)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(sizes, []int{3, 2, 3, 1}) || !reflect.DeepEqual(neighborhoods[1], want[1]) {
		t.Errorf("Go-owned results changed after Close: sizes=%v neighborhoods=%v", sizes, neighborhoods)
	}
}

func TestDirectedNeighborhoodModesAndDistanceBounds(t *testing.T) {
	graph, err := NewGraphFromEdges(5, []Edge{
		{0, 1}, {0, 1}, {1, 2}, {2, 3}, {4, 2}, {2, 2},
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	vertexTwo, err := VertexIDs(2)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		options NeighborhoodOptions
		want    [][]int
	}{
		{name: "out", options: NeighborhoodOptions{Order: 1, Direction: DirectionOut}, want: [][]int{{2, 3}}},
		{name: "in", options: NeighborhoodOptions{Order: 1, Direction: DirectionIn}, want: [][]int{{2, 1, 4}}},
		{name: "all", options: NeighborhoodOptions{Order: 1, Direction: DirectionAll}, want: [][]int{{2, 1, 3, 4}}},
		{name: "minimum distance", options: NeighborhoodOptions{Order: 2, MinDistance: 1, Direction: DirectionIn}, want: [][]int{{1, 4, 0}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := graph.Neighborhoods(vertexTwo, tt.options)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Neighborhoods = %v, want %v", got, tt.want)
			}
			sizes, err := graph.NeighborhoodSizes(vertexTwo, tt.options)
			if err != nil {
				t.Fatal(err)
			}
			if want := []int{len(tt.want[0])}; !reflect.DeepEqual(sizes, want) {
				t.Errorf("NeighborhoodSizes = %v, want %v", sizes, want)
			}
		})
	}
}

func TestZeroOrderNeighborhoodsPreserveDuplicateRoots(t *testing.T) {
	graph, err := NewGraphFromEdges(4, []Edge{{0, 1}, {1, 2}, {2, 3}}, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	vertices, err := VertexIDs(2, 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	options := NeighborhoodOptions{Order: 0, MinDistance: 0, Direction: DirectionAll}

	sizes, err := graph.NeighborhoodSizes(vertices, options)
	if err != nil {
		t.Fatal(err)
	}
	if want := []int{1, 1, 1}; !reflect.DeepEqual(sizes, want) {
		t.Errorf("NeighborhoodSizes = %v, want %v", sizes, want)
	}
	neighborhoods, err := graph.Neighborhoods(vertices, options)
	if err != nil {
		t.Fatal(err)
	}
	if want := [][]int{{2}, {0}, {2}}; !reflect.DeepEqual(neighborhoods, want) {
		t.Errorf("Neighborhoods = %v, want %v", neighborhoods, want)
	}
}

func TestAlgorithmQueriesHandleEmptySelections(t *testing.T) {
	graph, err := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	options := NeighborhoodOptions{Order: 2, Direction: DirectionAll}

	degrees, err := graph.Degree(NoVertices(), DegreeOptions{Direction: DirectionAll})
	if err != nil || degrees == nil || len(degrees) != 0 {
		t.Errorf("Degree(NoVertices) = %#v, %v, want non-nil empty slice", degrees, err)
	}
	sizes, err := graph.NeighborhoodSizes(NoVertices(), options)
	if err != nil || sizes == nil || len(sizes) != 0 {
		t.Errorf("NeighborhoodSizes(NoVertices) = %#v, %v, want non-nil empty slice", sizes, err)
	}
	neighborhoods, err := graph.Neighborhoods(NoVertices(), options)
	if err != nil || neighborhoods == nil || len(neighborhoods) != 0 {
		t.Errorf("Neighborhoods(NoVertices) = %#v, %v, want non-nil empty slice", neighborhoods, err)
	}
}

func TestAlgorithmQueriesHandleEmptyGraph(t *testing.T) {
	graph, err := NewGraphFromEdges(0, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })

	selectors := []struct {
		name     string
		selector VertexSelector
	}{
		{name: "all vertices", selector: AllVertices()},
		{name: "no vertices", selector: NoVertices()},
	}
	for _, tt := range selectors {
		t.Run(tt.name, func(t *testing.T) {
			degrees, err := graph.Degree(tt.selector, DegreeOptions{Direction: DirectionAll})
			if err != nil || degrees == nil || len(degrees) != 0 {
				t.Errorf("Degree = %#v, %v, want non-nil empty slice", degrees, err)
			}
			options := NeighborhoodOptions{Order: 0, Direction: DirectionAll}
			sizes, err := graph.NeighborhoodSizes(tt.selector, options)
			if err != nil || sizes == nil || len(sizes) != 0 {
				t.Errorf("NeighborhoodSizes = %#v, %v, want non-nil empty slice", sizes, err)
			}
			neighborhoods, err := graph.Neighborhoods(tt.selector, options)
			if err != nil || neighborhoods == nil || len(neighborhoods) != 0 {
				t.Errorf("Neighborhoods = %#v, %v, want non-nil empty slice", neighborhoods, err)
			}
		})
	}
}

func TestAlgorithmQueriesRejectInvalidInput(t *testing.T) {
	graph, err := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	outOfRange, err := VertexIDs(2)
	if err != nil {
		t.Fatal(err)
	}
	invalidKind := VertexSelector{kind: vertexSelectorKind(255)}

	for _, selector := range []VertexSelector{outOfRange, invalidKind} {
		if _, err := graph.Degree(selector, DegreeOptions{Direction: DirectionAll}); err == nil {
			t.Errorf("Degree(%#v) error = nil", selector)
		}
		if _, err := graph.NeighborhoodSizes(selector, NeighborhoodOptions{Order: 1, Direction: DirectionAll}); err == nil {
			t.Errorf("NeighborhoodSizes(%#v) error = nil", selector)
		}
		if _, err := graph.Neighborhoods(selector, NeighborhoodOptions{Order: 1, Direction: DirectionAll}); err == nil {
			t.Errorf("Neighborhoods(%#v) error = nil", selector)
		}
	}
	if _, err := graph.Degree(AllVertices(), DegreeOptions{Direction: DirectionMode(255)}); err == nil {
		t.Error("Degree invalid direction error = nil")
	}
	invalidOptions := []NeighborhoodOptions{
		{Order: -1, Direction: DirectionAll},
		{Order: 1, MinDistance: -1, Direction: DirectionAll},
		{Order: 1, MinDistance: 2, Direction: DirectionAll},
		{Order: 1, Direction: DirectionMode(255)},
	}
	for _, options := range invalidOptions {
		if _, err := graph.NeighborhoodSizes(AllVertices(), options); err == nil {
			t.Errorf("NeighborhoodSizes(%+v) error = nil", options)
		}
		if _, err := graph.Neighborhoods(AllVertices(), options); err == nil {
			t.Errorf("Neighborhoods(%+v) error = nil", options)
		}
	}
}

func TestAlgorithmQueriesRejectClosedAndNilGraphs(t *testing.T) {
	graph, err := NewGraphFromEdges(1, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	assertAlgorithmQueriesClosed(t, graph)
	var nilGraph *Graph
	assertAlgorithmQueriesClosed(t, nilGraph)
}

func assertAlgorithmQueriesClosed(t *testing.T, graph *Graph) {
	t.Helper()
	if _, err := graph.Degree(AllVertices(), DegreeOptions{Direction: DirectionAll}); !errors.Is(err, ErrClosed) {
		t.Errorf("Degree error = %v, want %v", err, ErrClosed)
	}
	options := NeighborhoodOptions{Order: 1, Direction: DirectionAll}
	if _, err := graph.NeighborhoodSizes(AllVertices(), options); !errors.Is(err, ErrClosed) {
		t.Errorf("NeighborhoodSizes error = %v, want %v", err, ErrClosed)
	}
	if _, err := graph.Neighborhoods(AllVertices(), options); !errors.Is(err, ErrClosed) {
		t.Errorf("Neighborhoods error = %v, want %v", err, ErrClosed)
	}
}

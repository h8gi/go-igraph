package igraph

import (
	"errors"
	"reflect"
	"testing"
)

func TestMaximumCardinalityOrderRoundTrip(t *testing.T) {
	g, err := NewGraphFromEdges(5, []Edge{{0, 1}, {1, 2}, {2, 3}, {3, 0}, {1, 4}}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	order, err := g.MaximumCardinalityOrder()
	if err != nil {
		t.Fatal(err)
	}
	if len(order.Vertices) != 5 || len(order.PositionByVertex) != 5 {
		t.Fatalf("order = %#v", order)
	}
	seen := make([]bool, 5)
	for position, vertex := range order.Vertices {
		if vertex < 0 || vertex >= 5 || seen[vertex] {
			t.Fatalf("invalid order = %#v", order)
		}
		seen[vertex] = true
		if order.PositionByVertex[vertex] != position {
			t.Fatalf("inverse mismatch for %d: %#v", vertex, order)
		}
	}

	empty, _ := NewGraph()
	defer empty.Close()
	emptyOrder, err := empty.MaximumCardinalityOrder()
	if err != nil || emptyOrder.Vertices == nil || emptyOrder.PositionByVertex == nil {
		t.Fatalf("empty order = %#v, %v", emptyOrder, err)
	}
}

func TestChordalityCompletionAndOrdering(t *testing.T) {
	cycle, _ := NewGraphFromEdges(4, []Edge{{0, 1}, {1, 2}, {2, 3}, {3, 0}}, false)
	defer cycle.Close()
	order, err := cycle.MaximumCardinalityOrder()
	if err != nil {
		t.Fatal(err)
	}
	result, err := cycle.Chordality(ChordalityOptions{Ordering: order.Vertices, Complete: true})
	if err != nil {
		t.Fatal(err)
	}
	if result.Chordal || len(result.FillEdges) == 0 || result.Completion == nil {
		t.Fatalf("cycle chordality = %#v", result)
	}
	for _, edge := range result.FillEdges {
		if edge.From >= edge.To {
			t.Fatalf("unnormalized fill edge: %#v", edge)
		}
	}
	completion := result.Completion
	defer completion.Close()
	cycle.Close()
	completed, err := completion.Chordality(ChordalityOptions{})
	if err != nil || !completed.Chordal || completed.FillEdges == nil || len(completed.FillEdges) != 0 {
		t.Fatalf("completion chordality = %#v, %v", completed, err)
	}

	chordal, _ := NewGraphFromEdges(4, []Edge{{0, 1}, {1, 2}, {2, 3}, {3, 0}, {0, 2}}, false)
	defer chordal.Close()
	already, err := chordal.Chordality(ChordalityOptions{Complete: true})
	if err != nil || !already.Chordal || already.FillEdges == nil || len(already.FillEdges) != 0 || already.Completion == nil {
		t.Fatalf("chordal = %#v, %v", already, err)
	}
	defer already.Completion.Close()
}

func TestChordalityDirectedMultigraphAndDegenerateInputs(t *testing.T) {
	directed, _ := NewGraphFromEdges(4, []Edge{{0, 1}, {1, 2}, {2, 3}, {3, 0}}, true)
	defer directed.Close()
	result, err := directed.Chordality(ChordalityOptions{})
	if err != nil || result.Chordal {
		t.Fatalf("directed cycle = %#v, %v", result, err)
	}

	multi, _ := NewGraphFromEdges(3, []Edge{{0, 0}, {0, 1}, {0, 1}, {1, 2}}, false)
	defer multi.Close()
	result, err = multi.Chordality(ChordalityOptions{})
	if err != nil || !result.Chordal {
		t.Fatalf("loop/parallel chordality = %#v, %v", result, err)
	}

	for vertices := 0; vertices <= 1; vertices++ {
		g, _ := NewGraphFromEdges(vertices, nil, false)
		result, err := g.Chordality(ChordalityOptions{})
		g.Close()
		if err != nil || !result.Chordal || result.FillEdges == nil || len(result.FillEdges) != 0 {
			t.Fatalf("%d-vertex chordality = %#v, %v", vertices, result, err)
		}
	}
}

func TestChordalityRejectsInvalidOrdering(t *testing.T) {
	g, _ := NewGraphFromEdges(3, nil, false)
	defer g.Close()
	for _, ordering := range [][]int{{0, 1}, {0, 1, 1}, {0, 1, 3}, {-1, 1, 2}} {
		if _, err := g.Chordality(ChordalityOptions{Ordering: ordering}); err == nil {
			t.Fatalf("ordering %v accepted", ordering)
		}
	}
	valid := []int{2, 0, 1}
	result, err := g.Chordality(ChordalityOptions{Ordering: valid})
	if err != nil || !result.Chordal || !reflect.DeepEqual(valid, []int{2, 0, 1}) {
		t.Fatalf("valid ordering = %#v, %v", result, err)
	}
}

func TestPerfectGraphRecognition(t *testing.T) {
	cases := []struct {
		name     string
		vertices int
		edges    []Edge
		want     bool
	}{
		{"empty", 0, nil, true}, {"singleton", 1, nil, true},
		{"five-cycle", 5, []Edge{{0, 1}, {1, 2}, {2, 3}, {3, 4}, {4, 0}}, false},
		{"six-cycle", 6, []Edge{{0, 1}, {1, 2}, {2, 3}, {3, 4}, {4, 5}, {5, 0}}, true},
		{"chordal", 4, []Edge{{0, 1}, {1, 2}, {2, 3}, {3, 0}, {0, 2}}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g, _ := NewGraphFromEdges(tc.vertices, tc.edges, false)
			defer g.Close()
			got, err := g.IsPerfect()
			if err != nil || got != tc.want {
				t.Fatalf("IsPerfect = %v, %v; want %v", got, err, tc.want)
			}
		})
	}
	directed, _ := NewGraphFromEdges(2, []Edge{{0, 1}}, true)
	defer directed.Close()
	if _, err := directed.IsPerfect(); err == nil {
		t.Fatal("directed graph accepted")
	}
	looped, _ := NewGraphFromEdges(2, []Edge{{0, 0}}, false)
	defer looped.Close()
	if _, err := looped.IsPerfect(); err == nil {
		t.Fatal("looped graph accepted")
	}
	parallel, _ := NewGraphFromEdges(2, []Edge{{0, 1}, {0, 1}}, false)
	defer parallel.Close()
	if _, err := parallel.IsPerfect(); err == nil {
		t.Fatal("parallel graph accepted")
	}
}

func TestChordalUseAfterCloseAndNil(t *testing.T) {
	var nilGraph *Graph
	if _, err := nilGraph.MaximumCardinalityOrder(); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil MCS = %v", err)
	}
	if _, err := nilGraph.Chordality(ChordalityOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil chordality = %v", err)
	}
	if _, err := nilGraph.IsPerfect(); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil perfect = %v", err)
	}
	g, _ := NewGraph()
	g.Close()
	if _, err := g.MaximumCardinalityOrder(); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed MCS = %v", err)
	}
	if _, err := g.Chordality(ChordalityOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed chordality = %v", err)
	}
	if _, err := g.IsPerfect(); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed perfect = %v", err)
	}
}

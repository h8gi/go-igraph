package igraph_test

import (
	"errors"
	"math"
	"reflect"
	"sort"
	"sync"
	"testing"

	"github.com/h8gi/go-igraph"
)

func TestIndependentEdgeAssignmentGame(t *testing.T) {
	seed := uint64(341)
	first, err := igraph.IndependentEdgeAssignmentGame(3, 20, true, false, igraph.ErdosRenyiOptions{Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := igraph.IndependentEdgeAssignmentGame(3, 20, true, false, igraph.ErdosRenyiOptions{Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	if !reflect.DeepEqual(mustEdges(t, first), mustEdges(t, second)) {
		t.Fatal("same seed differed")
	}
	if vertices, edges := mustCounts(t, first); vertices != 3 || edges != 20 {
		t.Fatalf("counts=(%d,%d)", vertices, edges)
	}
	seen := make(map[igraph.Edge]bool)
	hasParallel := false
	for _, edge := range mustEdges(t, first) {
		if edge.From == edge.To {
			t.Fatalf("unexpected loop: %#v", edge)
		}
		hasParallel = hasParallel || seen[edge]
		seen[edge] = true
	}
	if !hasParallel {
		t.Fatal("IEA sample unexpectedly had no parallel edge")
	}
	empty, err := igraph.IndependentEdgeAssignmentGame(0, 0, false, false, igraph.ErdosRenyiOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer empty.Close()
	loop, err := igraph.IndependentEdgeAssignmentGame(1, 3, false, true, igraph.ErdosRenyiOptions{Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	defer loop.Close()
	if _, edges := mustCounts(t, loop); edges != 3 {
		t.Fatalf("singleton loop edges=%d", edges)
	}
	for i, call := range []func() error{
		func() error {
			_, err := igraph.IndependentEdgeAssignmentGame(-1, 0, false, false, igraph.ErdosRenyiOptions{})
			return err
		},
		func() error {
			_, err := igraph.IndependentEdgeAssignmentGame(1, -1, false, false, igraph.ErdosRenyiOptions{})
			return err
		},
		func() error {
			_, err := igraph.IndependentEdgeAssignmentGame(0, 1, false, false, igraph.ErdosRenyiOptions{})
			return err
		},
		func() error {
			_, err := igraph.IndependentEdgeAssignmentGame(1, 1, false, false, igraph.ErdosRenyiOptions{})
			return err
		},
	} {
		if call() == nil {
			t.Errorf("validation %d accepted", i)
		}
	}
}

func TestRewireDirectedEdgesEndpointSelectionAndAttributes(t *testing.T) {
	edges := []igraph.Edge{{From: 0, To: 1}, {From: 0, To: 2}, {From: 0, To: 3}, {From: 1, To: 2}, {From: 1, To: 3}, {From: 2, To: 3}}
	seed := uint64(341)
	for _, tc := range []struct {
		name      string
		direction igraph.DirectionMode
		preserved func(igraph.Edge) int
	}{
		{name: "out rewires ends", direction: igraph.DirectionOut, preserved: func(edge igraph.Edge) int { return edge.From }},
		{name: "in rewires starts", direction: igraph.DirectionIn, preserved: func(edge igraph.Edge) int { return edge.To }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			graph, err := igraph.NewGraphFromEdges(4, edges, true)
			if err != nil {
				t.Fatal(err)
			}
			defer graph.Close()
			if err := graph.SetGraphStringAttribute("model", "directed"); err != nil {
				t.Fatal(err)
			}
			if err := graph.SetVertexNumericAttributes("vertex", []float64{0, 1, 2, 3}); err != nil {
				t.Fatal(err)
			}
			if err := graph.SetEdgeNumericAttributes("edge", []float64{0, 1, 2, 3, 4, 5}); err != nil {
				t.Fatal(err)
			}
			before := endpointIDs(edges, tc.preserved)
			if err := graph.RewireDirectedEdges(1, tc.direction, true, igraph.RewireOptions{Seed: &seed}); err != nil {
				t.Fatal(err)
			}
			afterEdges := mustEdges(t, graph)
			if after := endpointIDs(afterEdges, tc.preserved); !reflect.DeepEqual(before, after) {
				t.Fatalf("preserved endpoints=%v, want %v", after, before)
			}
			if got, _ := graph.GraphStringAttribute("model"); got != "directed" {
				t.Fatalf("graph attribute=%q", got)
			}
			if got, _ := graph.VertexNumericAttributes("vertex"); !reflect.DeepEqual(got, []float64{0, 1, 2, 3}) {
				t.Fatalf("vertex attributes=%v", got)
			}
			if got, _ := graph.EdgeNumericAttributes("edge"); !reflect.DeepEqual(got, []float64{0, 1, 2, 3, 4, 5}) {
				t.Fatalf("edge attributes=%v", got)
			}
		})
	}
}

func endpointIDs(edges []igraph.Edge, endpoint func(igraph.Edge) int) []int {
	result := make([]int, len(edges))
	for i, edge := range edges {
		result[i] = endpoint(edge)
	}
	sort.Ints(result)
	return result
}

func TestRewireDirectedEdgesBoundariesReproducibilityAndLifecycle(t *testing.T) {
	edges := []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 3}, {From: 3, To: 0}}
	seed := uint64(341)
	newGraph := func() *igraph.Graph {
		graph, err := igraph.NewGraphFromEdges(4, edges, true)
		if err != nil {
			t.Fatal(err)
		}
		return graph
	}
	first, second := newGraph(), newGraph()
	defer first.Close()
	defer second.Close()
	if err := first.RewireDirectedEdges(1, igraph.DirectionAll, true, igraph.RewireOptions{Seed: &seed}); err != nil {
		t.Fatal(err)
	}
	if err := second.RewireDirectedEdges(1, igraph.DirectionAll, true, igraph.RewireOptions{Seed: &seed}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(mustEdges(t, first), mustEdges(t, second)) {
		t.Fatal("same seed differed")
	}
	unchanged := newGraph()
	defer unchanged.Close()
	if err := unchanged.RewireDirectedEdges(0, igraph.DirectionOut, false, igraph.RewireOptions{}); err != nil {
		t.Fatal(err)
	}
	if got := mustEdges(t, unchanged); !reflect.DeepEqual(got, edges) {
		t.Fatalf("probability zero changed edges: %v", got)
	}
	for _, probability := range []float64{-0.1, 1.1, math.NaN(), math.Inf(1)} {
		if err := unchanged.RewireDirectedEdges(probability, igraph.DirectionOut, false, igraph.RewireOptions{}); err == nil {
			t.Errorf("accepted probability %g", probability)
		}
	}
	if err := unchanged.RewireDirectedEdges(0.5, igraph.DirectionMode(99), false, igraph.RewireOptions{}); err == nil {
		t.Fatal("accepted invalid direction")
	}
	empty, _ := igraph.NewGraphFromEdges(0, nil, true)
	defer empty.Close()
	if err := empty.RewireDirectedEdges(1, igraph.DirectionOut, false, igraph.RewireOptions{}); err != nil {
		t.Fatal(err)
	}
	singleton, _ := igraph.NewGraphFromEdges(1, nil, true)
	defer singleton.Close()
	if err := singleton.RewireDirectedEdges(1, igraph.DirectionIn, false, igraph.RewireOptions{}); err != nil {
		t.Fatal(err)
	}
	closed := newGraph()
	closed.Close()
	if err := closed.RewireDirectedEdges(0.5, igraph.DirectionOut, false, igraph.RewireOptions{}); !errors.Is(err, igraph.ErrClosed) {
		t.Fatalf("closed error=%v", err)
	}
	var nilGraph *igraph.Graph
	if err := nilGraph.RewireDirectedEdges(0.5, igraph.DirectionOut, false, igraph.RewireOptions{}); !errors.Is(err, igraph.ErrClosed) {
		t.Fatalf("nil error=%v", err)
	}
}

func TestRewireDirectedEdgesConcurrentUseAndClose(t *testing.T) {
	graph, err := igraph.NewGraphFromEdges(4, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 3}}, true)
	if err != nil {
		t.Fatal(err)
	}
	var wait sync.WaitGroup
	for i := 0; i < 8; i++ {
		wait.Add(1)
		go func(seed uint64) {
			defer wait.Done()
			err := graph.RewireDirectedEdges(0.5, igraph.DirectionOut, true, igraph.RewireOptions{Seed: &seed})
			if err != nil && !errors.Is(err, igraph.ErrClosed) {
				t.Errorf("rewire error=%v", err)
			}
		}(uint64(i))
	}
	wait.Add(1)
	go func() { defer wait.Done(); _ = graph.Close() }()
	wait.Wait()
}

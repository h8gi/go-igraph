package igraph_test

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func coloringGraph(t *testing.T, directed bool) *igraph.Graph {
	t.Helper()
	g, err := igraph.NewGraphFromEdges(5, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 3}, {From: 3, To: 4}, {From: 4, To: 0}}, directed)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = g.Close() })
	return g
}

func TestGreedyVertexColoring(t *testing.T) {
	for _, heuristic := range []igraph.ColoringHeuristic{igraph.ColoringColoredNeighbors, igraph.ColoringDSatur} {
		g := coloringGraph(t, false)
		colors, err := g.GreedyVertexColoring(heuristic)
		if err != nil {
			t.Fatal(err)
		}
		if len(colors) != 5 {
			t.Fatalf("colors = %v", colors)
		}
		valid, err := g.IsVertexColoring(colors)
		if err != nil || !valid {
			t.Fatalf("IsVertexColoring(%v) = %v, %v", colors, valid, err)
		}
		_ = g.Close()
		if len(colors) != 5 {
			t.Fatal("Go-owned result changed after Close")
		}
	}
	if _, err := coloringGraph(t, false).GreedyVertexColoring(99); err == nil {
		t.Fatal("invalid heuristic accepted")
	}
	empty, err := igraph.NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	defer empty.Close()
	colors, err := empty.GreedyVertexColoring(igraph.ColoringColoredNeighbors)
	if err != nil || colors == nil || len(colors) != 0 {
		t.Fatalf("empty colors = %#v, %v", colors, err)
	}
}

func TestVertexAndEdgeColoringValidation(t *testing.T) {
	g := coloringGraph(t, true)
	if valid, err := g.IsVertexColoring([]int{0, 1, 0, 1, 2}); err != nil || !valid {
		t.Fatalf("valid vertex coloring = %v, %v", valid, err)
	}
	if valid, err := g.IsVertexColoring([]int{0, 0, 1, 0, 1}); err != nil || valid {
		t.Fatalf("invalid vertex coloring = %v, %v", valid, err)
	}
	if valid, err := g.IsEdgeColoring([]int{0, 1, 0, 1, 2}); err != nil || !valid {
		t.Fatalf("valid edge coloring = %v, %v", valid, err)
	}
	if valid, err := g.IsEdgeColoring([]int{0, 0, 1, 2, 1}); err != nil || valid {
		t.Fatalf("invalid edge coloring = %v, %v", valid, err)
	}
	for _, colors := range [][]int{{0}, {0, 1, 2, 3, -1}} {
		if _, err := g.IsVertexColoring(colors); err == nil {
			t.Errorf("invalid colors accepted: %v", colors)
		}
	}
	parallel, err := igraph.NewGraphFromEdges(2, []igraph.Edge{{From: 0, To: 1}, {From: 0, To: 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer parallel.Close()
	if valid, err := parallel.IsEdgeColoring([]int{0, 1}); err != nil || !valid {
		t.Fatalf("parallel coloring = %v, %v", valid, err)
	}
	if valid, err := parallel.IsEdgeColoring([]int{0, 0}); err != nil || valid {
		t.Fatalf("parallel duplicate = %v, %v", valid, err)
	}
}

func TestBipartiteColoringDirectionAndLoops(t *testing.T) {
	partition := igraph.BipartitePartition{false, true}
	for _, test := range []struct {
		edges    []igraph.Edge
		directed bool
		want     igraph.DirectionMode
	}{
		{[]igraph.Edge{{From: 0, To: 1}}, true, igraph.DirectionOut},
		{[]igraph.Edge{{From: 1, To: 0}}, true, igraph.DirectionIn},
		{[]igraph.Edge{{From: 0, To: 1}, {From: 1, To: 0}}, true, igraph.DirectionAll},
		{[]igraph.Edge{{From: 0, To: 1}}, false, igraph.DirectionAll},
	} {
		g, err := igraph.NewGraphFromEdges(2, test.edges, test.directed)
		if err != nil {
			t.Fatal(err)
		}
		result, err := g.IsBipartiteColoring(partition)
		_ = g.Close()
		if err != nil || !result.Valid || result.Direction != test.want {
			t.Errorf("result = %#v, %v, want direction %v", result, err, test.want)
		}
	}
	invalid := coloringGraph(t, false)
	result, err := invalid.IsBipartiteColoring(igraph.BipartitePartition{false, true, false, true, false})
	if err != nil || result.Valid || result.Direction != igraph.DirectionAll {
		t.Fatalf("invalid = %#v, %v", result, err)
	}
	loop, err := igraph.NewGraphFromEdges(1, []igraph.Edge{{From: 0, To: 0}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer loop.Close()
	result, err = loop.IsBipartiteColoring(igraph.BipartitePartition{false})
	if err != nil || !result.Valid {
		t.Fatalf("loop result = %#v, %v", result, err)
	}
}

func TestColoringCloseAndConcurrentReads(t *testing.T) {
	g := coloringGraph(t, false)
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			colors, err := g.GreedyVertexColoring(igraph.ColoringDSatur)
			if err != nil {
				t.Error(err)
				return
			}
			valid, err := g.IsVertexColoring(colors)
			if err != nil || !valid {
				t.Errorf("valid = %v, %v", valid, err)
			}
		}()
	}
	wg.Wait()
	_ = g.Close()
	if _, err := g.GreedyVertexColoring(0); !errors.Is(err, igraph.ErrClosed) {
		t.Errorf("greedy after Close = %v", err)
	}
	if _, err := g.IsVertexColoring(nil); !errors.Is(err, igraph.ErrClosed) {
		t.Errorf("vertex after Close = %v", err)
	}
	if _, err := g.IsEdgeColoring(nil); !errors.Is(err, igraph.ErrClosed) {
		t.Errorf("edge after Close = %v", err)
	}
	if _, err := g.IsBipartiteColoring(nil); !errors.Is(err, igraph.ErrClosed) {
		t.Errorf("bipartite after Close = %v", err)
	}
	var nilGraph *igraph.Graph
	if got, err := nilGraph.GreedyVertexColoring(0); got != nil || !errors.Is(err, igraph.ErrClosed) {
		t.Errorf("nil graph = %v, %v", got, err)
	}
}

func TestColorInputsAreBorrowed(t *testing.T) {
	g := coloringGraph(t, false)
	colors := []int{0, 1, 0, 1, 2}
	before := append([]int{}, colors...)
	_, _ = g.IsVertexColoring(colors)
	if !reflect.DeepEqual(colors, before) {
		t.Fatalf("input mutated: %v", colors)
	}
}

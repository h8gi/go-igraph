package igraph

import (
	"errors"
	"math"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func TestSimpleCyclesDirectedModes(t *testing.T) {
	dag := newCycleTestGraph(t, 3, []Edge{{0, 1}, {2, 1}, {2, 0}}, true)
	defer dag.Close()
	for _, mode := range []DirectionMode{DirectionOut, DirectionIn} {
		result, err := dag.SimpleCycles(SimpleCycleOptions{Direction: mode, MaxResults: 10})
		if err != nil {
			t.Fatal(err)
		}
		if result.Cycles == nil || len(result.Cycles) != 0 || result.Truncated {
			t.Errorf("SimpleCycles(%v) on DAG = %#v", mode, result)
		}
	}
	all, err := dag.SimpleCycles(SimpleCycleOptions{Direction: DirectionAll, MaxResults: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(all.Cycles) != 1 || all.Truncated {
		t.Fatalf("SimpleCycles(DirectionAll) = %#v, want one cycle", all)
	}
	assertCycleWitness(t, dag, all.Cycles[0], DirectionAll)

	directed := newCycleTestGraph(t, 3, []Edge{{0, 1}, {1, 2}, {2, 0}}, true)
	defer directed.Close()
	for _, mode := range []DirectionMode{DirectionOut, DirectionIn, DirectionAll} {
		result, err := directed.SimpleCycles(SimpleCycleOptions{Direction: mode, MaxResults: 10})
		if err != nil {
			t.Fatal(err)
		}
		if len(result.Cycles) == 0 {
			t.Fatalf("SimpleCycles(%v) returned no cycle", mode)
		}
		for _, cycle := range result.Cycles {
			assertCycleWitness(t, directed, cycle, mode)
		}
	}
}

func TestSimpleCyclesKnownSetAndTruncation(t *testing.T) {
	edges := []Edge{{0, 1}, {1, 2}, {2, 0}, {1, 3}, {3, 2}}
	g := newCycleTestGraph(t, 4, edges, false)
	defer g.Close()

	all, err := g.SimpleCycles(SimpleCycleOptions{MaxResults: 10})
	if err != nil {
		t.Fatal(err)
	}
	if all.Truncated || len(all.Cycles) != 3 {
		t.Fatalf("all overlapping cycles = %#v, want 3 non-truncated", all)
	}
	for _, cycle := range all.Cycles {
		assertCycleWitness(t, g, cycle, DirectionAll)
	}
	want := []string{"0,1,2", "0,2,3,4", "1,3,4"}
	got := canonicalCycleEdgeSets(all.Cycles)
	if !slices.Equal(got, want) {
		t.Errorf("cycle edge sets = %v, want %v", got, want)
	}

	exact, err := g.SimpleCycles(SimpleCycleOptions{MaxResults: 3})
	if err != nil {
		t.Fatal(err)
	}
	if exact.Truncated || len(exact.Cycles) != 3 {
		t.Errorf("exact-limit result = %#v", exact)
	}
	limited, err := g.SimpleCycles(SimpleCycleOptions{MaxResults: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !limited.Truncated || len(limited.Cycles) != 1 {
		t.Errorf("one-result limit = %#v", limited)
	}
}

func TestSimpleCyclesInclusiveLengthBounds(t *testing.T) {
	edges := []Edge{
		{0, 0},
		{1, 2}, {2, 1},
		{3, 4}, {4, 5}, {5, 3},
		{6, 7}, {7, 8}, {8, 9}, {9, 6},
	}
	g := newCycleTestGraph(t, 10, edges, true)
	defer g.Close()
	for length := 1; length <= 4; length++ {
		bound := length
		result, err := g.SimpleCycles(SimpleCycleOptions{
			MinLength: &bound, MaxLength: &bound, MaxResults: 10,
		})
		if err != nil {
			t.Fatalf("length %d: %v", length, err)
		}
		if len(result.Cycles) != 1 || len(result.Cycles[0].Vertices) != length || result.Truncated {
			t.Errorf("exact length %d result = %#v", length, result)
		}
		if len(result.Cycles) == 1 {
			assertCycleWitness(t, g, result.Cycles[0], DirectionOut)
		}
	}
	min, max := 2, 3
	result, err := g.SimpleCycles(SimpleCycleOptions{MinLength: &min, MaxLength: &max, MaxResults: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Cycles) != 2 {
		t.Errorf("length range [2,3] returned %d cycles", len(result.Cycles))
	}
	noMatch := 5
	result, err = g.SimpleCycles(SimpleCycleOptions{MinLength: &noMatch, MaxResults: 10})
	if err != nil {
		t.Fatal(err)
	}
	if result.Cycles == nil || len(result.Cycles) != 0 || result.Truncated {
		t.Errorf("no-match result = %#v", result)
	}
}

func TestSimpleCyclesUndirectedDisconnectedLoopsAndParallelEdges(t *testing.T) {
	g := newCycleTestGraph(t, 6, []Edge{
		{0, 0}, {1, 2}, {1, 2}, {1, 2},
		{3, 4}, {4, 5}, {5, 3},
	}, false)
	defer g.Close()
	result, err := g.SimpleCycles(SimpleCycleOptions{Direction: DirectionIn, MaxResults: 20})
	if err != nil {
		t.Fatal(err)
	}
	if result.Truncated || len(result.Cycles) != 5 {
		t.Fatalf("loop/parallel/disconnected enumeration = %#v, want 5 cycles", result)
	}
	for _, cycle := range result.Cycles {
		assertCycleWitness(t, g, cycle, DirectionAll)
	}
	want := []string{"0", "1,2", "1,3", "2,3", "4,5,6"}
	if got := canonicalCycleEdgeSets(result.Cycles); !slices.Equal(got, want) {
		t.Errorf("loop/parallel/disconnected edge sets = %v, want %v", got, want)
	}
}

func TestSimpleCyclesEmptyGraph(t *testing.T) {
	g := newCycleTestGraph(t, 0, nil, true)
	defer g.Close()
	result, err := g.SimpleCycles(SimpleCycleOptions{MaxResults: 1})
	if err != nil {
		t.Fatal(err)
	}
	if result.Cycles == nil || len(result.Cycles) != 0 || result.Truncated {
		t.Errorf("empty graph result = %#v, want non-nil empty non-truncated result", result)
	}
}

func TestSimpleCyclesValidation(t *testing.T) {
	g := newCycleTestGraph(t, 0, nil, true)
	defer g.Close()
	zero, negative := 0, -1
	tests := []SimpleCycleOptions{
		{Direction: DirectionMode(99), MaxResults: 1},
		{},
		{MaxResults: -1},
		{MinLength: &zero, MaxResults: 1},
		{MaxLength: &negative, MaxResults: 1},
		{MinLength: intPointer(3), MaxLength: intPointer(2), MaxResults: 1},
		{MaxResults: math.MaxInt},
	}
	for _, options := range tests {
		if _, err := g.SimpleCycles(options); err == nil {
			t.Errorf("SimpleCycles(%+v) succeeded", options)
		}
	}
}

func TestSimpleCyclesResultSurvivesClose(t *testing.T) {
	g := newCycleTestGraph(t, 3, []Edge{{0, 1}, {1, 2}, {2, 0}}, true)
	result, err := g.SimpleCycles(SimpleCycleOptions{MaxResults: 10})
	if err != nil {
		t.Fatal(err)
	}
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	if len(result.Cycles) != 1 || len(result.Cycles[0].Vertices) != 3 || len(result.Cycles[0].Edges) != 3 {
		t.Fatalf("result after Close = %#v", result)
	}
	result.Cycles[0].Vertices[0] = 99
}

func TestSimpleCyclesUseAfterCloseAndNil(t *testing.T) {
	graphs := []*Graph{nil, newCycleTestGraph(t, 0, nil, true)}
	_ = graphs[1].Close()
	for _, g := range graphs {
		if _, err := g.SimpleCycles(SimpleCycleOptions{MaxResults: 1}); !errors.Is(err, ErrClosed) {
			t.Errorf("SimpleCycles error = %v", err)
		}
	}
}

func canonicalCycleEdgeSets(cycles []Cycle) []string {
	keys := make([]string, len(cycles))
	for i, cycle := range cycles {
		edges := append([]int(nil), cycle.Edges...)
		slices.Sort(edges)
		parts := make([]string, len(edges))
		for j, edge := range edges {
			parts[j] = strconv.Itoa(edge)
		}
		keys[i] = strings.Join(parts, ",")
	}
	slices.Sort(keys)
	return keys
}

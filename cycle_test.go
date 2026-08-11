package igraph

import (
	"errors"
	"math"
	"testing"
)

func TestCyclePredicatesKnownAnswers(t *testing.T) {
	tests := []struct {
		name     string
		vertices int
		edges    []Edge
		directed bool
		acyclic  bool
		dag      bool
	}{
		{name: "empty directed", directed: true, acyclic: true, dag: true},
		{name: "empty undirected", acyclic: true, dag: false},
		{name: "singleton directed", vertices: 1, directed: true, acyclic: true, dag: true},
		{name: "directed acyclic", vertices: 4, edges: []Edge{{0, 1}, {0, 2}, {2, 3}}, directed: true, acyclic: true, dag: true},
		{name: "directed cycle", vertices: 3, edges: []Edge{{0, 1}, {1, 2}, {2, 0}}, directed: true},
		{name: "directed self loop", vertices: 1, edges: []Edge{{0, 0}}, directed: true},
		{name: "undirected tree", vertices: 4, edges: []Edge{{0, 1}, {1, 2}, {1, 3}}, acyclic: true},
		{name: "undirected cycle", vertices: 3, edges: []Edge{{0, 1}, {1, 2}, {2, 0}}},
		{name: "undirected parallel", vertices: 2, edges: []Edge{{0, 1}, {0, 1}}},
		{name: "disconnected with cycle", vertices: 5, edges: []Edge{{0, 1}, {2, 3}, {3, 4}, {4, 2}}, directed: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := newCycleTestGraph(t, tc.vertices, tc.edges, tc.directed)
			defer g.Close()
			acyclic, err := g.IsAcyclic()
			if err != nil {
				t.Fatal(err)
			}
			if acyclic != tc.acyclic {
				t.Errorf("IsAcyclic() = %v, want %v", acyclic, tc.acyclic)
			}
			dag, err := g.IsDAG()
			if err != nil {
				t.Fatal(err)
			}
			if dag != tc.dag {
				t.Errorf("IsDAG() = %v, want %v", dag, tc.dag)
			}
		})
	}
}

func TestTopologicalSortDirectionsAndValidation(t *testing.T) {
	edges := []Edge{{0, 1}, {0, 2}, {1, 3}, {2, 3}}
	g := newCycleTestGraph(t, 4, edges, true)
	defer g.Close()

	for _, mode := range []DirectionMode{DirectionOut, DirectionIn} {
		order, err := g.TopologicalSort(mode)
		if err != nil {
			t.Fatalf("TopologicalSort(%v): %v", mode, err)
		}
		assertTopologicalOrder(t, order, 4, edges, mode)
	}
	if _, err := g.TopologicalSort(DirectionAll); err == nil {
		t.Error("TopologicalSort(DirectionAll) succeeded")
	}
	if _, err := g.TopologicalSort(DirectionMode(99)); err == nil {
		t.Error("TopologicalSort(invalid) succeeded")
	}

	undirected := newCycleTestGraph(t, 2, []Edge{{0, 1}}, false)
	defer undirected.Close()
	if _, err := undirected.TopologicalSort(DirectionOut); err == nil {
		t.Error("TopologicalSort on an undirected graph succeeded")
	}

	cyclic := newCycleTestGraph(t, 3, []Edge{{0, 1}, {1, 2}, {2, 0}}, true)
	defer cyclic.Close()
	if _, err := cyclic.TopologicalSort(DirectionOut); err == nil {
		t.Error("TopologicalSort on a directed cycle succeeded")
	}
}

func TestTopologicalSortPinnedSelfLoopException(t *testing.T) {
	edges := []Edge{{0, 0}, {0, 1}, {1, 2}}
	g := newCycleTestGraph(t, 3, edges, true)
	defer g.Close()

	acyclic, err := g.IsAcyclic()
	if err != nil {
		t.Fatal(err)
	}
	dag, err := g.IsDAG()
	if err != nil {
		t.Fatal(err)
	}
	if acyclic || dag {
		t.Fatalf("self-loop predicates = IsAcyclic %v, IsDAG %v; want both false", acyclic, dag)
	}
	order, err := g.TopologicalSort(DirectionOut)
	if err != nil {
		t.Fatalf("TopologicalSort with only a self-loop cycle: %v", err)
	}
	assertTopologicalOrder(t, order, 3, edges[1:], DirectionOut)
}

func TestTopologicalSortEmptyAndSingleton(t *testing.T) {
	for _, vertices := range []int{0, 1} {
		g := newCycleTestGraph(t, vertices, nil, true)
		order, err := g.TopologicalSort(DirectionOut)
		_ = g.Close()
		if err != nil {
			t.Fatalf("TopologicalSort(%d vertices): %v", vertices, err)
		}
		if len(order) != vertices || order == nil {
			t.Errorf("TopologicalSort(%d vertices) = %#v", vertices, order)
		}
	}
}

func TestFindCycleDirectionsAndEmptyResult(t *testing.T) {
	directed := newCycleTestGraph(t, 4, []Edge{{0, 1}, {1, 2}, {2, 0}, {2, 3}}, true)
	defer directed.Close()
	for _, mode := range []DirectionMode{DirectionOut, DirectionIn, DirectionAll} {
		cycle, err := directed.FindCycle(mode)
		if err != nil {
			t.Fatalf("FindCycle(%v): %v", mode, err)
		}
		assertCycleWitness(t, directed, cycle, mode)
	}

	undirectedTraversal := newCycleTestGraph(t, 3, []Edge{{0, 1}, {2, 1}, {2, 0}}, true)
	defer undirectedTraversal.Close()
	cycle, err := undirectedTraversal.FindCycle(DirectionAll)
	if err != nil {
		t.Fatal(err)
	}
	assertCycleWitness(t, undirectedTraversal, cycle, DirectionAll)
	for _, mode := range []DirectionMode{DirectionOut, DirectionIn} {
		cycle, err := undirectedTraversal.FindCycle(mode)
		if err != nil {
			t.Fatal(err)
		}
		if cycle.Vertices == nil || cycle.Edges == nil || len(cycle.Vertices) != 0 || len(cycle.Edges) != 0 {
			t.Errorf("FindCycle(%v) on DAG = %#v, want non-nil empty slices", mode, cycle)
		}
	}

	if _, err := directed.FindCycle(DirectionMode(99)); err == nil {
		t.Error("FindCycle(invalid) succeeded")
	}
}

func TestFindCycleUndirectedSelfLoopAndParallel(t *testing.T) {
	for _, tc := range []struct {
		name     string
		vertices int
		edges    []Edge
	}{
		{name: "undirected triangle", vertices: 3, edges: []Edge{{0, 1}, {1, 2}, {2, 0}}},
		{name: "self loop", vertices: 1, edges: []Edge{{0, 0}}},
		{name: "parallel", vertices: 2, edges: []Edge{{0, 1}, {0, 1}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCycleTestGraph(t, tc.vertices, tc.edges, false)
			defer g.Close()
			cycle, err := g.FindCycle(DirectionOut)
			if err != nil {
				t.Fatal(err)
			}
			assertCycleWitness(t, g, cycle, DirectionAll)
		})
	}
}

func TestGirthPinnedSemantics(t *testing.T) {
	tests := []struct {
		name     string
		vertices int
		edges    []Edge
		directed bool
		length   float64
	}{
		{name: "empty", length: math.Inf(1)},
		{name: "tree", vertices: 4, edges: []Edge{{0, 1}, {1, 2}, {1, 3}}, length: math.Inf(1)},
		{name: "directed triangle ignores direction", vertices: 3, edges: []Edge{{0, 1}, {2, 1}, {2, 0}}, directed: true, length: 3},
		{name: "four cycle", vertices: 4, edges: []Edge{{0, 1}, {1, 2}, {2, 3}, {3, 0}}, length: 4},
		{name: "disconnected shortest", vertices: 7, edges: []Edge{{0, 1}, {1, 2}, {2, 3}, {3, 0}, {4, 5}, {5, 6}, {6, 4}}, length: 3},
		{name: "loop ignored", vertices: 1, edges: []Edge{{0, 0}}, length: math.Inf(1)},
		{name: "parallel pair ignored", vertices: 2, edges: []Edge{{0, 1}, {0, 1}}, length: math.Inf(1)},
		{name: "loop and parallel ignored beside triangle", vertices: 3, edges: []Edge{{0, 0}, {0, 1}, {0, 1}, {0, 1}, {1, 2}, {2, 0}}, length: 3},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := newCycleTestGraph(t, tc.vertices, tc.edges, tc.directed)
			defer g.Close()
			result, err := g.Girth()
			if err != nil {
				t.Fatal(err)
			}
			if result.Length != tc.length {
				t.Errorf("Girth length = %v, want %v", result.Length, tc.length)
			}
			if result.Vertices == nil {
				t.Error("Girth vertices are nil")
			}
			if math.IsInf(tc.length, 1) && len(result.Vertices) != 0 {
				t.Errorf("acyclic Girth vertices = %v, want empty", result.Vertices)
			}
			if !math.IsInf(tc.length, 1) && len(result.Vertices) != int(tc.length) {
				t.Errorf("Girth vertex length = %d, want %d", len(result.Vertices), int(tc.length))
			}
			if !math.IsInf(tc.length, 1) {
				assertGirthWitness(t, g, result)
			}
		})
	}
}

func TestCycleResultsRemainValidAfterClose(t *testing.T) {
	g := newCycleTestGraph(t, 3, []Edge{{0, 1}, {1, 2}, {2, 0}}, true)
	cycle, err := g.FindCycle(DirectionOut)
	if err != nil {
		t.Fatal(err)
	}
	girth, err := g.Girth()
	if err != nil {
		t.Fatal(err)
	}
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	if len(cycle.Vertices) != 3 || len(cycle.Edges) != 3 || len(girth.Vertices) != 3 {
		t.Fatalf("results changed after Close: cycle %#v, girth %#v", cycle, girth)
	}
	cycle.Vertices[0] = 99
	girth.Vertices[0] = 98
}

func TestCycleAnalysisUseAfterCloseAndNil(t *testing.T) {
	graphs := []*Graph{nil, newCycleTestGraph(t, 0, nil, true)}
	_ = graphs[1].Close()
	for _, g := range graphs {
		if _, err := g.IsAcyclic(); !errors.Is(err, ErrClosed) {
			t.Errorf("IsAcyclic error = %v", err)
		}
		if _, err := g.IsDAG(); !errors.Is(err, ErrClosed) {
			t.Errorf("IsDAG error = %v", err)
		}
		if _, err := g.TopologicalSort(DirectionOut); !errors.Is(err, ErrClosed) {
			t.Errorf("TopologicalSort error = %v", err)
		}
		if _, err := g.FindCycle(DirectionOut); !errors.Is(err, ErrClosed) {
			t.Errorf("FindCycle error = %v", err)
		}
		if _, err := g.Girth(); !errors.Is(err, ErrClosed) {
			t.Errorf("Girth error = %v", err)
		}
	}
}

func newCycleTestGraph(t *testing.T, vertices int, edges []Edge, directed bool) *Graph {
	t.Helper()
	g, err := NewGraphFromEdges(vertices, edges, directed)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func assertTopologicalOrder(t *testing.T, order []int, vertexCount int, edges []Edge, mode DirectionMode) {
	t.Helper()
	if len(order) != vertexCount {
		t.Fatalf("topological order length = %d, want %d", len(order), vertexCount)
	}
	positions := make([]int, vertexCount)
	seen := make([]bool, vertexCount)
	for position, vertex := range order {
		if vertex < 0 || vertex >= vertexCount || seen[vertex] {
			t.Fatalf("invalid topological order: %v", order)
		}
		seen[vertex] = true
		positions[vertex] = position
	}
	for _, edge := range edges {
		from, to := edge.From, edge.To
		if mode == DirectionIn {
			from, to = to, from
		}
		if positions[from] >= positions[to] {
			t.Errorf("order %v violates edge %d -> %d in mode %v", order, edge.From, edge.To, mode)
		}
	}
}

func assertCycleWitness(t *testing.T, g *Graph, cycle Cycle, mode DirectionMode) {
	t.Helper()
	if cycle.Vertices == nil || cycle.Edges == nil {
		t.Fatalf("cycle contains nil slice: %#v", cycle)
	}
	if len(cycle.Vertices) == 0 || len(cycle.Vertices) != len(cycle.Edges) {
		t.Fatalf("invalid cycle lengths: %#v", cycle)
	}
	for i, edgeID := range cycle.Edges {
		edgeFrom, edgeTo, err := g.EdgeEndpoints(edgeID)
		if err != nil {
			t.Fatal(err)
		}
		from := cycle.Vertices[i]
		to := cycle.Vertices[(i+1)%len(cycle.Vertices)]
		valid := edgeFrom == from && edgeTo == to
		if mode == DirectionIn {
			valid = edgeFrom == to && edgeTo == from
		} else if mode == DirectionAll {
			valid = valid || edgeFrom == to && edgeTo == from
		}
		if !valid {
			t.Errorf("edge %d (%d -> %d) does not join witness step %d -> %d in mode %v", edgeID, edgeFrom, edgeTo, from, to, mode)
		}
	}
}

func assertGirthWitness(t *testing.T, g *Graph, result GirthResult) {
	t.Helper()
	vertexCount, err := g.VertexCount()
	if err != nil {
		t.Fatal(err)
	}
	edges, err := g.Edges()
	if err != nil {
		t.Fatal(err)
	}
	seen := make(map[int]struct{}, len(result.Vertices))
	for i, from := range result.Vertices {
		if from < 0 || from >= vertexCount {
			t.Errorf("Girth vertex %d is out of range [0, %d)", from, vertexCount)
		}
		if _, duplicate := seen[from]; duplicate {
			t.Errorf("Girth witness repeats vertex %d: %v", from, result.Vertices)
		}
		seen[from] = struct{}{}
		to := result.Vertices[(i+1)%len(result.Vertices)]
		adjacent := false
		for _, edge := range edges {
			if (edge.From == from && edge.To == to) || (edge.From == to && edge.To == from) {
				adjacent = true
				break
			}
		}
		if !adjacent {
			t.Errorf("Girth witness step %d -> %d has no edge: %v", from, to, result.Vertices)
		}
	}
}

package igraph

import (
	"errors"
	"reflect"
	"sort"
	"testing"
)

func TestUndirectedTopologyQueries(t *testing.T) {
	g, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = g.Close() })
	if err := g.AddVertices(3); err != nil {
		t.Fatal(err)
	}
	wantEdges := []Edge{{0, 1}, {0, 1}, {0, 0}, {1, 2}}
	if err := g.AddEdges(wantEdges); err != nil {
		t.Fatal(err)
	}

	for _, mode := range []DirectionMode{DirectionOut, DirectionIn, DirectionAll} {
		got, err := g.Neighbors(0, mode)
		if err != nil {
			t.Fatalf("Neighbors: %v", err)
		}
		assertIntMultiset(t, got, []int{0, 1, 1}, "undirected neighbors")
		incident, err := g.IncidentEdges(0, mode)
		if err != nil {
			t.Fatalf("IncidentEdges: %v", err)
		}
		assertIntMultiset(t, incident, []int{0, 1, 2}, "undirected incident edges")
	}
	if got, err := g.AreAdjacent(1, 0); err != nil || !got {
		t.Errorf("AreAdjacent(1,0) = %v, %v", got, err)
	}
	if got, err := g.AreAdjacent(0, 2); err != nil || got {
		t.Errorf("AreAdjacent(0,2) = %v, %v", got, err)
	}
	if id, found, err := g.EdgeID(1, 0, true); err != nil || !found || (id != 0 && id != 1) {
		t.Errorf("EdgeID = %d, %v, %v", id, found, err)
	}
	gotEdges, err := g.Edges()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotEdges, wantEdges) {
		t.Errorf("Edges = %#v, want %#v", gotEdges, wantEdges)
	}
	// Returned slices own their storage and survive subsequent C-backed calls.
	gotEdges[0] = Edge{2, 2}
	again, _ := g.Edges()
	if !reflect.DeepEqual(again, wantEdges) {
		t.Errorf("Edges after mutation = %#v", again)
	}
}

func TestDirectedTopologyQueries(t *testing.T) {
	g, err := NewLattice([]int{3}, 1, true, false, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = g.Close() })
	if err := g.AddEdges([]Edge{{0, 1}, {0, 0}}); err != nil {
		t.Fatal(err)
	}

	out, err := g.Neighbors(0, DirectionOut)
	if err != nil {
		t.Fatal(err)
	}
	assertIntMultiset(t, out, []int{0, 1, 1}, "out-neighbors")
	in, err := g.Neighbors(0, DirectionIn)
	if err != nil {
		t.Fatal(err)
	}
	assertIntMultiset(t, in, []int{0}, "in-neighbors")
	all, err := g.Neighbors(0, DirectionAll)
	if err != nil {
		t.Fatal(err)
	}
	assertIntMultiset(t, all, []int{0, 1, 1}, "all-neighbors")

	outEdges, err := g.IncidentEdges(0, DirectionOut)
	if err != nil {
		t.Fatal(err)
	}
	assertIntMultiset(t, outEdges, []int{0, 2, 3}, "out-edges")
	inEdges, err := g.IncidentEdges(0, DirectionIn)
	if err != nil {
		t.Fatal(err)
	}
	assertIntMultiset(t, inEdges, []int{3}, "in-edges")
	allEdges, err := g.IncidentEdges(0, DirectionAll)
	if err != nil {
		t.Fatal(err)
	}
	assertIntMultiset(t, allEdges, []int{0, 2, 3}, "all-edges")

	if got, err := g.AreAdjacent(0, 1); err != nil || !got {
		t.Errorf("AreAdjacent(0,1) = %v, %v", got, err)
	}
	if got, err := g.AreAdjacent(1, 0); err != nil || got {
		t.Errorf("AreAdjacent(1,0) = %v, %v", got, err)
	}
	if _, found, err := g.EdgeID(1, 0, true); err != nil || found {
		t.Errorf("directed reverse EdgeID found=%v err=%v", found, err)
	}
	if _, found, err := g.EdgeID(1, 0, false); err != nil || !found {
		t.Errorf("undirected reverse EdgeID found=%v err=%v", found, err)
	}
}

func TestEmptyTopologyQueries(t *testing.T) {
	g, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = g.Close() })
	if got, err := g.Edges(); err != nil || len(got) != 0 {
		t.Errorf("Edges = %#v, %v", got, err)
	}
	if _, err := g.Neighbors(0, DirectionAll); err == nil {
		t.Error("Neighbors invalid ID error = nil")
	}
	if _, err := g.IncidentEdges(0, DirectionAll); err == nil {
		t.Error("IncidentEdges invalid ID error = nil")
	}
	if _, err := g.AreAdjacent(0, 0); err == nil {
		t.Error("AreAdjacent invalid ID error = nil")
	}
	if _, _, err := g.EdgeID(0, 0, true); err == nil {
		t.Error("EdgeID invalid ID error = nil")
	}
}

func TestTopologyQueriesRejectInvalidInput(t *testing.T) {
	g, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	if err := g.AddVertices(2); err != nil {
		t.Fatal(err)
	}
	for _, vertex := range []int{-1, 2} {
		if _, err := g.Neighbors(vertex, DirectionAll); err == nil {
			t.Errorf("Neighbors(%d) error = nil", vertex)
		}
		if _, err := g.IncidentEdges(vertex, DirectionAll); err == nil {
			t.Errorf("IncidentEdges(%d) error = nil", vertex)
		}
	}
	if _, err := g.Neighbors(0, DirectionMode(99)); err == nil {
		t.Error("Neighbors invalid mode error = nil")
	}
	if _, err := g.IncidentEdges(0, DirectionMode(99)); err == nil {
		t.Error("IncidentEdges invalid mode error = nil")
	}
	if _, err := g.AreAdjacent(-1, 0); err == nil {
		t.Error("AreAdjacent invalid source error = nil")
	}
	if _, err := g.AreAdjacent(0, 2); err == nil {
		t.Error("AreAdjacent invalid target error = nil")
	}
	if _, _, err := g.EdgeID(-1, 0, true); err == nil {
		t.Error("EdgeID invalid source error = nil")
	}
	if _, _, err := g.EdgeID(0, 2, true); err == nil {
		t.Error("EdgeID invalid target error = nil")
	}
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	assertTopologyClosed(t, g)
	var nilGraph *Graph
	assertTopologyClosed(t, nilGraph)
}

func assertTopologyClosed(t *testing.T, g *Graph) {
	t.Helper()
	if _, err := g.Neighbors(0, DirectionAll); !errors.Is(err, ErrClosed) {
		t.Errorf("Neighbors error = %v", err)
	}
	if _, err := g.IncidentEdges(0, DirectionAll); !errors.Is(err, ErrClosed) {
		t.Errorf("IncidentEdges error = %v", err)
	}
	if _, err := g.AreAdjacent(0, 0); !errors.Is(err, ErrClosed) {
		t.Errorf("AreAdjacent error = %v", err)
	}
	if _, _, err := g.EdgeID(0, 0, true); !errors.Is(err, ErrClosed) {
		t.Errorf("EdgeID error = %v", err)
	}
	if _, err := g.Edges(); !errors.Is(err, ErrClosed) {
		t.Errorf("Edges error = %v", err)
	}
}

func assertIntMultiset(t *testing.T, got, want []int, label string) {
	t.Helper()
	sort.Ints(got)
	sort.Ints(want)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s = %v, want %v", label, got, want)
	}
}

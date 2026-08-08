package igraph

import (
	"errors"
	"reflect"
	"testing"
)

func TestEdgeSelectorsMaterialize(t *testing.T) {
	graph, err := NewGraphFromEdges(3, []Edge{{0, 1}, {0, 1}, {1, 0}, {0, 0}}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })

	explicit, err := EdgeIDs(3, 1, 3)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name     string
		selector EdgeSelector
		want     []int
	}{
		{name: "zero value is all", selector: EdgeSelector{}, want: []int{0, 1, 2, 3}},
		{name: "all", selector: AllEdges(), want: []int{0, 1, 2, 3}},
		{name: "none", selector: NoEdges(), want: []int{}},
		{name: "IDs preserve order and duplicates", selector: explicit, want: []int{3, 1, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := graph.edgeIDs(tt.selector)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("edgeIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEdgePairsDirectedLoopsParallelAndDuplicates(t *testing.T) {
	graph, err := NewGraphFromEdges(2, []Edge{{0, 1}, {0, 1}, {1, 0}, {0, 0}}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	selector, err := EdgePairs([]Edge{{0, 1}, {1, 0}, {0, 0}, {0, 1}}, true)
	if err != nil {
		t.Fatal(err)
	}
	got, err := graph.edgeIDs(selector)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 || (got[0] != 0 && got[0] != 1) || got[1] != 2 || got[2] != 3 || got[3] != got[0] {
		t.Errorf("edgeIDs(pairs) = %v, want [parallel-edge 2 3 same-parallel-edge]", got)
	}
}

func TestEdgePairsDirectedFlagAndUndirectedGraphs(t *testing.T) {
	directed, err := NewGraphFromEdges(2, []Edge{{0, 1}}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = directed.Close() })
	reverseDirected, _ := EdgePairs([]Edge{{1, 0}}, true)
	if _, err := directed.edgeIDs(reverseDirected); err == nil {
		t.Error("directed reverse pair error = nil")
	}
	reverseIgnoringDirection, _ := EdgePairs([]Edge{{1, 0}}, false)
	if got, err := directed.edgeIDs(reverseIgnoringDirection); err != nil || !reflect.DeepEqual(got, []int{0}) {
		t.Errorf("direction-ignoring pair = %v, %v", got, err)
	}

	undirected, err := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = undirected.Close() })
	if got, err := undirected.edgeIDs(reverseDirected); err != nil || !reflect.DeepEqual(got, []int{0}) {
		t.Errorf("undirected reverse pair = %v, %v", got, err)
	}
}

func TestEdgeSelectorCopiesInputAndIsReusable(t *testing.T) {
	ids := []int{1, 0}
	selector, err := EdgeIDs(ids...)
	if err != nil {
		t.Fatal(err)
	}
	ids[0] = 0
	pairs := []Edge{{0, 1}}
	pairSelector, err := EdgePairs(pairs, true)
	if err != nil {
		t.Fatal(err)
	}
	pairs[0] = Edge{1, 0}

	graph, err := NewGraphFromEdges(2, []Edge{{0, 1}, {1, 0}}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	for call := 0; call < 2; call++ {
		if got, err := graph.edgeIDs(selector); err != nil || !reflect.DeepEqual(got, []int{1, 0}) {
			t.Errorf("call %d explicit edgeIDs() = %v, %v", call, got, err)
		}
		if got, err := graph.edgeIDs(pairSelector); err != nil || !reflect.DeepEqual(got, []int{0}) {
			t.Errorf("call %d pair edgeIDs() = %v, %v", call, got, err)
		}
	}
}

func TestEdgeSelectorsOnEmptyGraph(t *testing.T) {
	graph, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	emptyIDs, err := EdgeIDs()
	if err != nil {
		t.Fatal(err)
	}
	emptyPairs, err := EdgePairs(nil, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, selector := range []EdgeSelector{AllEdges(), NoEdges(), emptyIDs, emptyPairs} {
		got, err := graph.edgeIDs(selector)
		if err != nil || got == nil || len(got) != 0 {
			t.Errorf("edgeIDs() = %#v, %v, want non-nil empty slice", got, err)
		}
	}
}

func TestEdgeSelectorRejectsInvalidConstruction(t *testing.T) {
	if selector, err := EdgeIDs(0, -1); err == nil || selector.kind != edgeSelectorAll {
		t.Errorf("EdgeIDs() = %#v, %v, want zero selector and error", selector, err)
	}
	for _, pair := range []Edge{{-1, 0}, {0, -1}} {
		if selector, err := EdgePairs([]Edge{pair}, true); err == nil || selector.kind != edgeSelectorAll {
			t.Errorf("EdgePairs(%v) = %#v, %v, want zero selector and error", pair, selector, err)
		}
	}
}

func TestEdgeSelectorValidatesGraphAtUse(t *testing.T) {
	graph, err := NewGraphFromEdges(2, []Edge{{0, 1}}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	invalidID, _ := EdgeIDs(1)
	if _, err := graph.edgeIDs(invalidID); err == nil {
		t.Error("out-of-range edge ID error = nil")
	}
	invalidEndpoint, _ := EdgePairs([]Edge{{0, 2}}, true)
	if _, err := graph.edgeIDs(invalidEndpoint); err == nil {
		t.Error("out-of-range endpoint error = nil")
	}
	missingPair, _ := EdgePairs([]Edge{{1, 0}}, true)
	if _, err := graph.edgeIDs(missingPair); err == nil {
		t.Error("missing pair error = nil")
	}
}

func TestEdgeSelectorRejectsClosedNilAndInvalidGraphs(t *testing.T) {
	graph, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := graph.edgeIDs(AllEdges()); !errors.Is(err, ErrClosed) {
		t.Errorf("closed graph error = %v, want %v", err, ErrClosed)
	}
	var nilGraph *Graph
	if _, err := nilGraph.edgeIDs(AllEdges()); !errors.Is(err, ErrClosed) {
		t.Errorf("nil graph error = %v, want %v", err, ErrClosed)
	}

	invalid := EdgeSelector{kind: edgeSelectorKind(255)}
	if cSelector, err := newCEdgeSelector(invalid); cSelector != nil || err == nil {
		t.Errorf("newCEdgeSelector() = %v, %v, want nil, error", cSelector, err)
	}
	open, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = open.Close() })
	if _, err := open.edgeIDs(invalid); err == nil {
		t.Error("edgeIDs(invalid kind) error = nil")
	}
}

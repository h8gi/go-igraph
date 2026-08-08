package igraph

import (
	"errors"
	"reflect"
	"testing"
)

func TestVertexSelectorsMaterialize(t *testing.T) {
	graph, err := NewGraphFromEdges(5, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })

	explicit, err := VertexIDs(3, 1, 3)
	if err != nil {
		t.Fatal(err)
	}
	rangeSelector, err := VertexRange(1, 4)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name     string
		selector VertexSelector
		want     []int
	}{
		{name: "zero value is all", selector: VertexSelector{}, want: []int{0, 1, 2, 3, 4}},
		{name: "all", selector: AllVertices(), want: []int{0, 1, 2, 3, 4}},
		{name: "none", selector: NoVertices(), want: []int{}},
		{name: "IDs preserve order and duplicates", selector: explicit, want: []int{3, 1, 3}},
		{name: "range", selector: rangeSelector, want: []int{1, 2, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := graph.vertexIDs(tt.selector)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("vertexIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVertexSelectorCopiesIDsAndIsReusable(t *testing.T) {
	ids := []int{2, 0}
	selector, err := VertexIDs(ids...)
	if err != nil {
		t.Fatal(err)
	}
	ids[0] = 1

	graph, err := NewGraphFromEdges(3, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	for call := 0; call < 2; call++ {
		got, err := graph.vertexIDs(selector)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, []int{2, 0}) {
			t.Errorf("call %d vertexIDs() = %v", call, got)
		}
	}
}

func TestVertexSelectorsOnEmptyGraph(t *testing.T) {
	graph, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	emptyRange, err := VertexRange(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, selector := range []VertexSelector{AllVertices(), NoVertices(), emptyRange} {
		got, err := graph.vertexIDs(selector)
		if err != nil || got == nil || len(got) != 0 {
			t.Errorf("vertexIDs() = %#v, %v, want non-nil empty slice", got, err)
		}
	}
}

func TestVertexSelectorRejectsInvalidConstruction(t *testing.T) {
	if selector, err := VertexIDs(0, -1); err == nil || selector.kind != vertexSelectorAll {
		t.Errorf("VertexIDs() = %#v, %v, want zero selector and error", selector, err)
	}
	for _, bounds := range [][2]int{{-1, 0}, {2, 1}} {
		if selector, err := VertexRange(bounds[0], bounds[1]); err == nil || selector.kind != vertexSelectorAll {
			t.Errorf("VertexRange(%d, %d) = %#v, %v, want zero selector and error", bounds[0], bounds[1], selector, err)
		}
	}
}

func TestVertexSelectorValidatesGraphBoundsAtUse(t *testing.T) {
	graph, err := NewGraphFromEdges(3, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })

	ids, err := VertexIDs(0, 3)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := graph.vertexIDs(ids); err == nil {
		t.Error("out-of-range ID error = nil")
	}
	rangeSelector, err := VertexRange(2, 4)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := graph.vertexIDs(rangeSelector); err == nil {
		t.Error("out-of-range range error = nil")
	}
}

func TestVertexSelectorRejectsClosedAndNilGraphs(t *testing.T) {
	graph, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := graph.vertexIDs(AllVertices()); !errors.Is(err, ErrClosed) {
		t.Errorf("closed graph error = %v, want %v", err, ErrClosed)
	}
	var nilGraph *Graph
	if _, err := nilGraph.vertexIDs(AllVertices()); !errors.Is(err, ErrClosed) {
		t.Errorf("nil graph error = %v, want %v", err, ErrClosed)
	}
}

func TestVertexSelectorRejectsInvalidKind(t *testing.T) {
	selector := VertexSelector{kind: vertexSelectorKind(255)}
	if cSelector, err := newCVertexSelector(selector); cSelector != nil || err == nil {
		t.Errorf("newCVertexSelector() = %v, %v, want nil, error", cSelector, err)
	}
	graph, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	if _, err := graph.vertexIDs(selector); err == nil {
		t.Error("vertexIDs(invalid kind) error = nil")
	}
}

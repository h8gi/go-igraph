package igraph

import (
	"errors"
	"math"
	"sync"
	"testing"
)

func TestBurtConstraintKnownAnswersSelectorOrderAndOwnership(t *testing.T) {
	graph, err := NewPath(3, false, false)
	if err != nil {
		t.Fatal(err)
	}
	vertices, _ := VertexIDs(1, 0, 1, 2)
	values, err := graph.BurtConstraint(vertices, nil)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, values, []float64{0.5, 1, 0.5, 1})
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, values, []float64{0.5, 1, 0.5, 1})
}

func TestBurtConstraintWeightedDirectedIsolateAndMultigraph(t *testing.T) {
	t.Run("weighted", func(t *testing.T) {
		graph, err := NewPath(3, false, false)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = graph.Close() })
		values, err := graph.BurtConstraint(AllVertices(), []float64{1, 3})
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSlice(t, values, []float64{1, 0.625, 1})
	})

	t.Run("directed combines directions", func(t *testing.T) {
		graph, err := NewGraphFromEdges(3, []Edge{{0, 1}, {2, 1}}, true)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = graph.Close() })
		values, err := graph.BurtConstraint(AllVertices(), nil)
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSlice(t, values, []float64{1, 0.5, 1})
	})

	t.Run("isolate and zero strength", func(t *testing.T) {
		graph, err := NewGraphFromEdges(3, []Edge{{0, 1}}, false)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = graph.Close() })
		values, err := graph.BurtConstraint(AllVertices(), []float64{0})
		if err != nil {
			t.Fatal(err)
		}
		if !math.IsNaN(values[0]) || !math.IsNaN(values[1]) || !math.IsNaN(values[2]) {
			t.Errorf("zero-strength and isolate values = %#v, want all NaN", values)
		}
	})

	t.Run("parallel edges and loop", func(t *testing.T) {
		graph, err := NewGraphFromEdges(3, []Edge{{0, 1}, {0, 1}, {1, 2}, {1, 1}}, false)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = graph.Close() })
		values, err := graph.BurtConstraint(AllVertices(), []float64{1, 2, 3, 100})
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSlice(t, values, []float64{1, 0.5, 1})
	})
}

func TestBurtConstraintEmptyInvalidAndClosed(t *testing.T) {
	empty, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	values, err := empty.BurtConstraint(AllVertices(), nil)
	if err != nil || values == nil || len(values) != 0 {
		t.Fatalf("empty graph result = %#v, %v", values, err)
	}
	if err := empty.Close(); err != nil {
		t.Fatal(err)
	}

	graph, err := NewPath(3, false, false)
	if err != nil {
		t.Fatal(err)
	}
	emptySelection, err := graph.BurtConstraint(NoVertices(), nil)
	if err != nil || emptySelection == nil || len(emptySelection) != 0 {
		t.Fatalf("empty selection = %#v, %v", emptySelection, err)
	}
	badSelector := VertexSelector{kind: vertexSelectorIDs, ids: []int{3}}
	if _, err := graph.BurtConstraint(badSelector, nil); err == nil {
		t.Error("invalid selector error = nil")
	}
	for _, weights := range [][]float64{{}, {1}, {1, -1}, {1, math.NaN()}, {1, math.Inf(1)}} {
		if _, err := graph.BurtConstraint(AllVertices(), weights); err == nil {
			t.Errorf("invalid weights %#v error = nil", weights)
		}
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := graph.BurtConstraint(AllVertices(), nil); !errors.Is(err, ErrClosed) {
		t.Errorf("call after Close error = %v", err)
	}
	var nilGraph *Graph
	if _, err := nilGraph.BurtConstraint(AllVertices(), nil); !errors.Is(err, ErrClosed) {
		t.Errorf("nil graph error = %v", err)
	}
}

func TestBurtConstraintConcurrentReads(t *testing.T) {
	graph, err := NewPath(20, false, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	var group sync.WaitGroup
	for index := 0; index < 8; index++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if _, err := graph.BurtConstraint(AllVertices(), nil); err != nil {
				t.Errorf("concurrent call: %v", err)
			}
		}()
	}
	group.Wait()
}

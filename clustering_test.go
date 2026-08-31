package igraph

import (
	"errors"
	"math"
	"sync"
	"testing"
)

func TestBarratTransitivityWeightedSelectorUndefinedAndOwnership(t *testing.T) {
	graph, err := NewGraphFromEdges(4, []Edge{{0, 1}, {1, 2}, {2, 0}}, false)
	if err != nil {
		t.Fatal(err)
	}
	vertices, _ := VertexIDs(2, 3, 0, 2)
	values, err := graph.BarratTransitivity(vertices, []float64{1, 2, 4}, TransitivityNaN)
	if err != nil {
		t.Fatal(err)
	}
	if values[0] != 1 || !math.IsNaN(values[1]) || values[2] != 1 || values[3] != 1 {
		t.Errorf("Barrat values = %#v", values)
	}
	zero, err := graph.BarratTransitivity(vertices, []float64{1, 2, 4}, TransitivityZero)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, zero, []float64{1, 0, 1, 1})
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, zero, []float64{1, 0, 1, 1})
}

func TestBarratTransitivityDirectionWeightsSimpleGraphAndEmpty(t *testing.T) {
	var nilGraph *Graph
	if _, err := nilGraph.BarratTransitivity(AllVertices(), []float64{}, TransitivityNaN); !errors.Is(err, ErrClosed) {
		t.Errorf("nil graph Barrat error = %v", err)
	}

	t.Run("directions ignored", func(t *testing.T) {
		graph, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}, {2, 0}}, true)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = graph.Close() })
		values, err := graph.BarratTransitivity(AllVertices(), []float64{1, 1, 1}, TransitivityNaN)
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSlice(t, values, []float64{1, 1, 1})
	})

	t.Run("zero strength", func(t *testing.T) {
		graph, err := NewPath(3, false, false)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = graph.Close() })
		nanValues, err := graph.BarratTransitivity(AllVertices(), []float64{0, 0}, TransitivityNaN)
		if err != nil {
			t.Fatal(err)
		}
		for index, value := range nanValues {
			if !math.IsNaN(value) {
				t.Errorf("NaN mode value %d = %v", index, value)
			}
		}
		zeroValues, err := graph.BarratTransitivity(AllVertices(), []float64{0, 0}, TransitivityZero)
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSlice(t, zeroValues, []float64{0, 0, 0})
	})

	t.Run("invalid weights and graph shape", func(t *testing.T) {
		graph, err := NewPath(3, false, false)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = graph.Close() })
		for _, weights := range [][]float64{nil, {}, {1}, {1, math.NaN()}, {1, math.Inf(1)}} {
			if _, err := graph.BarratTransitivity(AllVertices(), weights, TransitivityNaN); err == nil {
				t.Errorf("invalid weights %#v error = nil", weights)
			}
		}
		badSelector := VertexSelector{kind: vertexSelectorIDs, ids: []int{3}}
		if _, err := graph.BarratTransitivity(badSelector, []float64{1, 1}, TransitivityNaN); err == nil {
			t.Error("invalid selector error = nil")
		}
		if _, err := graph.BarratTransitivity(AllVertices(), []float64{1, 1}, TransitivityMode(99)); err == nil {
			t.Error("invalid mode error = nil")
		}

		shapes := []struct {
			name     string
			edges    []Edge
			directed bool
		}{
			{name: "loop", edges: []Edge{{0, 0}}},
			{name: "parallel", edges: []Edge{{0, 1}, {0, 1}}},
			{name: "mutual", edges: []Edge{{0, 1}, {1, 0}}, directed: true},
		}
		for _, shape := range shapes {
			t.Run(shape.name, func(t *testing.T) {
				bad, err := NewGraphFromEdges(2, shape.edges, shape.directed)
				if err != nil {
					t.Fatal(err)
				}
				defer bad.Close()
				weights := make([]float64, len(shape.edges))
				for index := range weights {
					weights[index] = 1
				}
				if _, err := bad.BarratTransitivity(AllVertices(), weights, TransitivityNaN); err == nil {
					t.Error("non-simple graph error = nil")
				}
			})
		}
	})

	t.Run("empty", func(t *testing.T) {
		empty, err := NewGraph()
		if err != nil {
			t.Fatal(err)
		}
		defer empty.Close()
		values, err := empty.BarratTransitivity(AllVertices(), []float64{}, TransitivityNaN)
		if err != nil || values == nil || len(values) != 0 {
			t.Errorf("empty result = %#v, %v", values, err)
		}
	})
}

func TestEdgeClusteringKnownAnswersOptionsSelectorAndOwnership(t *testing.T) {
	triangle, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}, {2, 0}}, false)
	if err != nil {
		t.Fatal(err)
	}
	edges, _ := EdgeIDs(2, 0, 2, 1)
	tests := []struct {
		options EdgeClusteringOptions
		want    float64
	}{
		{options: EdgeClusteringOptions{CycleSize: 3}, want: 1},
		{options: EdgeClusteringOptions{CycleSize: 3, Offset: true}, want: 2},
		{options: EdgeClusteringOptions{CycleSize: 3, Normalize: true}, want: 1},
		{options: EdgeClusteringOptions{CycleSize: 3, Offset: true, Normalize: true}, want: 2},
	}
	for _, test := range tests {
		values, err := triangle.EdgeClustering(edges, test.options)
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSlice(t, values, []float64{test.want, test.want, test.want, test.want})
	}
	owned, err := triangle.EdgeClustering(AllEdges(), EdgeClusteringOptions{CycleSize: 3})
	if err != nil {
		t.Fatal(err)
	}
	if err := triangle.Close(); err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, owned, []float64{1, 1, 1})

	square, err := NewGraphFromEdges(4, []Edge{{0, 1}, {1, 2}, {2, 3}, {3, 0}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer square.Close()
	values, err := square.EdgeClustering(AllEdges(), EdgeClusteringOptions{CycleSize: 4, Normalize: true})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, values, []float64{1, 1, 1, 1})
}

func TestEdgeClusteringDirectedMultiplicityLoopsEmptyInvalidAndConcurrent(t *testing.T) {
	var nilGraph *Graph
	if _, err := nilGraph.EdgeClustering(AllEdges(), EdgeClusteringOptions{CycleSize: 3}); !errors.Is(err, ErrClosed) {
		t.Errorf("nil graph edge clustering error = %v", err)
	}

	directed, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}, {2, 0}}, true)
	if err != nil {
		t.Fatal(err)
	}
	directedValues, err := directed.EdgeClustering(AllEdges(), EdgeClusteringOptions{CycleSize: 3, Normalize: true})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, directedValues, []float64{1, 1, 1})
	_ = directed.Close()

	multi, err := NewGraphFromEdges(3, []Edge{{0, 1}, {0, 1}, {1, 2}, {2, 0}, {0, 0}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer multi.Close()
	values, err := multi.EdgeClustering(AllEdges(), EdgeClusteringOptions{CycleSize: 3, Normalize: true})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, values[:2], []float64{0.5, 0.5})
	if !math.IsNaN(values[4]) {
		t.Errorf("normalized loop = %v, want NaN", values[4])
	}
	loopRaw, err := multi.EdgeClustering(EdgeSelector{kind: edgeSelectorIDs, ids: []int{4}}, EdgeClusteringOptions{CycleSize: 3, Offset: true})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, loopRaw, []float64{1})

	empty, err := multi.EdgeClustering(NoEdges(), EdgeClusteringOptions{CycleSize: 3})
	if err != nil || empty == nil || len(empty) != 0 {
		t.Errorf("empty selection = %#v, %v", empty, err)
	}
	for _, size := range []int{-1, 0, 2, 5} {
		if _, err := multi.EdgeClustering(AllEdges(), EdgeClusteringOptions{CycleSize: size}); err == nil {
			t.Errorf("cycle size %d error = nil", size)
		}
	}
	badEdge := EdgeSelector{kind: edgeSelectorIDs, ids: []int{5}}
	if _, err := multi.EdgeClustering(badEdge, EdgeClusteringOptions{CycleSize: 3}); err == nil {
		t.Error("invalid selector error = nil")
	}
	missingPair, _ := EdgePairs([]Edge{{1, 1}}, false)
	if _, err := multi.EdgeClustering(missingPair, EdgeClusteringOptions{CycleSize: 3}); err == nil {
		t.Error("missing edge pair error = nil")
	}

	graph, err := NewPath(20, false, false)
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for index := 0; index < 8; index++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if _, err := graph.EdgeClustering(AllEdges(), EdgeClusteringOptions{CycleSize: 3}); err != nil {
				t.Errorf("concurrent edge clustering: %v", err)
			}
			if _, err := graph.BarratTransitivity(AllVertices(), makeOnes(19), TransitivityNaN); err != nil {
				t.Errorf("concurrent Barrat transitivity: %v", err)
			}
		}()
	}
	group.Wait()
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := graph.EdgeClustering(AllEdges(), EdgeClusteringOptions{CycleSize: 3}); !errors.Is(err, ErrClosed) {
		t.Errorf("edge clustering after Close = %v", err)
	}
	if _, err := graph.BarratTransitivity(AllVertices(), makeOnes(19), TransitivityNaN); !errors.Is(err, ErrClosed) {
		t.Errorf("Barrat after Close = %v", err)
	}
}

func makeOnes(length int) []float64 {
	values := make([]float64, length)
	for index := range values {
		values[index] = 1
	}
	return values
}

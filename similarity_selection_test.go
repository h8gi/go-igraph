package igraph

import (
	"errors"
	"reflect"
	"testing"
)

func TestNeighborhoodSimilarityPairsMatchMatrix(t *testing.T) {
	graph := newSimilarityTestGraph(t)
	defer graph.Close()

	pairs := []Edge{
		{From: 4, To: 0},
		{From: 0, To: 1},
		{From: 4, To: 0},
		{From: 2, To: 2},
	}
	for _, metric := range []NeighborhoodSimilarityMetric{SimilarityJaccard, SimilarityDice} {
		options := NeighborhoodSimilarityOptions{Metric: metric, Direction: DirectionOut}
		got, err := graph.NeighborhoodSimilarityPairs(pairs, options)
		if err != nil {
			t.Fatalf("NeighborhoodSimilarityPairs(%v) error = %v", metric, err)
		}
		want := similarityPairMatrixValues(t, graph, pairs, options)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("NeighborhoodSimilarityPairs(%v) = %v, want %v", metric, got, want)
		}
	}

	loopPair := []Edge{{From: 2, To: 0}}
	withoutLoops, err := graph.NeighborhoodSimilarityPairs(loopPair, NeighborhoodSimilarityOptions{})
	if err != nil {
		t.Fatal(err)
	}
	withLoops, err := graph.NeighborhoodSimilarityPairs(loopPair, NeighborhoodSimilarityOptions{IncludeLoops: true})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(withoutLoops, []float64{0}) || !reflect.DeepEqual(withLoops, []float64{1.0 / 3.0}) {
		t.Errorf("pair loop results = %v and %v", withoutLoops, withLoops)
	}

	empty, err := graph.NeighborhoodSimilarityPairs(nil, NeighborhoodSimilarityOptions{})
	if err != nil {
		t.Fatalf("NeighborhoodSimilarityPairs(empty) error = %v", err)
	}
	if empty == nil || len(empty) != 0 {
		t.Fatalf("NeighborhoodSimilarityPairs(empty) = %#v, want non-nil empty", empty)
	}
}

func TestNeighborhoodSimilarityEdgesMatchMatrix(t *testing.T) {
	graph := newSimilarityTestGraph(t)
	defer graph.Close()

	selectedIDs := []int{4, 0, 4}
	selector, err := EdgeIDs(selectedIDs...)
	if err != nil {
		t.Fatal(err)
	}
	options := NeighborhoodSimilarityOptions{
		Metric: SimilarityDice, Direction: DirectionOut, IncludeLoops: true,
	}
	got, err := graph.NeighborhoodSimilarityEdges(selector, options)
	if err != nil {
		t.Fatalf("NeighborhoodSimilarityEdges() error = %v", err)
	}
	pairs := make([]Edge, len(selectedIDs))
	for index, edgeID := range selectedIDs {
		pairs[index].From, pairs[index].To, err = graph.EdgeEndpoints(edgeID)
		if err != nil {
			t.Fatal(err)
		}
	}
	want := similarityPairMatrixValues(t, graph, pairs, options)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("NeighborhoodSimilarityEdges() = %v, want %v", got, want)
	}

	pairSelector, err := EdgePairs([]Edge{{From: 0, To: 2}}, true)
	if err != nil {
		t.Fatal(err)
	}
	fromPair, err := graph.NeighborhoodSimilarityEdges(pairSelector, options)
	if err != nil {
		t.Fatalf("NeighborhoodSimilarityEdges(pair selector) error = %v", err)
	}
	if !reflect.DeepEqual(fromPair, want[1:2]) {
		t.Errorf("NeighborhoodSimilarityEdges(pair selector) = %v, want %v", fromPair, want[1:2])
	}

	empty, err := graph.NeighborhoodSimilarityEdges(NoEdges(), NeighborhoodSimilarityOptions{})
	if err != nil {
		t.Fatalf("NeighborhoodSimilarityEdges(empty) error = %v", err)
	}
	if empty == nil || len(empty) != 0 {
		t.Fatalf("NeighborhoodSimilarityEdges(empty) = %#v, want non-nil empty", empty)
	}
}

func TestSelectedSimilarityValidationOwnershipAndFailures(t *testing.T) {
	graph := newSimilarityTestGraph(t)

	for _, pairs := range [][]Edge{
		{{From: -1, To: 0}},
		{{From: 0, To: 5}},
	} {
		if _, err := graph.NeighborhoodSimilarityPairs(pairs, NeighborhoodSimilarityOptions{}); err == nil {
			t.Errorf("NeighborhoodSimilarityPairs(%v) error = nil", pairs)
		}
	}
	if _, err := graph.NeighborhoodSimilarityPairs(nil, NeighborhoodSimilarityOptions{Metric: NeighborhoodSimilarityMetric(99)}); err == nil {
		t.Error("NeighborhoodSimilarityPairs(invalid metric) error = nil")
	}
	if _, err := graph.NeighborhoodSimilarityPairs(nil, NeighborhoodSimilarityOptions{Direction: DirectionMode(99)}); err == nil {
		t.Error("NeighborhoodSimilarityPairs(invalid direction) error = nil")
	}
	invalidEdge, _ := EdgeIDs(5)
	invalidKind := EdgeSelector{kind: edgeSelectorKind(99)}
	missingPair, _ := EdgePairs([]Edge{{From: 2, To: 4}}, true)
	for _, selector := range []EdgeSelector{invalidEdge, invalidKind, missingPair} {
		if _, err := graph.NeighborhoodSimilarityEdges(selector, NeighborhoodSimilarityOptions{}); err == nil {
			t.Errorf("NeighborhoodSimilarityEdges(%v) error = nil", selector)
		}
	}
	if _, err := graph.NeighborhoodSimilarityEdges(NoEdges(), NeighborhoodSimilarityOptions{Metric: NeighborhoodSimilarityMetric(99)}); err == nil {
		t.Error("NeighborhoodSimilarityEdges(invalid metric) error = nil")
	}

	initializationError := errors.New("initialize selected similarity result")
	initializationHooks := similarityVectorHooks{
		newResult: func() (*realVector, error) { return nil, initializationError },
	}
	if _, err := graph.neighborhoodSimilarityPairs(nil, NeighborhoodSimilarityOptions{}, initializationHooks); !errors.Is(err, initializationError) {
		t.Errorf("pair initialization error = %v, want %v", err, initializationError)
	}
	if _, err := graph.neighborhoodSimilarityEdges(NoEdges(), NeighborhoodSimilarityOptions{}, initializationHooks); !errors.Is(err, initializationError) {
		t.Errorf("edge initialization error = %v, want %v", err, initializationError)
	}
	operationError := errors.New("selected similarity operation failed")
	operationHooks := similarityVectorHooks{run: func() error { return operationError }}
	if _, err := graph.neighborhoodSimilarityPairs(nil, NeighborhoodSimilarityOptions{}, operationHooks); !errors.Is(err, operationError) {
		t.Errorf("pair operation error = %v, want %v", err, operationError)
	}
	if _, err := graph.neighborhoodSimilarityEdges(NoEdges(), NeighborhoodSimilarityOptions{}, operationHooks); !errors.Is(err, operationError) {
		t.Errorf("edge operation error = %v, want %v", err, operationError)
	}

	result, err := graph.NeighborhoodSimilarityPairs([]Edge{{From: 0, To: 1}}, NeighborhoodSimilarityOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result, []float64{1}) {
		t.Errorf("pair result after Close = %v", result)
	}
	if _, err := graph.NeighborhoodSimilarityPairs(nil, NeighborhoodSimilarityOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("closed pair graph error = %v, want %v", err, ErrClosed)
	}
	if _, err := graph.NeighborhoodSimilarityEdges(NoEdges(), NeighborhoodSimilarityOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("closed edge graph error = %v, want %v", err, ErrClosed)
	}
	var nilGraph *Graph
	if _, err := nilGraph.NeighborhoodSimilarityPairs(nil, NeighborhoodSimilarityOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("nil pair graph error = %v, want %v", err, ErrClosed)
	}
	if _, err := nilGraph.NeighborhoodSimilarityEdges(NoEdges(), NeighborhoodSimilarityOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("nil edge graph error = %v, want %v", err, ErrClosed)
	}

	wrongSize, err := newRealVectorSize(1)
	if err != nil {
		t.Fatal(err)
	}
	defer wrongSize.close()
	if _, err := checkedSimilarityValues(wrongSize, 2); err == nil {
		t.Error("checkedSimilarityValues(mismatched length) error = nil")
	}
}

func similarityPairMatrixValues(
	t *testing.T,
	graph *Graph,
	pairs []Edge,
	options NeighborhoodSimilarityOptions,
) []float64 {
	t.Helper()
	result := make([]float64, len(pairs))
	for index, pair := range pairs {
		rows, err := VertexIDs(pair.From)
		if err != nil {
			t.Fatal(err)
		}
		columns, err := VertexIDs(pair.To)
		if err != nil {
			t.Fatal(err)
		}
		matrix, err := graph.NeighborhoodSimilarity(rows, columns, options)
		if err != nil {
			t.Fatal(err)
		}
		result[index], err = matrix.At(0, 0)
		if err != nil {
			t.Fatal(err)
		}
	}
	return result
}

package igraph

import (
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestDistanceCentralitiesKnownAnswersAndSelectorOrder(t *testing.T) {
	graph, err := NewPath(4, false, false)
	if err != nil {
		t.Fatal(err)
	}
	selector, err := VertexIDs(2, 0, 2, 1)
	if err != nil {
		t.Fatal(err)
	}

	closeness, err := graph.Closeness(selector, DistanceCentralityOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, closeness.Scores, []float64{1.0 / 4, 1.0 / 6, 1.0 / 4, 1.0 / 4})
	if want := []int{3, 3, 3, 3}; !reflect.DeepEqual(closeness.ReachableCounts, want) {
		t.Errorf("ReachableCounts = %v, want %v", closeness.ReachableCounts, want)
	}
	if !closeness.AllReachable {
		t.Error("AllReachable = false, want true")
	}

	normalized, err := graph.Closeness(selector, DistanceCentralityOptions{Normalized: true})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, normalized.Scores, []float64{3.0 / 4, 1.0 / 2, 3.0 / 4, 3.0 / 4})

	harmonic, err := graph.HarmonicCentrality(selector, DistanceCentralityOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, harmonic, []float64{2.5, 11.0 / 6, 2.5, 2.5})
	harmonicNormalized, err := graph.HarmonicCentrality(selector, DistanceCentralityOptions{Normalized: true})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, harmonicNormalized, []float64{5.0 / 6, 11.0 / 18, 5.0 / 6, 5.0 / 6})

	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, closeness.Scores, []float64{1.0 / 4, 1.0 / 6, 1.0 / 4, 1.0 / 4})
	if !reflect.DeepEqual(closeness.ReachableCounts, []int{3, 3, 3, 3}) {
		t.Errorf("Go-owned reachability changed after Close: %v", closeness.ReachableCounts)
	}
	assertFloatSlice(t, harmonic, []float64{2.5, 11.0 / 6, 2.5, 2.5})
}

func TestDistanceCentralitiesDirectionWeightsAndCutoff(t *testing.T) {
	graph, err := NewPath(4, true, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	selector, _ := VertexIDs(0, 2, 3)

	out, err := graph.Closeness(selector, DistanceCentralityOptions{Direction: DirectionOut})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, out.Scores, []float64{1.0 / 6, 1, math.NaN()})
	if want := []int{3, 1, 0}; !reflect.DeepEqual(out.ReachableCounts, want) {
		t.Errorf("out reachable = %v, want %v", out.ReachableCounts, want)
	}
	if out.AllReachable {
		t.Error("directed path unexpectedly reports all vertices reachable")
	}
	in, err := graph.Closeness(selector, DistanceCentralityOptions{Direction: DirectionIn})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, in.Scores, []float64{math.NaN(), 1.0 / 3, 1.0 / 6})
	all, err := graph.Closeness(selector, DistanceCentralityOptions{Direction: DirectionAll})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, all.Scores, []float64{1.0 / 6, 1.0 / 4, 1.0 / 6})

	weighted, err := graph.Closeness(selector, DistanceCentralityOptions{
		Direction: DirectionOut,
		Weights:   []float64{2, 3, 4},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, weighted.Scores, []float64{1.0 / 16, 1.0 / 4, math.NaN()})
	weightedHarmonic, err := graph.HarmonicCentrality(selector, DistanceCentralityOptions{
		Direction: DirectionOut,
		Weights:   []float64{2, 3, 4},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, weightedHarmonic, []float64{1.0/2 + 1.0/5 + 1.0/9, 1.0 / 4, 0})

	cutoff := 1.0
	limited, err := graph.Closeness(AllVertices(), DistanceCentralityOptions{
		Direction: DirectionOut,
		Cutoff:    &cutoff,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, limited.Scores, []float64{1, 1, 1, math.NaN()})
	if want := []int{1, 1, 1, 0}; !reflect.DeepEqual(limited.ReachableCounts, want) {
		t.Errorf("cutoff reachable = %v, want %v", limited.ReachableCounts, want)
	}
	limitedHarmonic, err := graph.HarmonicCentrality(AllVertices(), DistanceCentralityOptions{
		Direction: DirectionOut,
		Cutoff:    &cutoff,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, limitedHarmonic, []float64{1, 1, 1, 0})
}

func TestDistanceCentralitiesDisconnectedEmptyAndZeroCutoff(t *testing.T) {
	graph, err := NewGraphFromEdges(4, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	result, err := graph.Closeness(AllVertices(), DistanceCentralityOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, result.Scores, []float64{1, 1, math.NaN(), math.NaN()})
	if want := []int{1, 1, 0, 0}; !reflect.DeepEqual(result.ReachableCounts, want) {
		t.Errorf("reachable = %v, want %v", result.ReachableCounts, want)
	}
	if result.AllReachable {
		t.Error("disconnected graph reports all reachable")
	}
	harmonic, err := graph.HarmonicCentrality(AllVertices(), DistanceCentralityOptions{Normalized: true})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, harmonic, []float64{1.0 / 3, 1.0 / 3, 0, 0})

	zero := 0.0
	zeroResult, err := graph.Closeness(AllVertices(), DistanceCentralityOptions{Cutoff: &zero})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, zeroResult.Scores, []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN()})
	if want := []int{0, 0, 0, 0}; !reflect.DeepEqual(zeroResult.ReachableCounts, want) {
		t.Errorf("zero-cutoff reachable = %v, want %v", zeroResult.ReachableCounts, want)
	}

	emptySelection, err := graph.Closeness(NoVertices(), DistanceCentralityOptions{})
	if err != nil || emptySelection.Scores == nil || emptySelection.ReachableCounts == nil ||
		len(emptySelection.Scores) != 0 || len(emptySelection.ReachableCounts) != 0 {
		t.Errorf("Closeness(NoVertices) = %#v, %v", emptySelection, err)
	}
	emptyHarmonic, err := graph.HarmonicCentrality(NoVertices(), DistanceCentralityOptions{})
	if err != nil || emptyHarmonic == nil || len(emptyHarmonic) != 0 {
		t.Errorf("HarmonicCentrality(NoVertices) = %#v, %v", emptyHarmonic, err)
	}

	emptyGraph, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = emptyGraph.Close() })
	emptyGraphResult, err := emptyGraph.Closeness(AllVertices(), DistanceCentralityOptions{})
	if err != nil || emptyGraphResult.Scores == nil || emptyGraphResult.ReachableCounts == nil {
		t.Errorf("empty graph closeness = %#v, %v", emptyGraphResult, err)
	}
	emptyGraphHarmonic, err := emptyGraph.HarmonicCentrality(AllVertices(), DistanceCentralityOptions{})
	if err != nil || emptyGraphHarmonic == nil {
		t.Errorf("empty graph harmonic = %#v, %v", emptyGraphHarmonic, err)
	}
}

func TestDistanceCentralitiesRejectInvalidInputsAndClosedGraph(t *testing.T) {
	graph, err := NewPath(3, false, false)
	if err != nil {
		t.Fatal(err)
	}
	badSelector := VertexSelector{kind: vertexSelectorIDs, ids: []int{3}}
	if _, err := graph.Closeness(badSelector, DistanceCentralityOptions{}); err == nil || !strings.Contains(err.Error(), "selector") {
		t.Errorf("invalid Closeness selector error = %v", err)
	}
	if _, err := graph.HarmonicCentrality(VertexSelector{kind: vertexSelectorKind(99)}, DistanceCentralityOptions{}); err == nil {
		t.Error("invalid HarmonicCentrality selector error = nil")
	}

	invalidCutoffs := []float64{-1, math.NaN(), math.Inf(1), math.Inf(-1)}
	for _, cutoff := range invalidCutoffs {
		options := DistanceCentralityOptions{Cutoff: &cutoff}
		if _, err := graph.Closeness(AllVertices(), options); err == nil {
			t.Errorf("Closeness cutoff %v error = nil", cutoff)
		}
		if _, err := graph.HarmonicCentrality(AllVertices(), options); err == nil {
			t.Errorf("HarmonicCentrality cutoff %v error = nil", cutoff)
		}
	}
	invalidOptions := []DistanceCentralityOptions{
		{Direction: DirectionMode(99)},
		{Weights: []float64{}},
		{Weights: []float64{1}},
		{Weights: []float64{1, 0}},
		{Weights: []float64{1, -1}},
		{Weights: []float64{1, math.NaN()}},
		{Weights: []float64{1, math.Inf(1)}},
	}
	for _, options := range invalidOptions {
		if _, err := graph.Closeness(AllVertices(), options); err == nil {
			t.Errorf("Closeness(%#v) error = nil", options)
		}
		if _, err := graph.HarmonicCentrality(AllVertices(), options); err == nil {
			t.Errorf("HarmonicCentrality(%#v) error = nil", options)
		}
	}

	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := graph.Closeness(AllVertices(), DistanceCentralityOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("Closeness after Close error = %v", err)
	}
	if _, err := graph.HarmonicCentrality(AllVertices(), DistanceCentralityOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("HarmonicCentrality after Close error = %v", err)
	}
	var nilGraph *Graph
	if _, err := nilGraph.Closeness(AllVertices(), DistanceCentralityOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("nil Closeness error = %v", err)
	}
	if _, err := nilGraph.HarmonicCentrality(AllVertices(), DistanceCentralityOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("nil HarmonicCentrality error = %v", err)
	}
}

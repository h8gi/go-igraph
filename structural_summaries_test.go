package igraph

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestDegreeSummariesKnownAnswers(t *testing.T) {
	graph := mustDiagnosticGraph(t, false, 4, []Edge{{From: 0, To: 1}, {From: 1, To: 2}})
	defer graph.Close()

	if got, err := graph.MeanDegree(false); err != nil || got != 1 {
		t.Fatalf("MeanDegree(false) = %v, %v, want 1, nil", got, err)
	}
	if got, err := graph.MaxDegree(AllVertices(), DegreeOptions{}); err != nil || got != 2 {
		t.Fatalf("MaxDegree() = %d, %v, want 2, nil", got, err)
	}
	if got, err := graph.MaxDegree(NoVertices(), DegreeOptions{}); err != nil || got != 0 {
		t.Fatalf("empty MaxDegree() = %d, %v, want 0, nil", got, err)
	}

	result, err := graph.AverageNearestNeighborDegree(AllVertices(), NearestNeighborDegreeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertSummaryFloatSlice(t, result.ByVertex, []float64{2, 1, 2, math.NaN()})
	assertSummaryFloatSlice(t, result.ByDegree, []float64{2, 1})

	correlation, err := graph.DegreeCorrelation(DegreeCorrelationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertSummaryFloatSlice(t, correlation, []float64{math.NaN(), 2, 1})

	descending, err := graph.VerticesByDegree(AllVertices(), DegreeOptions{}, true)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(descending, []int{1, 0, 2, 3}) {
		t.Fatalf("VerticesByDegree(descending) = %v, want [1 0 2 3]", descending)
	}
	ascending, err := graph.VerticesByDegree(AllVertices(), DegreeOptions{}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(ascending, []int{3, 0, 2, 1}) {
		t.Fatalf("VerticesByDegree(ascending) = %v, want [3 0 2 1]", ascending)
	}
}

func TestWeightedNeighborDegreeAndDiversity(t *testing.T) {
	graph := mustDiagnosticGraph(t, false, 4, []Edge{{From: 0, To: 1}, {From: 1, To: 2}})
	defer graph.Close()
	weights := []float64{1, 3}

	neighbor, err := graph.AverageNearestNeighborDegree(AllVertices(), NearestNeighborDegreeOptions{Weights: weights})
	if err != nil {
		t.Fatal(err)
	}
	assertSummaryFloatSlice(t, neighbor.ByVertex, []float64{2, 1, 2, math.NaN()})

	selected, err := VertexIDs(3, 1, 3)
	if err != nil {
		t.Fatal(err)
	}
	diversity, err := graph.Diversity(selected, weights)
	if err != nil {
		t.Fatal(err)
	}
	assertSummaryFloatSlice(t, diversity, []float64{math.NaN(), 0.8112781244591328, math.NaN()})
}

func TestReciprocityKnownAnswers(t *testing.T) {
	graph := mustDiagnosticGraph(t, true, 3, []Edge{
		{From: 0, To: 1}, {From: 1, To: 0}, {From: 1, To: 2},
	})
	defer graph.Close()
	if got, err := graph.Reciprocity(ReciprocityDefault, true); err != nil || math.Abs(got-2.0/3.0) > 1e-12 {
		t.Fatalf("default Reciprocity() = %v, %v, want 2/3, nil", got, err)
	}
	if got, err := graph.Reciprocity(ReciprocityRatio, true); err != nil || math.Abs(got-0.5) > 1e-12 {
		t.Fatalf("ratio Reciprocity() = %v, %v, want 0.5, nil", got, err)
	}
	if _, err := graph.Reciprocity(ReciprocityMode(99), true); err == nil {
		t.Fatal("Reciprocity(invalid mode) returned nil error")
	}
}

func TestRichClubSequenceAndValidation(t *testing.T) {
	graph := mustDiagnosticGraph(t, false, 4, []Edge{{From: 0, To: 1}, {From: 1, To: 2}})
	defer graph.Close()
	order := []int{1, 0, 2, 3}
	result, err := graph.RichClubSequence(RichClubOptions{VertexOrder: order})
	if err != nil {
		t.Fatal(err)
	}
	assertSummaryFloatSlice(t, result, []float64{2, 0, 0, 0})
	if _, err := graph.RichClubSequence(RichClubOptions{VertexOrder: []int{0, 1}}); err == nil {
		t.Fatal("RichClubSequence(short order) returned nil error")
	}
	if _, err := graph.RichClubSequence(RichClubOptions{VertexOrder: []int{0, 1, 1, 3}}); err == nil {
		t.Fatal("RichClubSequence(duplicate order) returned nil error")
	}
	if _, err := graph.RichClubSequence(RichClubOptions{VertexOrder: []int{0, 1, 2, 4}}); err == nil {
		t.Fatal("RichClubSequence(out-of-range order) returned nil error")
	}
}

func TestStructuralSummariesErrorsAndClose(t *testing.T) {
	graph := mustDiagnosticGraph(t, false, 2, []Edge{{From: 0, To: 1}})
	invalidSelector := VertexSelector{kind: vertexSelectorIDs, ids: []int{2}}
	if _, err := graph.MaxDegree(invalidSelector, DegreeOptions{}); err == nil {
		t.Fatal("MaxDegree(invalid selector) returned nil error")
	}
	if _, err := graph.MaxDegree(AllVertices(), DegreeOptions{Direction: DirectionMode(99)}); err == nil {
		t.Fatal("MaxDegree(invalid direction) returned nil error")
	}
	if _, err := graph.AverageNearestNeighborDegree(AllVertices(), NearestNeighborDegreeOptions{Weights: []float64{1, 2}}); err == nil {
		t.Fatal("AverageNearestNeighborDegree(bad weights) returned nil error")
	}
	if _, err := graph.AverageNearestNeighborDegree(invalidSelector, NearestNeighborDegreeOptions{}); err == nil {
		t.Fatal("AverageNearestNeighborDegree(invalid selector) returned nil error")
	}
	if _, err := graph.AverageNearestNeighborDegree(AllVertices(), NearestNeighborDegreeOptions{NeighborDegreeDirection: DirectionMode(99)}); err == nil {
		t.Fatal("AverageNearestNeighborDegree(invalid neighbor direction) returned nil error")
	}
	if _, err := graph.DegreeCorrelation(DegreeCorrelationOptions{FromDirection: DirectionMode(99)}); err == nil {
		t.Fatal("DegreeCorrelation(invalid direction) returned nil error")
	}
	if _, err := graph.DegreeCorrelation(DegreeCorrelationOptions{ToDirection: DirectionMode(99)}); err == nil {
		t.Fatal("DegreeCorrelation(invalid target direction) returned nil error")
	}
	if _, err := graph.Diversity(AllVertices(), nil); err == nil {
		t.Fatal("Diversity(nil weights) returned nil error")
	}
	if _, err := graph.Diversity(AllVertices(), []float64{-1}); err == nil {
		t.Fatal("Diversity(negative weight) returned nil error")
	}
	if _, err := graph.Diversity(invalidSelector, []float64{1}); err == nil {
		t.Fatal("Diversity(invalid selector) returned nil error")
	}
	if _, err := graph.VerticesByDegree(invalidSelector, DegreeOptions{}, false); err == nil {
		t.Fatal("VerticesByDegree(invalid selector) returned nil error")
	}
	if _, err := graph.VerticesByDegree(AllVertices(), DegreeOptions{Direction: DirectionMode(99)}, false); err == nil {
		t.Fatal("VerticesByDegree(invalid direction) returned nil error")
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := graph.MeanDegree(false); !errors.Is(err, ErrClosed) {
		t.Fatalf("MeanDegree after Close error = %v, want ErrClosed", err)
	}
	if _, err := graph.MaxDegree(AllVertices(), DegreeOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("MaxDegree after Close error = %v, want ErrClosed", err)
	}
	if _, err := graph.DegreeCorrelation(DegreeCorrelationOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("DegreeCorrelation after Close error = %v, want ErrClosed", err)
	}
	if _, err := graph.AverageNearestNeighborDegree(AllVertices(), NearestNeighborDegreeOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("AverageNearestNeighborDegree after Close error = %v, want ErrClosed", err)
	}
	if _, err := graph.Reciprocity(ReciprocityDefault, true); !errors.Is(err, ErrClosed) {
		t.Fatalf("Reciprocity after Close error = %v, want ErrClosed", err)
	}
	if _, err := graph.Diversity(AllVertices(), []float64{1}); !errors.Is(err, ErrClosed) {
		t.Fatalf("Diversity after Close error = %v, want ErrClosed", err)
	}
	if _, err := graph.RichClubSequence(RichClubOptions{VertexOrder: []int{0, 1}}); !errors.Is(err, ErrClosed) {
		t.Fatalf("RichClubSequence after Close error = %v, want ErrClosed", err)
	}
	if _, err := graph.VerticesByDegree(AllVertices(), DegreeOptions{}, false); !errors.Is(err, ErrClosed) {
		t.Fatalf("VerticesByDegree after Close error = %v, want ErrClosed", err)
	}
}

func TestStructuralSummaryNilReceivers(t *testing.T) {
	var graph *Graph
	if _, err := graph.MeanDegree(false); !errors.Is(err, ErrClosed) {
		t.Fatal("nil MeanDegree did not return ErrClosed")
	}
	if _, err := graph.MaxDegree(AllVertices(), DegreeOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatal("nil MaxDegree did not return ErrClosed")
	}
	if _, err := graph.AverageNearestNeighborDegree(AllVertices(), NearestNeighborDegreeOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatal("nil AverageNearestNeighborDegree did not return ErrClosed")
	}
	if _, err := graph.DegreeCorrelation(DegreeCorrelationOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatal("nil DegreeCorrelation did not return ErrClosed")
	}
	if _, err := graph.Reciprocity(ReciprocityDefault, false); !errors.Is(err, ErrClosed) {
		t.Fatal("nil Reciprocity did not return ErrClosed")
	}
	if _, err := graph.Diversity(AllVertices(), nil); !errors.Is(err, ErrClosed) {
		t.Fatal("nil Diversity did not return ErrClosed")
	}
	if _, err := graph.RichClubSequence(RichClubOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatal("nil RichClubSequence did not return ErrClosed")
	}
	if _, err := graph.VerticesByDegree(AllVertices(), DegreeOptions{}, false); !errors.Is(err, ErrClosed) {
		t.Fatal("nil VerticesByDegree did not return ErrClosed")
	}
}

func TestStructuralSummaryLoopAndNormalizedBranches(t *testing.T) {
	graph := mustDiagnosticGraph(t, false, 2, []Edge{{From: 0, To: 0}, {From: 0, To: 1}})
	defer graph.Close()
	if got, err := graph.MaxDegree(AllVertices(), DegreeOptions{CountLoops: true}); err != nil || got != 3 {
		t.Fatalf("loop MaxDegree() = %d, %v, want 3, nil", got, err)
	}
	if got, err := graph.MeanDegree(true); err != nil || got != 2 {
		t.Fatalf("loop MeanDegree() = %v, %v, want 2, nil", got, err)
	}
	correlation, err := graph.DegreeCorrelation(DegreeCorrelationOptions{Weights: []float64{2, 1}})
	if err != nil || correlation == nil {
		t.Fatalf("weighted DegreeCorrelation() = %v, %v", correlation, err)
	}
	sequence, err := graph.RichClubSequence(RichClubOptions{
		VertexOrder: []int{0, 1}, Weights: []float64{2, 1}, Normalized: true,
		IncludeLoops: true,
	})
	if err != nil || len(sequence) != 2 {
		t.Fatalf("normalized RichClubSequence() = %v, %v", sequence, err)
	}
}

func TestMeanDegreeNullGraph(t *testing.T) {
	graph := mustDiagnosticGraph(t, false, 0, nil)
	defer graph.Close()
	value, err := graph.MeanDegree(true)
	if err != nil || !math.IsNaN(value) {
		t.Fatalf("null MeanDegree() = %v, %v, want NaN, nil", value, err)
	}
}

func assertSummaryFloatSlice(t *testing.T, got, want []float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("float slice length = %d, want %d: %v", len(got), len(want), got)
	}
	for index := range want {
		if math.IsNaN(want[index]) {
			if !math.IsNaN(got[index]) {
				t.Fatalf("float slice[%d] = %v, want NaN; full=%v", index, got[index], got)
			}
			continue
		}
		if math.Abs(got[index]-want[index]) > 1e-12 {
			t.Fatalf("float slice[%d] = %v, want %v; full=%v", index, got[index], want[index], got)
		}
	}
}

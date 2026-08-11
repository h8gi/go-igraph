package igraph

import (
	"errors"
	"math"
	"slices"
	"testing"
)

func TestFeedbackEdgeSetKnownAnswersAndValidity(t *testing.T) {
	tests := []struct {
		name     string
		vertices int
		edges    []Edge
		directed bool
		wantSize int
	}{
		{name: "empty directed", directed: true},
		{name: "DAG", vertices: 3, edges: []Edge{{0, 1}, {1, 2}}, directed: true},
		{name: "forest", vertices: 3, edges: []Edge{{0, 1}, {1, 2}}},
		{name: "directed cycle", vertices: 3, edges: []Edge{{0, 1}, {1, 2}, {2, 0}}, directed: true, wantSize: 1},
		{name: "undirected cycle", vertices: 3, edges: []Edge{{0, 1}, {1, 2}, {2, 0}}, wantSize: 1},
		{name: "disconnected directed cycles", vertices: 6, edges: []Edge{{0, 1}, {1, 2}, {2, 0}, {3, 4}, {4, 5}, {5, 3}}, directed: true, wantSize: 2},
		{name: "self loop", vertices: 1, edges: []Edge{{0, 0}}, directed: true, wantSize: 1},
		{name: "parallel undirected", vertices: 2, edges: []Edge{{0, 1}, {0, 1}, {0, 1}}, wantSize: 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := newCycleTestGraph(t, tc.vertices, tc.edges, tc.directed)
			result, err := g.FeedbackEdgeSet(FeedbackEdgeOptions{})
			_ = g.Close()
			if err != nil {
				t.Fatal(err)
			}
			if result == nil || len(result) != tc.wantSize {
				t.Errorf("FeedbackEdgeSet = %v, want non-nil size %d", result, tc.wantSize)
			}
			assertFeedbackEdgeSetValid(t, tc.vertices, tc.edges, tc.directed, result)
		})
	}
}

func TestFeedbackEdgeSetUndirectedMaximumSpanningForest(t *testing.T) {
	edges := []Edge{{0, 0}, {0, 1}, {0, 1}, {1, 2}, {2, 0}}
	weights := []float64{100, 5, 1, 4, 3}
	g := newCycleTestGraph(t, 3, edges, false)
	defer g.Close()
	for _, strategy := range []FeedbackEdgeStrategy{FeedbackEdgeExact, FeedbackEdgeApproximateEades} {
		result, err := g.FeedbackEdgeSet(FeedbackEdgeOptions{Strategy: strategy, Weights: weights})
		if err != nil {
			t.Fatal(err)
		}
		slices.Sort(result)
		want := []int{0, 2, 4}
		if !slices.Equal(result, want) {
			t.Errorf("strategy %d result = %v, want %v", strategy, result, want)
		}
		assertFeedbackEdgeSetValid(t, 3, edges, false, result)
	}
}

func TestFeedbackEdgeSetExactWeightedObjective(t *testing.T) {
	edges := []Edge{{0, 1}, {1, 2}, {2, 0}, {1, 3}, {3, 0}}
	weights := []float64{10, 1, 1, 1, 1}
	g := newCycleTestGraph(t, 4, edges, true)
	defer g.Close()
	result, err := g.FeedbackEdgeSet(FeedbackEdgeOptions{Weights: weights})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Errorf("weighted exact result cardinality = %d, want 2 to avoid expensive shared edge", len(result))
	}
	gotWeight := selectedWeight(result, weights)
	wantWeight := exhaustiveMinimumFeedbackEdgeWeight(t, 4, edges, true, weights)
	if gotWeight != wantWeight || gotWeight != 2 {
		t.Errorf("weighted exact result %v cost = %v, exhaustive optimum = %v", result, gotWeight, wantWeight)
	}
	assertFeedbackEdgeSetValid(t, 4, edges, true, result)
}

func TestFeedbackEdgeSetApproximateAndZeroWeights(t *testing.T) {
	edges := []Edge{{0, 1}, {1, 2}, {2, 0}, {1, 3}, {3, 0}}
	g := newCycleTestGraph(t, 4, edges, true)
	defer g.Close()
	approximate, err := g.FeedbackEdgeSet(FeedbackEdgeOptions{
		Strategy: FeedbackEdgeApproximateEades,
		Weights:  []float64{1, 2, 3, 4, 5},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertFeedbackEdgeSetValid(t, 4, edges, true, approximate)
	zeros := make([]float64, len(edges))
	zeroResult, err := g.FeedbackEdgeSet(FeedbackEdgeOptions{Weights: zeros})
	if err != nil {
		t.Fatal(err)
	}
	if selectedWeight(zeroResult, zeros) != 0 {
		t.Errorf("zero-weight result has non-zero cost: %v", zeroResult)
	}
	assertFeedbackEdgeSetValid(t, 4, edges, true, zeroResult)
}

func TestFeedbackVertexSetKnownAnswersAndWeightedObjective(t *testing.T) {
	for _, tc := range []struct {
		name     string
		vertices int
		edges    []Edge
		directed bool
		wantSize int
	}{
		{name: "empty", directed: true},
		{name: "DAG", vertices: 3, edges: []Edge{{0, 1}, {1, 2}}, directed: true},
		{name: "directed triangle", vertices: 3, edges: []Edge{{0, 1}, {1, 2}, {2, 0}}, directed: true, wantSize: 1},
		{name: "undirected triangle", vertices: 3, edges: []Edge{{0, 1}, {1, 2}, {2, 0}}, wantSize: 1},
		{name: "self loop", vertices: 2, edges: []Edge{{1, 1}}, wantSize: 1},
		{name: "parallel", vertices: 2, edges: []Edge{{0, 1}, {0, 1}}, wantSize: 1},
		{name: "disconnected", vertices: 6, edges: []Edge{{0, 1}, {1, 2}, {2, 0}, {3, 4}, {4, 5}, {5, 3}}, wantSize: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCycleTestGraph(t, tc.vertices, tc.edges, tc.directed)
			result, err := g.FeedbackVertexSet(nil)
			_ = g.Close()
			if err != nil {
				t.Fatal(err)
			}
			if result == nil || len(result) != tc.wantSize {
				t.Errorf("FeedbackVertexSet = %v, want non-nil size %d", result, tc.wantSize)
			}
			assertFeedbackVertexSetValid(t, tc.vertices, tc.edges, tc.directed, result)
		})
	}

	edges := []Edge{{0, 1}, {1, 2}, {2, 0}, {1, 3}, {3, 0}}
	weights := []float64{10, 10, 1, 1}
	g := newCycleTestGraph(t, 4, edges, true)
	defer g.Close()
	result, err := g.FeedbackVertexSet(weights)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Errorf("weighted vertex result cardinality = %d, want 2", len(result))
	}
	gotWeight := selectedWeight(result, weights)
	wantWeight := exhaustiveMinimumFeedbackVertexWeight(t, 4, edges, true, weights)
	if gotWeight != wantWeight || gotWeight != 2 {
		t.Errorf("weighted vertex result %v cost = %v, optimum = %v", result, gotWeight, wantWeight)
	}
	assertFeedbackVertexSetValid(t, 4, edges, true, result)
}

func TestFeedbackVertexSetZeroWeightsAndOwnership(t *testing.T) {
	edges := []Edge{{0, 1}, {1, 2}, {2, 0}}
	g := newCycleTestGraph(t, 3, edges, true)
	result, err := g.FeedbackVertexSet([]float64{0, 0, 0})
	if err != nil {
		t.Fatal(err)
	}
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	if result == nil || len(result) == 0 || selectedWeight(result, []float64{0, 0, 0}) != 0 {
		t.Errorf("zero-weight vertex result = %v", result)
	}
	assertFeedbackVertexSetValid(t, 3, edges, true, result)
	result[0] = 99
}

func TestFeedbackValidationAndUseAfterClose(t *testing.T) {
	g := newCycleTestGraph(t, 2, []Edge{{0, 1}}, true)
	defer g.Close()
	invalidEdgeWeights := [][]float64{
		{}, {-1}, {math.NaN()}, {math.Inf(1)}, {math.Inf(-1)},
	}
	for _, weights := range invalidEdgeWeights {
		if _, err := g.FeedbackEdgeSet(FeedbackEdgeOptions{Weights: weights}); err == nil {
			t.Errorf("FeedbackEdgeSet weights %v succeeded", weights)
		}
	}
	if _, err := g.FeedbackEdgeSet(FeedbackEdgeOptions{Strategy: FeedbackEdgeStrategy(99)}); err == nil {
		t.Error("FeedbackEdgeSet invalid strategy succeeded")
	}
	invalidVertexWeights := [][]float64{
		{}, {-1, 0}, {math.NaN(), 0}, {math.Inf(1), 0}, {math.Inf(-1), 0},
	}
	for _, weights := range invalidVertexWeights {
		if _, err := g.FeedbackVertexSet(weights); err == nil {
			t.Errorf("FeedbackVertexSet weights %v succeeded", weights)
		}
	}

	closed := newCycleTestGraph(t, 0, nil, true)
	_ = closed.Close()
	for _, graph := range []*Graph{nil, closed} {
		if _, err := graph.FeedbackEdgeSet(FeedbackEdgeOptions{}); !errors.Is(err, ErrClosed) {
			t.Errorf("FeedbackEdgeSet closed error = %v", err)
		}
		if _, err := graph.FeedbackVertexSet(nil); !errors.Is(err, ErrClosed) {
			t.Errorf("FeedbackVertexSet closed error = %v", err)
		}
	}
}

func assertFeedbackEdgeSetValid(t *testing.T, vertexCount int, edges []Edge, directed bool, ids []int) {
	t.Helper()
	assertFeedbackIDs(t, ids, len(edges), "edge")
	selector, err := EdgeIDs(ids...)
	if err != nil {
		t.Fatal(err)
	}
	g := newCycleTestGraph(t, vertexCount, edges, directed)
	defer g.Close()
	if _, err := g.DeleteEdges(selector); err != nil {
		t.Fatal(err)
	}
	acyclic, err := g.IsAcyclic()
	if err != nil {
		t.Fatal(err)
	}
	if !acyclic {
		t.Errorf("removing feedback edges %v did not make graph acyclic", ids)
	}
}

func assertFeedbackVertexSetValid(t *testing.T, vertexCount int, edges []Edge, directed bool, ids []int) {
	t.Helper()
	assertFeedbackIDs(t, ids, vertexCount, "vertex")
	selector, err := VertexIDs(ids...)
	if err != nil {
		t.Fatal(err)
	}
	g := newCycleTestGraph(t, vertexCount, edges, directed)
	defer g.Close()
	if _, err := g.DeleteVertices(selector); err != nil {
		t.Fatal(err)
	}
	acyclic, err := g.IsAcyclic()
	if err != nil {
		t.Fatal(err)
	}
	if !acyclic {
		t.Errorf("removing feedback vertices %v did not make graph acyclic", ids)
	}
}

func assertFeedbackIDs(t *testing.T, ids []int, count int, kind string) {
	t.Helper()
	seen := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		if id < 0 || id >= count {
			t.Errorf("feedback %s ID %d out of range [0, %d)", kind, id, count)
		}
		if _, duplicate := seen[id]; duplicate {
			t.Errorf("feedback %s ID %d is duplicated in %v", kind, id, ids)
		}
		seen[id] = struct{}{}
	}
}

func selectedWeight(ids []int, weights []float64) float64 {
	var total float64
	for _, id := range ids {
		total += weights[id]
	}
	return total
}

func exhaustiveMinimumFeedbackEdgeWeight(t *testing.T, vertexCount int, edges []Edge, directed bool, weights []float64) float64 {
	t.Helper()
	minimum := math.Inf(1)
	for mask := 0; mask < 1<<len(edges); mask++ {
		ids := maskIDs(mask, len(edges))
		if selectedWeight(ids, weights) >= minimum {
			continue
		}
		if feedbackEdgeSetIsValid(t, vertexCount, edges, directed, ids) {
			minimum = selectedWeight(ids, weights)
		}
	}
	return minimum
}

func exhaustiveMinimumFeedbackVertexWeight(t *testing.T, vertexCount int, edges []Edge, directed bool, weights []float64) float64 {
	t.Helper()
	minimum := math.Inf(1)
	for mask := 0; mask < 1<<vertexCount; mask++ {
		ids := maskIDs(mask, vertexCount)
		if selectedWeight(ids, weights) >= minimum {
			continue
		}
		if feedbackVertexSetIsValid(t, vertexCount, edges, directed, ids) {
			minimum = selectedWeight(ids, weights)
		}
	}
	return minimum
}

func feedbackEdgeSetIsValid(t *testing.T, vertexCount int, edges []Edge, directed bool, ids []int) bool {
	t.Helper()
	selector, _ := EdgeIDs(ids...)
	g := newCycleTestGraph(t, vertexCount, edges, directed)
	defer g.Close()
	if _, err := g.DeleteEdges(selector); err != nil {
		t.Fatal(err)
	}
	acyclic, err := g.IsAcyclic()
	if err != nil {
		t.Fatal(err)
	}
	return acyclic
}

func feedbackVertexSetIsValid(t *testing.T, vertexCount int, edges []Edge, directed bool, ids []int) bool {
	t.Helper()
	selector, _ := VertexIDs(ids...)
	g := newCycleTestGraph(t, vertexCount, edges, directed)
	defer g.Close()
	if _, err := g.DeleteVertices(selector); err != nil {
		t.Fatal(err)
	}
	acyclic, err := g.IsAcyclic()
	if err != nil {
		t.Fatal(err)
	}
	return acyclic
}

func maskIDs(mask, count int) []int {
	result := make([]int, 0, count)
	for id := 0; id < count; id++ {
		if mask&(1<<id) != 0 {
			result = append(result, id)
		}
	}
	return result
}

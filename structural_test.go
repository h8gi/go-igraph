package igraph

import (
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestDensityKnownAnswersLoopsAndWeights(t *testing.T) {
	path := newStructuralGraph(t, 4, []Edge{{0, 1}, {1, 2}, {2, 3}}, false)
	assertFloat(t, density(t, path, DensityOptions{}), 0.5)
	assertFloat(t, density(t, path, DensityOptions{IncludeLoops: true}), 0.3)
	assertFloat(t, density(t, path, DensityOptions{Weights: []float64{2, 3, 4}}), 1.5)

	cycle := newStructuralGraph(t, 4, []Edge{{0, 1}, {1, 2}, {2, 3}, {3, 0}}, false)
	assertFloat(t, density(t, cycle, DensityOptions{}), 2.0/3.0)
	complete := newStructuralGraph(t, 4, []Edge{
		{0, 1}, {0, 2}, {0, 3}, {1, 2}, {1, 3}, {2, 3},
	}, false)
	assertFloat(t, density(t, complete, DensityOptions{}), 1)

	directed := newStructuralGraph(t, 3, []Edge{{0, 1}, {1, 2}}, true)
	assertFloat(t, density(t, directed, DensityOptions{}), 1.0/3.0)
	assertFloat(t, density(t, directed, DensityOptions{IncludeLoops: true}), 2.0/9.0)

	loops := newStructuralGraph(t, 2, []Edge{{0, 0}, {0, 1}}, false)
	assertFloat(t, density(t, loops, DensityOptions{IncludeLoops: true}), 2.0/3.0)
	assertFloat(t, density(t, loops, DensityOptions{
		IncludeLoops: true,
		Weights:      []float64{-1, 4},
	}), 1)

	empty := newStructuralGraph(t, 0, nil, false)
	if got := density(t, empty, DensityOptions{Weights: []float64{}}); !math.IsNaN(got) {
		t.Errorf("empty density = %v, want NaN", got)
	}
}

func TestDiameterKnownAnswersWeightsAndConsistency(t *testing.T) {
	path := newStructuralGraph(t, 4, []Edge{{0, 1}, {1, 2}, {2, 3}}, false)
	result, err := path.Diameter(DistanceSummaryOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertFloat(t, result.Length, 3)
	assertDiameterConsistency(t, path, result, DirectionOut, nil)

	weighted, err := path.Diameter(DistanceSummaryOptions{Weights: []float64{2, 3, 4}})
	if err != nil {
		t.Fatal(err)
	}
	assertFloat(t, weighted.Length, 9)
	assertDiameterConsistency(t, path, weighted, DirectionOut, []float64{2, 3, 4})

	cycle := newStructuralGraph(t, 4, []Edge{{0, 1}, {1, 2}, {2, 3}, {3, 0}}, false)
	cycleResult, err := cycle.Diameter(DistanceSummaryOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertFloat(t, cycleResult.Length, 2)
	assertDiameterConsistency(t, cycle, cycleResult, DirectionOut, nil)

	complete := newStructuralGraph(t, 4, []Edge{
		{0, 1}, {0, 2}, {0, 3}, {1, 2}, {1, 3}, {2, 3},
	}, false)
	completeResult, err := complete.Diameter(DistanceSummaryOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertFloat(t, completeResult.Length, 1)
	assertDiameterConsistency(t, complete, completeResult, DirectionOut, nil)
}

func TestDiameterDirectionDisconnectedAndEmpty(t *testing.T) {
	directed := newStructuralGraph(t, 3, []Edge{{0, 1}, {1, 2}}, true)
	outgoing, err := directed.Diameter(DistanceSummaryOptions{
		Direction:         DirectionOut,
		IgnoreUnreachable: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if outgoing.From != 0 || outgoing.To != 2 || !reflect.DeepEqual(outgoing.Path.Vertices, []int{0, 1, 2}) {
		t.Errorf("outgoing diameter = %#v", outgoing)
	}
	assertDiameterConsistency(t, directed, outgoing, DirectionOut, nil)
	incoming, err := directed.Diameter(DistanceSummaryOptions{
		Direction:         DirectionIn,
		IgnoreUnreachable: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if incoming.From != 2 || incoming.To != 0 || !reflect.DeepEqual(incoming.Path.Vertices, []int{2, 1, 0}) {
		t.Errorf("incoming diameter = %#v", incoming)
	}
	assertDiameterConsistency(t, directed, incoming, DirectionIn, nil)
	all, err := directed.Diameter(DistanceSummaryOptions{Direction: DirectionAll})
	if err != nil {
		t.Fatal(err)
	}
	assertFloat(t, all.Length, 2)
	assertDiameterConsistency(t, directed, all, DirectionAll, nil)

	disconnected := newStructuralGraph(t, 4, []Edge{{0, 1}, {1, 2}}, false)
	within, err := disconnected.Diameter(DistanceSummaryOptions{IgnoreUnreachable: true})
	if err != nil {
		t.Fatal(err)
	}
	assertFloat(t, within.Length, 2)
	assertDiameterConsistency(t, disconnected, within, DirectionOut, nil)
	strict, err := disconnected.Diameter(DistanceSummaryOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsInf(strict.Length, 1) || strict.Path.Found || strict.From != -1 || strict.To != -1 || strict.Path.Vertices == nil || strict.Path.Edges == nil {
		t.Errorf("strict disconnected diameter = %#v", strict)
	}

	empty := newStructuralGraph(t, 0, nil, false)
	emptyResult, err := empty.Diameter(DistanceSummaryOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsNaN(emptyResult.Length) || emptyResult.From != -1 || emptyResult.To != -1 || emptyResult.Path.Found || emptyResult.Path.Vertices == nil || emptyResult.Path.Edges == nil {
		t.Errorf("empty diameter = %#v", emptyResult)
	}
}

func TestAveragePathLengthKnownAnswersWeightsAndDisconnectedPairs(t *testing.T) {
	path := newStructuralGraph(t, 4, []Edge{{0, 1}, {1, 2}, {2, 3}}, false)
	assertAverage(t, averagePathLength(t, path, DistanceSummaryOptions{}), 5.0/3.0, 0)
	assertAverage(t, averagePathLength(t, path, DistanceSummaryOptions{
		Weights: []float64{2, 3, 4},
	}), 5, 0)

	cycle := newStructuralGraph(t, 4, []Edge{{0, 1}, {1, 2}, {2, 3}, {3, 0}}, false)
	assertAverage(t, averagePathLength(t, cycle, DistanceSummaryOptions{}), 4.0/3.0, 0)
	complete := newStructuralGraph(t, 4, []Edge{
		{0, 1}, {0, 2}, {0, 3}, {1, 2}, {1, 3}, {2, 3},
	}, false)
	assertAverage(t, averagePathLength(t, complete, DistanceSummaryOptions{}), 1, 0)

	disconnected := newStructuralGraph(t, 4, []Edge{{0, 1}, {1, 2}}, false)
	assertAverage(t, averagePathLength(t, disconnected, DistanceSummaryOptions{
		IgnoreUnreachable: true,
	}), 4.0/3.0, 6)
	strict := averagePathLength(t, disconnected, DistanceSummaryOptions{})
	if !math.IsInf(strict.Length, 1) || strict.UnreachablePairs != 6 {
		t.Errorf("strict disconnected average = %#v", strict)
	}

	empty := newStructuralGraph(t, 0, nil, false)
	emptyResult := averagePathLength(t, empty, DistanceSummaryOptions{})
	if !math.IsNaN(emptyResult.Length) || emptyResult.UnreachablePairs != 0 {
		t.Errorf("empty average = %#v", emptyResult)
	}
}

func TestAveragePathLengthDirection(t *testing.T) {
	graph := newStructuralGraph(t, 3, []Edge{{0, 1}, {0, 2}}, true)
	for _, direction := range []DirectionMode{DirectionOut, DirectionIn} {
		assertAverage(t, averagePathLength(t, graph, DistanceSummaryOptions{
			Direction:         direction,
			IgnoreUnreachable: true,
		}), 1, 4)
	}
	assertAverage(t, averagePathLength(t, graph, DistanceSummaryOptions{
		Direction: DirectionAll,
	}), 4.0/3.0, 0)
}

func TestTransitivityKnownAnswersAndUndefinedModes(t *testing.T) {
	graph := newStructuralGraph(t, 4, []Edge{{0, 1}, {1, 2}, {2, 0}, {2, 3}}, false)
	global, err := graph.GlobalTransitivity(TransitivityNaN)
	if err != nil {
		t.Fatal(err)
	}
	assertFloat(t, global, 3.0/5.0)
	averageNaN, err := graph.AverageLocalTransitivity(TransitivityNaN)
	if err != nil {
		t.Fatal(err)
	}
	assertFloat(t, averageNaN, 7.0/9.0)
	averageZero, err := graph.AverageLocalTransitivity(TransitivityZero)
	if err != nil {
		t.Fatal(err)
	}
	assertFloat(t, averageZero, 7.0/12.0)

	selector, _ := VertexIDs(2, 0, 2, 3, 1)
	localNaN, err := graph.LocalTransitivity(selector, TransitivityNaN)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, localNaN, []float64{1.0 / 3.0, 1, 1.0 / 3.0, math.NaN(), 1})
	localZero, err := graph.LocalTransitivity(selector, TransitivityZero)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, localZero, []float64{1.0 / 3.0, 1, 1.0 / 3.0, 0, 1})

	path := newStructuralGraph(t, 4, []Edge{{0, 1}, {1, 2}, {2, 3}}, false)
	assertFloat(t, globalTransitivity(t, path, TransitivityNaN), 0)
	assertFloat(t, averageLocalTransitivity(t, path, TransitivityNaN), 0)
	cycle := newStructuralGraph(t, 4, []Edge{{0, 1}, {1, 2}, {2, 3}, {3, 0}}, false)
	assertFloat(t, globalTransitivity(t, cycle, TransitivityNaN), 0)
	complete := newStructuralGraph(t, 4, []Edge{
		{0, 1}, {0, 2}, {0, 3}, {1, 2}, {1, 3}, {2, 3},
	}, false)
	assertFloat(t, globalTransitivity(t, complete, TransitivityNaN), 1)
	assertFloat(t, averageLocalTransitivity(t, complete, TransitivityNaN), 1)

	isolates := newStructuralGraph(t, 3, nil, false)
	if got := globalTransitivity(t, isolates, TransitivityNaN); !math.IsNaN(got) {
		t.Errorf("isolate global transitivity = %v, want NaN", got)
	}
	assertFloat(t, globalTransitivity(t, isolates, TransitivityZero), 0)
	if got := averageLocalTransitivity(t, isolates, TransitivityNaN); !math.IsNaN(got) {
		t.Errorf("isolate average local transitivity = %v, want NaN", got)
	}
	assertFloat(t, averageLocalTransitivity(t, isolates, TransitivityZero), 0)
	assertFloatSlice(t, localTransitivity(t, isolates, AllVertices(), TransitivityNaN), []float64{
		math.NaN(), math.NaN(), math.NaN(),
	})
	assertFloatSlice(t, localTransitivity(t, isolates, AllVertices(), TransitivityZero), []float64{0, 0, 0})
}

func TestTransitivityIgnoresDirectionMultiplicityAndEmptySelections(t *testing.T) {
	directed := newStructuralGraph(t, 3, []Edge{{0, 1}, {1, 2}, {2, 0}}, true)
	assertFloat(t, globalTransitivity(t, directed, TransitivityNaN), 1)
	assertFloatSlice(t, localTransitivity(t, directed, AllVertices(), TransitivityNaN), []float64{1, 1, 1})

	multiple := newStructuralGraph(t, 3, []Edge{{0, 1}, {0, 1}, {1, 2}, {2, 0}, {0, 0}}, false)
	assertFloat(t, globalTransitivity(t, multiple, TransitivityNaN), 1)
	assertFloatSlice(t, localTransitivity(t, multiple, AllVertices(), TransitivityNaN), []float64{1, 1, 1})

	emptySelection := localTransitivity(t, directed, NoVertices(), TransitivityNaN)
	if emptySelection == nil || len(emptySelection) != 0 {
		t.Errorf("empty local result = %#v, want non-nil empty", emptySelection)
	}
	emptyGraph := newStructuralGraph(t, 0, nil, false)
	emptyAll := localTransitivity(t, emptyGraph, AllVertices(), TransitivityZero)
	if emptyAll == nil || len(emptyAll) != 0 {
		t.Errorf("empty graph local result = %#v, want non-nil empty", emptyAll)
	}
}

func TestStructuralMetricsRejectInvalidInputsAndPropagateUpstreamErrors(t *testing.T) {
	graph := newStructuralGraph(t, 3, []Edge{{0, 1}, {1, 2}}, true)
	invalidDirection := DistanceSummaryOptions{Direction: DirectionMode(99)}
	if _, err := graph.Diameter(invalidDirection); err == nil || !strings.Contains(err.Error(), "direction mode") {
		t.Errorf("invalid Diameter direction error = %v", err)
	}
	if _, err := graph.AveragePathLength(invalidDirection); err == nil || !strings.Contains(err.Error(), "direction mode") {
		t.Errorf("invalid AveragePathLength direction error = %v", err)
	}

	for _, weights := range [][]float64{
		{}, {1}, {1, 2, 3}, {math.NaN(), 1}, {math.Inf(1), 1}, {math.Inf(-1), 1},
	} {
		if _, err := graph.Density(DensityOptions{Weights: weights}); err == nil {
			t.Errorf("Density weights %#v error = nil", weights)
		}
		options := DistanceSummaryOptions{Weights: weights}
		if _, err := graph.Diameter(options); err == nil {
			t.Errorf("Diameter weights %#v error = nil", weights)
		}
		if _, err := graph.AveragePathLength(options); err == nil {
			t.Errorf("AveragePathLength weights %#v error = nil", weights)
		}
	}
	negative := DistanceSummaryOptions{Weights: []float64{1, -1}}
	if _, err := graph.Diameter(negative); err == nil || !strings.Contains(err.Error(), "calculate diameter") {
		t.Errorf("negative Diameter error = %v", err)
	}
	if _, err := graph.AveragePathLength(negative); err == nil || !strings.Contains(err.Error(), "calculate average path length") {
		t.Errorf("negative AveragePathLength error = %v", err)
	}

	invalidSelector, _ := VertexIDs(3)
	if _, err := graph.LocalTransitivity(invalidSelector, TransitivityNaN); err == nil {
		t.Error("invalid LocalTransitivity selector error = nil")
	}
	invalidKind := VertexSelector{kind: vertexSelectorKind(255)}
	if _, err := graph.LocalTransitivity(invalidKind, TransitivityNaN); err == nil {
		t.Error("invalid-kind LocalTransitivity selector error = nil")
	}
	invalidMode := TransitivityMode(99)
	if _, err := graph.GlobalTransitivity(invalidMode); err == nil {
		t.Error("invalid GlobalTransitivity mode error = nil")
	}
	if _, err := graph.LocalTransitivity(AllVertices(), invalidMode); err == nil {
		t.Error("invalid LocalTransitivity mode error = nil")
	}
	if _, err := graph.AverageLocalTransitivity(invalidMode); err == nil {
		t.Error("invalid AverageLocalTransitivity mode error = nil")
	}
}

func TestStructuralResultsAreGoOwnedAndClosedGraphsFail(t *testing.T) {
	graph := newStructuralGraph(t, 4, []Edge{{0, 1}, {1, 2}, {2, 0}, {2, 3}}, false)
	diameter, err := graph.Diameter(DistanceSummaryOptions{})
	if err != nil {
		t.Fatal(err)
	}
	local, err := graph.LocalTransitivity(AllVertices(), TransitivityNaN)
	if err != nil {
		t.Fatal(err)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	diameter.Path.Vertices[0] = 99
	local[0] = 99
	assertStructuralClosed(t, graph)
	var nilGraph *Graph
	assertStructuralClosed(t, nilGraph)
}

func newStructuralGraph(t *testing.T, vertexCount int, edges []Edge, directed bool) *Graph {
	t.Helper()
	graph, err := NewGraphFromEdges(vertexCount, edges, directed)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	return graph
}

func density(t *testing.T, graph *Graph, options DensityOptions) float64 {
	t.Helper()
	result, err := graph.Density(options)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func averagePathLength(t *testing.T, graph *Graph, options DistanceSummaryOptions) AveragePathLengthResult {
	t.Helper()
	result, err := graph.AveragePathLength(options)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func globalTransitivity(t *testing.T, graph *Graph, mode TransitivityMode) float64 {
	t.Helper()
	result, err := graph.GlobalTransitivity(mode)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func localTransitivity(t *testing.T, graph *Graph, selector VertexSelector, mode TransitivityMode) []float64 {
	t.Helper()
	result, err := graph.LocalTransitivity(selector, mode)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func averageLocalTransitivity(t *testing.T, graph *Graph, mode TransitivityMode) float64 {
	t.Helper()
	result, err := graph.AverageLocalTransitivity(mode)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func assertDiameterConsistency(t *testing.T, graph *Graph, result DiameterResult, direction DirectionMode, weights []float64) {
	t.Helper()
	if !result.Path.Found || len(result.Path.Vertices) == 0 {
		t.Fatalf("diameter has no path: %#v", result)
	}
	if result.From != result.Path.Vertices[0] || result.To != result.Path.Vertices[len(result.Path.Vertices)-1] {
		t.Errorf("diameter endpoints (%d, %d) disagree with path %#v", result.From, result.To, result.Path.Vertices)
	}
	assertPathEdges(t, graph, result.Path, direction)
	length := float64(len(result.Path.Edges))
	if weights != nil {
		length = 0
		for _, edgeID := range result.Path.Edges {
			length += weights[edgeID]
		}
	}
	assertFloat(t, result.Length, length)
}

func assertAverage(t *testing.T, got AveragePathLengthResult, length, unreachable float64) {
	t.Helper()
	assertFloat(t, got.Length, length)
	assertFloat(t, got.UnreachablePairs, unreachable)
}

func assertFloatSlice(t *testing.T, got, want []float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("slice length = %d, want %d: %#v", len(got), len(want), got)
	}
	for index := range want {
		assertFloat(t, got[index], want[index])
	}
}

func assertFloat(t *testing.T, got, want float64) {
	t.Helper()
	if math.IsNaN(want) {
		if !math.IsNaN(got) {
			t.Errorf("value = %v, want NaN", got)
		}
		return
	}
	if math.IsNaN(got) {
		t.Errorf("value = NaN, want %v", want)
		return
	}
	if math.IsInf(want, 0) {
		if got != want {
			t.Errorf("value = %v, want %v", got, want)
		}
		return
	}
	if math.IsInf(got, 0) {
		t.Errorf("value = %v, want %v", got, want)
		return
	}
	if math.Abs(got-want) > 1e-12 {
		t.Errorf("value = %.16g, want %.16g", got, want)
	}
}

func assertStructuralClosed(t *testing.T, graph *Graph) {
	t.Helper()
	if _, err := graph.Density(DensityOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("Density closed error = %v, want %v", err, ErrClosed)
	}
	if _, err := graph.Diameter(DistanceSummaryOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("Diameter closed error = %v, want %v", err, ErrClosed)
	}
	if _, err := graph.AveragePathLength(DistanceSummaryOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("AveragePathLength closed error = %v, want %v", err, ErrClosed)
	}
	if _, err := graph.GlobalTransitivity(TransitivityNaN); !errors.Is(err, ErrClosed) {
		t.Errorf("GlobalTransitivity closed error = %v, want %v", err, ErrClosed)
	}
	if _, err := graph.LocalTransitivity(AllVertices(), TransitivityNaN); !errors.Is(err, ErrClosed) {
		t.Errorf("LocalTransitivity closed error = %v, want %v", err, ErrClosed)
	}
	if _, err := graph.AverageLocalTransitivity(TransitivityNaN); !errors.Is(err, ErrClosed) {
		t.Errorf("AverageLocalTransitivity closed error = %v, want %v", err, ErrClosed)
	}
}

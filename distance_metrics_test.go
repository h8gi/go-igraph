package igraph

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestCutoffDistancesPreserveSelectors(t *testing.T) {
	graph := newPathTestGraph(t)
	sources, _ := VertexIDs(0, 2)
	targets, _ := VertexIDs(3, 1, 3)
	got, err := graph.CutoffDistances(sources, targets, 1, PathOptions{Direction: DirectionOut})
	if err != nil {
		t.Fatal(err)
	}
	assertMatrixRows(t, got, [][]float64{
		{math.Inf(1), 1, math.Inf(1)},
		{1, math.Inf(1), 1},
	})

	weighted, err := graph.CutoffDistances(sources, targets, 2, PathOptions{
		Direction: DirectionOut,
		Weights:   []float64{2, 3, 10, 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertMatrixRows(t, weighted, [][]float64{
		{math.Inf(1), 2, math.Inf(1)},
		{1, math.Inf(1), 1},
	})
}

func TestDistanceDerivedMetrics(t *testing.T) {
	graph, err := NewGraphFromEdges(4, []Edge{{0, 1}, {1, 2}, {2, 3}}, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })

	vertices, _ := VertexIDs(3, 1, 3)
	eccentricities, err := graph.Eccentricities(vertices, PathOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(eccentricities, []float64{3, 2, 3}) {
		t.Errorf("eccentricities = %v", eccentricities)
	}
	if radius, err := graph.Radius(PathOptions{}); err != nil || radius != 2 {
		t.Errorf("radius = %v, %v", radius, err)
	}
	center, err := graph.GraphCenter(PathOptions{})
	if err != nil || !reflect.DeepEqual(center, []int{1, 2}) {
		t.Errorf("center = %v, %v", center, err)
	}

	global, err := graph.GlobalEfficiency(false, nil)
	if err != nil || math.Abs(global-13.0/18.0) > 1e-12 {
		t.Errorf("global efficiency = %v, %v", global, err)
	}
	local, err := graph.LocalEfficiencies(AllVertices(), false, DirectionAll, nil)
	if err != nil || !reflect.DeepEqual(local, []float64{0, 0, 0, 0}) {
		t.Errorf("local efficiencies = %v, %v", local, err)
	}
	if average, err := graph.AverageLocalEfficiency(false, DirectionAll, nil); err != nil || average != 0 {
		t.Errorf("average local efficiency = %v, %v", average, err)
	}

	histogram, err := graph.PathLengthHistogram(false)
	if err != nil || !reflect.DeepEqual(histogram.Counts, []float64{3, 2, 1}) || histogram.Unreachable != 0 {
		t.Errorf("histogram = %#v, %v", histogram, err)
	}
	start := 0
	pseudo, err := graph.PseudoDiameter(PseudoDiameterOptions{Start: &start, Disconnected: true})
	if err != nil || pseudo.Diameter != 3 || pseudo.From != 0 || pseudo.To != 3 {
		t.Errorf("pseudo-diameter = %#v, %v", pseudo, err)
	}
}

func TestDistanceMetricsEmptyInvalidAndClosed(t *testing.T) {
	empty, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	if radius, err := empty.Radius(PathOptions{}); err != nil || !math.IsNaN(radius) {
		t.Errorf("empty radius = %v, %v", radius, err)
	}
	center, err := empty.GraphCenter(PathOptions{})
	if err != nil || center == nil || len(center) != 0 {
		t.Errorf("empty center = %v, %v", center, err)
	}
	pseudo, err := empty.PseudoDiameter(PseudoDiameterOptions{})
	if err != nil || !math.IsNaN(pseudo.Diameter) || pseudo.From != -1 || pseudo.To != -1 {
		t.Errorf("empty pseudo-diameter = %#v, %v", pseudo, err)
	}
	if err := empty.Close(); err != nil {
		t.Fatal(err)
	}
	assertDistanceMetricsClosed(t, empty)
	var nilGraph *Graph
	assertDistanceMetricsClosed(t, nilGraph)

	graph := newPathTestGraph(t)
	if _, err := graph.CutoffDistances(AllVertices(), AllVertices(), -1, PathOptions{}); err == nil {
		t.Error("negative cutoff accepted")
	}
	if _, err := graph.Eccentricities(AllVertices(), PathOptions{Weights: []float64{1, -1, 1, 1}}); err == nil {
		t.Error("negative weight accepted")
	}
	invalidSelector, _ := VertexIDs(5)
	if _, err := graph.Eccentricities(invalidSelector, PathOptions{}); err == nil {
		t.Error("invalid eccentricity selector accepted")
	}
	if _, err := graph.LocalEfficiencies(invalidSelector, false, DirectionAll, nil); err == nil {
		t.Error("invalid local-efficiency selector accepted")
	}
	if _, err := graph.Radius(PathOptions{Direction: DirectionMode(99)}); err == nil {
		t.Error("invalid radius direction accepted")
	}
	if _, err := graph.GraphCenter(PathOptions{Weights: []float64{1}}); err == nil {
		t.Error("invalid center weights accepted")
	}
	if _, err := graph.GlobalEfficiency(false, []float64{1}); err == nil {
		t.Error("invalid efficiency weights accepted")
	}
	if _, err := graph.AverageLocalEfficiency(false, DirectionMode(99), nil); err == nil {
		t.Error("invalid local-efficiency direction accepted")
	}
	badStart := 5
	if _, err := graph.PseudoDiameter(PseudoDiameterOptions{Start: &badStart}); err == nil {
		t.Error("invalid pseudo-diameter start accepted")
	}
}

func assertDistanceMetricsClosed(t *testing.T, graph *Graph) {
	t.Helper()
	checks := []error{}
	_, err := graph.CutoffDistances(AllVertices(), AllVertices(), 1, PathOptions{})
	checks = append(checks, err)
	_, err = graph.Eccentricities(AllVertices(), PathOptions{})
	checks = append(checks, err)
	_, err = graph.Radius(PathOptions{})
	checks = append(checks, err)
	_, err = graph.GraphCenter(PathOptions{})
	checks = append(checks, err)
	_, err = graph.PseudoDiameter(PseudoDiameterOptions{})
	checks = append(checks, err)
	_, err = graph.GlobalEfficiency(false, nil)
	checks = append(checks, err)
	_, err = graph.LocalEfficiencies(AllVertices(), false, DirectionAll, nil)
	checks = append(checks, err)
	_, err = graph.AverageLocalEfficiency(false, DirectionAll, nil)
	checks = append(checks, err)
	_, err = graph.PathLengthHistogram(false)
	checks = append(checks, err)
	for i, err := range checks {
		if !errors.Is(err, ErrClosed) {
			t.Errorf("closed check %d error = %v", i, err)
		}
	}
}

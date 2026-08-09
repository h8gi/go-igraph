package igraph

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestCalculateCentralizationRawNormalizedOwnershipAndValidation(t *testing.T) {
	scores := []float64{3, 1, 1, 1}
	raw, err := CalculateCentralization(scores, 6, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(raw.Scores, scores) || raw.Value != 6 || raw.TheoreticalMaximum != 6 || raw.Normalized {
		t.Errorf("raw centralization = %#v", raw)
	}
	normalized, err := CalculateCentralization(scores, 6, true)
	if err != nil {
		t.Fatal(err)
	}
	if normalized.Value != 1 || !normalized.Normalized {
		t.Errorf("normalized centralization = %#v", normalized)
	}
	scores[0] = 0
	if !reflect.DeepEqual(raw.Scores, []float64{3, 1, 1, 1}) {
		t.Errorf("result retained input storage: %v", raw.Scores)
	}
	empty, err := CalculateCentralization(nil, 0, false)
	if err != nil || empty.Scores == nil || len(empty.Scores) != 0 || !math.IsNaN(empty.Value) {
		t.Errorf("empty centralization = %#v, %v", empty, err)
	}

	invalidScores := [][]float64{{math.NaN()}, {math.Inf(1)}, {math.Inf(-1)}}
	for _, values := range invalidScores {
		if _, err := CalculateCentralization(values, 1, false); err == nil {
			t.Errorf("CalculateCentralization(%v) error = nil", values)
		}
	}
	for _, maximum := range []float64{-1, math.NaN(), math.Inf(1)} {
		if _, err := CalculateCentralization([]float64{1}, maximum, false); err == nil {
			t.Errorf("theoretical maximum %v error = nil", maximum)
		}
	}
	if _, err := CalculateCentralization([]float64{1}, 0, true); err == nil {
		t.Error("normalized zero theoretical maximum error = nil")
	}
}

func TestSpecializedCentralizationKnownAnswers(t *testing.T) {
	star, err := NewStar(4, 0, StarUndirected)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = star.Close() })
	degree, err := star.DegreeCentralization(DegreeCentralizationOptions{Normalized: true})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, degree.Scores, []float64{3, 1, 1, 1})
	assertCentralization(t, degree, 1, 6, true)
	betweenness, err := star.BetweennessCentralization(BetweennessCentralizationOptions{Normalized: true})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, betweenness.Scores, []float64{3, 0, 0, 0})
	assertCentralization(t, betweenness, 1, 9, true)
	closeness, err := star.ClosenessCentralization(ClosenessCentralizationOptions{Normalized: true})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, closeness.Scores, []float64{1, 0.6, 0.6, 0.6})
	assertCentralization(t, closeness, 1, 1.2, true)
	eigenvector, err := star.EigenvectorCentralization(EigenvectorCentralizationOptions{Normalized: true})
	if err != nil {
		t.Fatal(err)
	}
	leaf := 1 / math.Sqrt(3)
	assertFloatSlice(t, eigenvector.Scores, []float64{1, leaf, leaf, leaf})
	assertCentralization(t, eigenvector, 3*(1-leaf)/2, 2, true)

	path, err := NewPath(4, false, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = path.Close() })
	pathDegree, err := path.DegreeCentralization(DegreeCentralizationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, pathDegree.Scores, []float64{1, 2, 2, 1})
	assertCentralization(t, pathDegree, 2, 6, false)
	pathBetweenness, err := path.BetweennessCentralization(BetweennessCentralizationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, pathBetweenness.Scores, []float64{0, 2, 2, 0})
	assertCentralization(t, pathBetweenness, 4, 9, false)

	complete, err := NewFull(4, false, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = complete.Close() })
	for name, calculate := range map[string]func() (CentralizationResult, error){
		"degree": func() (CentralizationResult, error) {
			return complete.DegreeCentralization(DegreeCentralizationOptions{})
		},
		"betweenness": func() (CentralizationResult, error) {
			return complete.BetweennessCentralization(BetweennessCentralizationOptions{})
		},
		"closeness": func() (CentralizationResult, error) {
			return complete.ClosenessCentralization(ClosenessCentralizationOptions{})
		},
		"eigenvector": func() (CentralizationResult, error) {
			return complete.EigenvectorCentralization(EigenvectorCentralizationOptions{})
		},
	} {
		t.Run(name, func(t *testing.T) {
			result, err := calculate()
			if err != nil {
				t.Fatal(err)
			}
			assertFloat(t, result.Value, 0)
		})
	}
}

func TestCentralizationScoresAgreeWithNodeAPIs(t *testing.T) {
	graph, err := NewGraphFromEdges(5, []Edge{{0, 1}, {0, 2}, {1, 2}, {2, 3}}, true)
	if err != nil {
		t.Fatal(err)
	}

	degree, err := graph.DegreeCentralization(DegreeCentralizationOptions{Direction: DirectionIn})
	if err != nil {
		t.Fatal(err)
	}
	degreeScores, err := graph.Degree(AllVertices(), DegreeOptions{Direction: DirectionIn})
	if err != nil {
		t.Fatal(err)
	}
	for index, score := range degreeScores {
		assertFloat(t, degree.Scores[index], float64(score))
	}
	betweenness, err := graph.BetweennessCentralization(BetweennessCentralizationOptions{DirectedPaths: true})
	if err != nil {
		t.Fatal(err)
	}
	vertexBetweenness, err := graph.VertexBetweenness(AllVertices(), BetweennessOptions{DirectedPaths: true})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, betweenness.Scores, vertexBetweenness)
	closeness, err := graph.ClosenessCentralization(ClosenessCentralizationOptions{Direction: DirectionOut})
	if err != nil {
		t.Fatal(err)
	}
	vertexCloseness, err := graph.Closeness(AllVertices(), DistanceCentralityOptions{Direction: DirectionOut, Normalized: true})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, closeness.Scores, vertexCloseness.Scores)
	eigenvector, err := graph.EigenvectorCentralization(EigenvectorCentralizationOptions{Direction: DirectionAll})
	if err != nil {
		t.Fatal(err)
	}
	vertexEigenvector, err := graph.EigenvectorCentrality(EigenvectorCentralityOptions{Direction: DirectionAll})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, eigenvector.Scores, vertexEigenvector.Scores)

	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, betweenness.Scores, vertexBetweenness)
	assertFloatSlice(t, closeness.Scores, vertexCloseness.Scores)
}

func TestCentralizationDirectedModesLoopsDisconnectedAndDegenerate(t *testing.T) {
	directed, err := NewStar(4, 0, StarOut)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = directed.Close() })
	out, err := directed.DegreeCentralization(DegreeCentralizationOptions{Direction: DirectionOut, Normalized: true})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, out.Scores, []float64{3, 0, 0, 0})
	assertFloat(t, out.Value, 1)
	in, err := directed.DegreeCentralization(DegreeCentralizationOptions{Direction: DirectionIn})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, in.Scores, []float64{0, 1, 1, 1})
	assertFloat(t, in.Value, 1)

	withLoop, err := NewGraphFromEdges(3, []Edge{{0, 1}, {0, 2}, {0, 0}}, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = withLoop.Close() })
	withoutLoops, err := withLoop.DegreeCentralization(DegreeCentralizationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	withLoops, err := withLoop.DegreeCentralization(DegreeCentralizationOptions{CountLoops: true})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, withoutLoops.Scores, []float64{2, 1, 1})
	assertFloatSlice(t, withLoops.Scores, []float64{4, 1, 1})

	disconnected, err := NewGraphFromEdges(4, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = disconnected.Close() })
	disconnectedCloseness, err := disconnected.ClosenessCentralization(ClosenessCentralizationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsNaN(disconnectedCloseness.Value) {
		t.Errorf("disconnected closeness centralization = %v, want NaN", disconnectedCloseness.Value)
	}

	single, err := NewGraphFromEdges(1, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = single.Close() })
	for name, calculate := range map[string]func(bool) (CentralizationResult, error){
		"degree": func(normalized bool) (CentralizationResult, error) {
			return single.DegreeCentralization(DegreeCentralizationOptions{Normalized: normalized})
		},
		"betweenness": func(normalized bool) (CentralizationResult, error) {
			return single.BetweennessCentralization(BetweennessCentralizationOptions{Normalized: normalized})
		},
		"closeness": func(normalized bool) (CentralizationResult, error) {
			return single.ClosenessCentralization(ClosenessCentralizationOptions{Normalized: normalized})
		},
		"eigenvector": func(normalized bool) (CentralizationResult, error) {
			return single.EigenvectorCentralization(EigenvectorCentralizationOptions{Normalized: normalized})
		},
	} {
		t.Run(name, func(t *testing.T) {
			raw, err := calculate(false)
			if err != nil || raw.Scores == nil || len(raw.Scores) != 1 || raw.TheoreticalMaximum != 0 {
				t.Errorf("single raw = %#v, %v", raw, err)
			}
			if _, err := calculate(true); err == nil {
				t.Error("single normalized centralization error = nil")
			}
		})
	}

	empty, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = empty.Close() })
	for name, calculate := range map[string]func(bool) (CentralizationResult, error){
		"degree": func(normalized bool) (CentralizationResult, error) {
			return empty.DegreeCentralization(DegreeCentralizationOptions{Normalized: normalized})
		},
		"betweenness": func(normalized bool) (CentralizationResult, error) {
			return empty.BetweennessCentralization(BetweennessCentralizationOptions{Normalized: normalized})
		},
		"closeness": func(normalized bool) (CentralizationResult, error) {
			return empty.ClosenessCentralization(ClosenessCentralizationOptions{Normalized: normalized})
		},
		"eigenvector": func(normalized bool) (CentralizationResult, error) {
			return empty.EigenvectorCentralization(EigenvectorCentralizationOptions{Normalized: normalized})
		},
	} {
		t.Run("empty "+name, func(t *testing.T) {
			raw, err := calculate(false)
			if err != nil || raw.Scores == nil || len(raw.Scores) != 0 || raw.TheoreticalMaximum != 0 {
				t.Errorf("empty raw = %#v, %v", raw, err)
			}
			if _, err := calculate(true); err == nil {
				t.Error("empty normalized centralization error = nil")
			}
		})
	}
}

func TestCentralizationRejectsInvalidOptionsAndClosedGraph(t *testing.T) {
	graph, err := NewPath(3, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := graph.DegreeCentralization(DegreeCentralizationOptions{Direction: DirectionMode(99)}); err == nil {
		t.Error("invalid degree direction error = nil")
	}
	if _, err := graph.ClosenessCentralization(ClosenessCentralizationOptions{Direction: DirectionMode(99)}); err == nil {
		t.Error("invalid closeness direction error = nil")
	}
	if _, err := graph.EigenvectorCentralization(EigenvectorCentralizationOptions{Direction: DirectionMode(99)}); err == nil {
		t.Error("invalid eigenvector direction error = nil")
	}
	invalidSolvers := []SpectralSolverOptions{{MaxIterations: -1}, {Tolerance: -1}, {Tolerance: math.NaN()}}
	for _, solver := range invalidSolvers {
		if _, err := graph.EigenvectorCentralization(EigenvectorCentralizationOptions{Solver: solver}); err == nil {
			t.Errorf("invalid eigenvector solver %#v error = nil", solver)
		}
	}

	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := graph.DegreeCentralization(DegreeCentralizationOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("DegreeCentralization after Close error = %v", err)
	}
	if _, err := graph.BetweennessCentralization(BetweennessCentralizationOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("BetweennessCentralization after Close error = %v", err)
	}
	if _, err := graph.ClosenessCentralization(ClosenessCentralizationOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("ClosenessCentralization after Close error = %v", err)
	}
	if _, err := graph.EigenvectorCentralization(EigenvectorCentralizationOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("EigenvectorCentralization after Close error = %v", err)
	}
	var nilGraph *Graph
	if _, err := nilGraph.DegreeCentralization(DegreeCentralizationOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("nil DegreeCentralization error = %v", err)
	}
	if _, err := nilGraph.BetweennessCentralization(BetweennessCentralizationOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("nil BetweennessCentralization error = %v", err)
	}
	if _, err := nilGraph.ClosenessCentralization(ClosenessCentralizationOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("nil ClosenessCentralization error = %v", err)
	}
	if _, err := nilGraph.EigenvectorCentralization(EigenvectorCentralizationOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("nil EigenvectorCentralization error = %v", err)
	}
}

func assertCentralization(t *testing.T, result CentralizationResult, value, maximum float64, normalized bool) {
	t.Helper()
	assertFloat(t, result.Value, value)
	assertFloat(t, result.TheoreticalMaximum, maximum)
	if result.Normalized != normalized {
		t.Errorf("Normalized = %v, want %v", result.Normalized, normalized)
	}
}

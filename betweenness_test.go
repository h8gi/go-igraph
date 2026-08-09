package igraph

import (
	"errors"
	"math"
	"strings"
	"testing"
)

func TestBetweennessPathKnownAnswersNormalizationAndSelectorOrder(t *testing.T) {
	graph, err := NewPath(4, false, false)
	if err != nil {
		t.Fatal(err)
	}
	vertices, _ := VertexIDs(2, 0, 2, 1)
	edges, _ := EdgeIDs(2, 0, 2, 1)

	vertexScores, err := graph.VertexBetweenness(vertices, BetweennessOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, vertexScores, []float64{2, 0, 2, 2})
	vertexNormalized, err := graph.VertexBetweenness(vertices, BetweennessOptions{Normalized: true})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, vertexNormalized, []float64{1.0 / 3, 0, 1.0 / 3, 1.0 / 3})

	edgeScores, err := graph.EdgeBetweenness(edges, BetweennessOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, edgeScores, []float64{3, 3, 3, 4})
	edgeNormalized, err := graph.EdgeBetweenness(edges, BetweennessOptions{Normalized: true})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, edgeNormalized, []float64{0.5, 0.5, 0.5, 2.0 / 3})

	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, vertexScores, []float64{2, 0, 2, 2})
	assertFloatSlice(t, edgeScores, []float64{3, 3, 3, 4})
}

func TestBetweennessDirectedPathsWeightsCutoffParallelEdgesAndLoops(t *testing.T) {
	t.Run("directed paths", func(t *testing.T) {
		graph, err := NewGraphFromEdges(3, []Edge{{0, 1}, {2, 1}}, true)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = graph.Close() })
		directed, err := graph.VertexBetweenness(AllVertices(), BetweennessOptions{DirectedPaths: true})
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSlice(t, directed, []float64{0, 0, 0})
		undirected, err := graph.VertexBetweenness(AllVertices(), BetweennessOptions{})
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSlice(t, undirected, []float64{0, 1, 0})
		directedEdges, err := graph.EdgeBetweenness(AllEdges(), BetweennessOptions{DirectedPaths: true})
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSlice(t, directedEdges, []float64{1, 1})
		undirectedEdges, err := graph.EdgeBetweenness(AllEdges(), BetweennessOptions{})
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSlice(t, undirectedEdges, []float64{2, 2})
	})

	t.Run("weighted", func(t *testing.T) {
		graph, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}, {0, 2}}, false)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = graph.Close() })
		options := BetweennessOptions{Weights: []float64{1, 1, 3}}
		vertices, _ := VertexIDs(1, 0, 1)
		vertexScores, err := graph.VertexBetweenness(vertices, options)
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSlice(t, vertexScores, []float64{1, 0, 1})
		edges, _ := EdgeIDs(2, 0, 1, 0)
		edgeScores, err := graph.EdgeBetweenness(edges, options)
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSlice(t, edgeScores, []float64{0, 2, 2, 2})
	})

	t.Run("cutoff", func(t *testing.T) {
		graph, err := NewPath(4, false, false)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = graph.Close() })
		one := 1.0
		vertices, err := graph.VertexBetweenness(AllVertices(), BetweennessOptions{Cutoff: &one})
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSlice(t, vertices, []float64{0, 0, 0, 0})
		edges, err := graph.EdgeBetweenness(AllEdges(), BetweennessOptions{Cutoff: &one})
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSlice(t, edges, []float64{1, 1, 1})
		two := 2.0
		vertices, err = graph.VertexBetweenness(AllVertices(), BetweennessOptions{Cutoff: &two})
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSlice(t, vertices, []float64{0, 1, 1, 0})
		edges, err = graph.EdgeBetweenness(AllEdges(), BetweennessOptions{Cutoff: &two})
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSlice(t, edges, []float64{2, 3, 2})
	})

	t.Run("parallel edges and loop", func(t *testing.T) {
		parallel, err := NewGraphFromEdges(3, []Edge{{0, 1}, {0, 1}, {1, 2}}, false)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = parallel.Close() })
		scores, err := parallel.EdgeBetweenness(AllEdges(), BetweennessOptions{})
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSlice(t, scores, []float64{1, 1, 2})

		withLoop, err := NewGraphFromEdges(2, []Edge{{0, 0}, {0, 1}}, false)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = withLoop.Close() })
		loopScores, err := withLoop.EdgeBetweenness(AllEdges(), BetweennessOptions{})
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSlice(t, loopScores, []float64{0, 1})
	})
}

func TestBetweennessEdgePairsEmptyAndDegenerateGraphs(t *testing.T) {
	graph, err := NewPath(3, false, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	pairs, _ := EdgePairs([]Edge{{1, 2}, {0, 1}, {1, 2}}, false)
	scores, err := graph.EdgeBetweenness(pairs, BetweennessOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, scores, []float64{2, 2, 2})

	emptyVertices, err := graph.VertexBetweenness(NoVertices(), BetweennessOptions{})
	if err != nil || emptyVertices == nil || len(emptyVertices) != 0 {
		t.Errorf("VertexBetweenness(NoVertices) = %#v, %v", emptyVertices, err)
	}
	emptyEdges, err := graph.EdgeBetweenness(NoEdges(), BetweennessOptions{})
	if err != nil || emptyEdges == nil || len(emptyEdges) != 0 {
		t.Errorf("EdgeBetweenness(NoEdges) = %#v, %v", emptyEdges, err)
	}

	emptyGraph, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = emptyGraph.Close() })
	emptyVertices, err = emptyGraph.VertexBetweenness(AllVertices(), BetweennessOptions{})
	if err != nil || emptyVertices == nil || len(emptyVertices) != 0 {
		t.Errorf("empty graph vertex betweenness = %#v, %v", emptyVertices, err)
	}
	emptyEdges, err = emptyGraph.EdgeBetweenness(AllEdges(), BetweennessOptions{})
	if err != nil || emptyEdges == nil || len(emptyEdges) != 0 {
		t.Errorf("empty graph edge betweenness = %#v, %v", emptyEdges, err)
	}

	single, err := NewGraphFromEdges(1, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = single.Close() })
	singleScores, err := single.VertexBetweenness(AllVertices(), BetweennessOptions{Normalized: true})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, singleScores, []float64{math.NaN()})
}

func TestBetweennessRejectsInvalidInputsAndClosedGraph(t *testing.T) {
	graph, err := NewPath(3, false, false)
	if err != nil {
		t.Fatal(err)
	}
	invalidVertices := VertexSelector{kind: vertexSelectorIDs, ids: []int{3}}
	if _, err := graph.VertexBetweenness(invalidVertices, BetweennessOptions{}); err == nil || !strings.Contains(err.Error(), "selector") {
		t.Errorf("invalid vertex selector error = %v", err)
	}
	invalidEdges := EdgeSelector{kind: edgeSelectorIDs, ids: []int{2}}
	if _, err := graph.EdgeBetweenness(invalidEdges, BetweennessOptions{}); err == nil || !strings.Contains(err.Error(), "selector") {
		t.Errorf("invalid edge selector error = %v", err)
	}
	missingPair, _ := EdgePairs([]Edge{{0, 2}}, false)
	if _, err := graph.EdgeBetweenness(missingPair, BetweennessOptions{}); err == nil {
		t.Error("missing edge pair error = nil")
	}
	if _, err := graph.VertexBetweenness(VertexSelector{kind: vertexSelectorKind(99)}, BetweennessOptions{}); err == nil {
		t.Error("invalid vertex selector kind error = nil")
	}
	if _, err := graph.EdgeBetweenness(EdgeSelector{kind: edgeSelectorKind(99)}, BetweennessOptions{}); err == nil {
		t.Error("invalid edge selector kind error = nil")
	}
	badPair := EdgeSelector{kind: edgeSelectorPairs, pairs: []Edge{{From: 0, To: 3}}}
	if _, err := graph.EdgeBetweenness(badPair, BetweennessOptions{}); err == nil {
		t.Error("out-of-range edge pair error = nil")
	}

	invalidCutoffs := []float64{-1, math.NaN(), math.Inf(1)}
	for _, cutoff := range invalidCutoffs {
		options := BetweennessOptions{Cutoff: &cutoff}
		if _, err := graph.VertexBetweenness(AllVertices(), options); err == nil {
			t.Errorf("vertex cutoff %v error = nil", cutoff)
		}
		if _, err := graph.EdgeBetweenness(AllEdges(), options); err == nil {
			t.Errorf("edge cutoff %v error = nil", cutoff)
		}
	}
	invalidOptions := []BetweennessOptions{
		{Weights: []float64{}},
		{Weights: []float64{1}},
		{Weights: []float64{1, 0}},
		{Weights: []float64{1, -1}},
		{Weights: []float64{1, math.NaN()}},
		{Weights: []float64{1, math.Inf(1)}},
	}
	for _, options := range invalidOptions {
		if _, err := graph.VertexBetweenness(AllVertices(), options); err == nil {
			t.Errorf("VertexBetweenness(%#v) error = nil", options)
		}
		if _, err := graph.EdgeBetweenness(AllEdges(), options); err == nil {
			t.Errorf("EdgeBetweenness(%#v) error = nil", options)
		}
	}

	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := graph.VertexBetweenness(AllVertices(), BetweennessOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("VertexBetweenness after Close error = %v", err)
	}
	if _, err := graph.EdgeBetweenness(AllEdges(), BetweennessOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("EdgeBetweenness after Close error = %v", err)
	}
	var nilGraph *Graph
	if _, err := nilGraph.VertexBetweenness(AllVertices(), BetweennessOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("nil VertexBetweenness error = %v", err)
	}
	if _, err := nilGraph.EdgeBetweenness(AllEdges(), BetweennessOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("nil EdgeBetweenness error = %v", err)
	}
}

func TestCollectBetweennessPropagatesInitializationAndUpstreamErrors(t *testing.T) {
	forced := errors.New("forced initialization failure")
	values, err := collectBetweennessWithInitializer(
		"test betweenness",
		func(*realVector) int { return 0 },
		func(int) (*realVector, error) { return nil, forced },
	)
	if values != nil || !errors.Is(err, forced) {
		t.Errorf("initialization failure = %#v, %v", values, err)
	}
	values, err = collectBetweenness("test betweenness", func(*realVector) int {
		return 2 // IGRAPH_ENOMEM
	})
	if values != nil || err == nil || !strings.Contains(err.Error(), "test betweenness") {
		t.Errorf("upstream failure = %#v, %v", values, err)
	}
}

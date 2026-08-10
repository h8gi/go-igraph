package igraph

import (
	"errors"
	"math"
	"strings"
	"testing"
)

func TestEigenvectorCentralityKnownAnswersWeightsAndOwnership(t *testing.T) {
	path, err := NewPath(3, false, false)
	if err != nil {
		t.Fatal(err)
	}
	result, err := path.EigenvectorCentrality(EigenvectorCentralityOptions{
		Solver: SpectralSolverOptions{MaxIterations: 1000, Tolerance: 1e-12},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, result.Scores, []float64{math.Sqrt(0.5), 1, math.Sqrt(0.5)})
	assertFloat(t, result.Eigenvalue, math.Sqrt(2))
	if err := path.Close(); err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, result.Scores, []float64{math.Sqrt(0.5), 1, math.Sqrt(0.5)})

	weighted, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = weighted.Close() })
	weightedResult, err := weighted.EigenvectorCentrality(EigenvectorCentralityOptions{
		Weights: []float64{1, 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, weightedResult.Scores, []float64{1 / math.Sqrt(5), 1, 2 / math.Sqrt(5)})
	assertFloat(t, weightedResult.Eigenvalue, math.Sqrt(5))

	zeroWeight, err := weighted.EigenvectorCentrality(EigenvectorCentralityOptions{Weights: []float64{0, 0}})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, zeroWeight.Scores, []float64{1, 1, 1})
	assertFloat(t, zeroWeight.Eigenvalue, 0)
}

func TestEigenvectorCentralityDirectedModes(t *testing.T) {
	graph, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 0}, {1, 2}}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	incoming, err := graph.EigenvectorCentrality(EigenvectorCentralityOptions{Direction: DirectionIn})
	if err != nil {
		t.Fatal(err)
	}
	outgoing, err := graph.EigenvectorCentrality(EigenvectorCentralityOptions{Direction: DirectionOut})
	if err != nil {
		t.Fatal(err)
	}
	if !(incoming.Scores[2] < outgoing.Scores[2]) {
		t.Errorf("direction modes did not change sink score: in=%v out=%v", incoming.Scores, outgoing.Scores)
	}
	all, err := graph.EigenvectorCentrality(EigenvectorCentralityOptions{Direction: DirectionAll})
	if err != nil {
		t.Fatal(err)
	}
	if len(all.Scores) != 3 {
		t.Errorf("DirectionAll scores = %v", all.Scores)
	}
}

func TestHITSKnownAnswersAndWeights(t *testing.T) {
	star, err := NewStar(4, 0, StarOut)
	if err != nil {
		t.Fatal(err)
	}
	result, err := star.HITS(HITSOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, result.HubScores, []float64{1, 0, 0, 0})
	assertFloatSlice(t, result.AuthorityScores, []float64{0, 1, 1, 1})
	assertFloat(t, result.Eigenvalue, 3)

	weighted, err := star.HITS(HITSOptions{Weights: []float64{1, 2, 3}})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, weighted.HubScores, []float64{1, 0, 0, 0})
	assertFloatSlice(t, weighted.AuthorityScores, []float64{0, 1.0 / 3, 2.0 / 3, 1})
	assertFloat(t, weighted.Eigenvalue, 14)
	zeroWeight, err := star.HITS(HITSOptions{Weights: []float64{0, 0, 0}})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, zeroWeight.HubScores, []float64{1, 1, 1, 1})
	assertFloatSlice(t, zeroWeight.AuthorityScores, []float64{1, 1, 1, 1})
	assertFloat(t, zeroWeight.Eigenvalue, 0)
	if err := star.Close(); err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, result.HubScores, []float64{1, 0, 0, 0})
	assertFloatSlice(t, result.AuthorityScores, []float64{0, 1, 1, 1})
}

func TestPageRankUniformSelectorOrderAlgorithmsAndDirection(t *testing.T) {
	ring, err := NewRing(4, true, false)
	if err != nil {
		t.Fatal(err)
	}
	selector, _ := VertexIDs(2, 0, 2, 3)
	for _, algorithm := range []PageRankAlgorithm{PageRankPRPACK, PageRankARPACK} {
		result, err := ring.PageRank(selector, PageRankOptions{Algorithm: algorithm})
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSlice(t, result.Scores, []float64{0.25, 0.25, 0.25, 0.25})
		assertFloat(t, result.Eigenvalue, 1)
	}
	owned, err := ring.PageRank(AllVertices(), PageRankOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := ring.Close(); err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, owned.Scores, []float64{0.25, 0.25, 0.25, 0.25})

	star, err := NewStar(3, 0, StarOut)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = star.Close() })
	respect, err := star.PageRank(AllVertices(), PageRankOptions{Direction: PageRankRespectDirections})
	if err != nil {
		t.Fatal(err)
	}
	ignore, err := star.PageRank(AllVertices(), PageRankOptions{Direction: PageRankIgnoreDirections})
	if err != nil {
		t.Fatal(err)
	}
	if !(respect.Scores[0] < respect.Scores[1] && ignore.Scores[0] > ignore.Scores[1]) {
		t.Errorf("PageRank direction results: respect=%v ignore=%v", respect.Scores, ignore.Scores)
	}
}

func TestPersonalizedPageRankDistributionVerticesDampingAndWeights(t *testing.T) {
	graph, err := NewGraphFromEdges(3, []Edge{{0, 1}, {0, 2}}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	zero := 0.0
	distribution := []float64{1, 0, 0}
	fromDistribution, err := graph.PageRank(AllVertices(), PageRankOptions{
		Damping: &zero, ResetDistribution: distribution,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, fromDistribution.Scores, []float64{1, 0, 0})
	distribution[0] = 0
	assertFloatSlice(t, fromDistribution.Scores, []float64{1, 0, 0})

	resetVertices, _ := VertexIDs(1, 1, 2)
	fromVertices, err := graph.PageRank(AllVertices(), PageRankOptions{
		Damping: &zero, ResetVertices: &resetVertices,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, fromVertices.Scores, []float64{0, 0.5, 0.5})

	weighted, err := graph.PageRank(AllVertices(), PageRankOptions{Weights: []float64{1, 3}})
	if err != nil {
		t.Fatal(err)
	}
	if !(weighted.Scores[2] > weighted.Scores[1]) {
		t.Errorf("weighted PageRank = %v, want vertex 2 > vertex 1", weighted.Scores)
	}
	zeroWeights, err := graph.PageRank(AllVertices(), PageRankOptions{Weights: []float64{0, 0}})
	if err != nil {
		t.Fatal(err)
	}
	assertProbabilityScores(t, zeroWeights.Scores)
}

func TestSpectralAndPageRankEmptySelectionsAndDegenerateGraphs(t *testing.T) {
	graph, err := NewGraphFromEdges(1, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	pageRank, err := graph.PageRank(NoVertices(), PageRankOptions{})
	if err != nil || pageRank.Scores == nil || len(pageRank.Scores) != 0 {
		t.Errorf("PageRank(NoVertices) = %#v, %v", pageRank, err)
	}
	eigenvector, err := graph.EigenvectorCentrality(EigenvectorCentralityOptions{})
	if err != nil || eigenvector.Scores == nil || len(eigenvector.Scores) != 1 {
		t.Errorf("single-vertex eigenvector = %#v, %v", eigenvector, err)
	}
	assertFloatSlice(t, eigenvector.Scores, []float64{1})
	assertFloat(t, eigenvector.Eigenvalue, 0)
	hits, err := graph.HITS(HITSOptions{})
	if err != nil || hits.HubScores == nil || hits.AuthorityScores == nil {
		t.Errorf("single-vertex HITS = %#v, %v", hits, err)
	}
	assertFloatSlice(t, hits.HubScores, []float64{1})
	assertFloatSlice(t, hits.AuthorityScores, []float64{1})
	assertFloat(t, hits.Eigenvalue, 0)

	empty, err := NewGraph()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = empty.Close() })
	emptyPageRank, err := empty.PageRank(AllVertices(), PageRankOptions{})
	if err != nil || emptyPageRank.Scores == nil || len(emptyPageRank.Scores) != 0 {
		t.Errorf("empty PageRank = %#v, %v", emptyPageRank, err)
	}
	emptyEigenvector, eigenErr := empty.EigenvectorCentrality(EigenvectorCentralityOptions{})
	if eigenErr != nil || emptyEigenvector.Scores == nil || len(emptyEigenvector.Scores) != 0 || emptyEigenvector.Eigenvalue != 0 {
		t.Errorf("empty eigenvector = %#v, %v", emptyEigenvector, eigenErr)
	}
	emptyHITS, hitsErr := empty.HITS(HITSOptions{})
	if hitsErr != nil || emptyHITS.HubScores == nil || emptyHITS.AuthorityScores == nil ||
		len(emptyHITS.HubScores) != 0 || len(emptyHITS.AuthorityScores) != 0 || emptyHITS.Eigenvalue != 0 {
		t.Errorf("empty HITS = %#v, %v", emptyHITS, hitsErr)
	}
}

func TestSpectralAndPageRankRejectInvalidOptionsAndClosedGraph(t *testing.T) {
	graph, err := NewPath(3, true, false)
	if err != nil {
		t.Fatal(err)
	}
	invalidSolvers := []SpectralSolverOptions{
		{MaxIterations: -1},
		{Tolerance: -1},
		{Tolerance: math.NaN()},
		{Tolerance: math.Inf(1)},
	}
	for _, solver := range invalidSolvers {
		if _, err := graph.EigenvectorCentrality(EigenvectorCentralityOptions{Solver: solver}); err == nil {
			t.Errorf("EigenvectorCentrality solver %#v error = nil", solver)
		}
		if _, err := graph.HITS(HITSOptions{Solver: solver}); err == nil {
			t.Errorf("HITS solver %#v error = nil", solver)
		}
		if _, err := graph.PageRank(AllVertices(), PageRankOptions{Solver: solver}); err == nil {
			t.Errorf("PageRank solver %#v error = nil", solver)
		}
	}
	if _, err := graph.EigenvectorCentrality(EigenvectorCentralityOptions{Direction: DirectionMode(99)}); err == nil {
		t.Error("invalid eigenvector direction error = nil")
	}
	if _, err := graph.PageRank(AllVertices(), PageRankOptions{Direction: PageRankDirection(99)}); err == nil {
		t.Error("invalid PageRank direction error = nil")
	}
	if _, err := graph.PageRank(AllVertices(), PageRankOptions{Algorithm: PageRankAlgorithm(99)}); err == nil {
		t.Error("invalid PageRank algorithm error = nil")
	}
	invalidWeights := [][]float64{
		{}, {1}, {1, -1}, {1, math.NaN()}, {1, math.Inf(1)},
	}
	for _, weights := range invalidWeights {
		if _, err := graph.EigenvectorCentrality(EigenvectorCentralityOptions{Weights: weights}); err == nil {
			t.Errorf("EigenvectorCentrality weights %v error = nil", weights)
		}
		if _, err := graph.HITS(HITSOptions{Weights: weights}); err == nil {
			t.Errorf("HITS weights %v error = nil", weights)
		}
		if _, err := graph.PageRank(AllVertices(), PageRankOptions{Weights: weights}); err == nil {
			t.Errorf("PageRank weights %v error = nil", weights)
		}
	}

	invalidDamping := []float64{-0.1, 1, math.NaN(), math.Inf(1)}
	for _, damping := range invalidDamping {
		if _, err := graph.PageRank(AllVertices(), PageRankOptions{Damping: &damping}); err == nil {
			t.Errorf("PageRank damping %v error = nil", damping)
		}
	}
	resetVertices := AllVertices()
	if _, err := graph.PageRank(AllVertices(), PageRankOptions{
		ResetDistribution: []float64{1, 0, 0}, ResetVertices: &resetVertices,
	}); err == nil {
		t.Error("mutually exclusive resets error = nil")
	}
	badResets := [][]float64{
		{}, {1}, {0, 0, 0}, {-1, 1, 1}, {math.NaN(), 1, 1},
		{math.Inf(1), 1, 1}, {math.MaxFloat64, math.MaxFloat64, 0},
	}
	for _, reset := range badResets {
		if _, err := graph.PageRank(AllVertices(), PageRankOptions{ResetDistribution: reset}); err == nil {
			t.Errorf("PageRank reset %v error = nil", reset)
		}
	}
	emptyReset := NoVertices()
	if _, err := graph.PageRank(AllVertices(), PageRankOptions{ResetVertices: &emptyReset}); err == nil {
		t.Error("empty reset selector error = nil")
	}
	invalidReset := VertexSelector{kind: vertexSelectorIDs, ids: []int{3}}
	if _, err := graph.PageRank(AllVertices(), PageRankOptions{ResetVertices: &invalidReset}); err == nil {
		t.Error("invalid reset selector error = nil")
	}
	invalidResult := VertexSelector{kind: vertexSelectorIDs, ids: []int{3}}
	if _, err := graph.PageRank(invalidResult, PageRankOptions{}); err == nil || !strings.Contains(err.Error(), "result selector") {
		t.Errorf("invalid result selector error = %v", err)
	}

	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := graph.EigenvectorCentrality(EigenvectorCentralityOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("EigenvectorCentrality after Close error = %v", err)
	}
	if _, err := graph.HITS(HITSOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("HITS after Close error = %v", err)
	}
	if _, err := graph.PageRank(AllVertices(), PageRankOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("PageRank after Close error = %v", err)
	}
	var nilGraph *Graph
	if _, err := nilGraph.EigenvectorCentrality(EigenvectorCentralityOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("nil EigenvectorCentrality error = %v", err)
	}
	if _, err := nilGraph.HITS(HITSOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("nil HITS error = %v", err)
	}
	if _, err := nilGraph.PageRank(AllVertices(), PageRankOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("nil PageRank error = %v", err)
	}
}

func assertProbabilityScores(t *testing.T, scores []float64) {
	t.Helper()
	if scores == nil {
		t.Fatal("probability scores are nil")
	}
	sum := 0.0
	for _, score := range scores {
		if math.IsNaN(score) || math.IsInf(score, 0) || score < 0 {
			t.Fatalf("invalid probability scores: %v", scores)
		}
		sum += score
	}
	if math.Abs(sum-1) > 1e-12 {
		t.Errorf("probability score sum = %.15g, want 1: %v", sum, scores)
	}
}

func TestPageRankDuplicateVertexIDsSelector(t *testing.T) {
	graph, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()

	selector, err := VertexIDs(0, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	res, err := graph.PageRank(selector, PageRankOptions{})
	if err != nil {
		t.Fatalf("PageRank with duplicate vertex IDs failed: %v", err)
	}
	if len(res.Scores) != 3 {
		t.Errorf("expected 3 scores for duplicate selector, got %d", len(res.Scores))
	}
}

package igraph

import (
	"math"
	"testing"
)

func TestDistancesInvalidDirectionMode(t *testing.T) {
	g, err := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	if _, err := g.Distances(AllVertices(), AllVertices(), PathOptions{Direction: DirectionMode(99)}); err == nil {
		t.Error("expected error for invalid DirectionMode in Distances")
	}

	if _, err := g.ShortestPath(0, 1, PathOptions{Direction: DirectionMode(99)}); err == nil {
		t.Error("expected error for invalid DirectionMode in ShortestPath")
	}
}

func TestDistancesDuplicateSelectors(t *testing.T) {
	g, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	sources, err := VertexIDs(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	targets, err := VertexIDs(1, 1)
	if err != nil {
		t.Fatal(err)
	}

	dist, err := g.Distances(sources, targets, PathOptions{})
	if err != nil {
		t.Fatalf("Distances with duplicate selectors failed: %v", err)
	}
	r, c := dist.Dims()
	if r != 2 || c != 2 {
		t.Errorf("expected 2x2 matrix for duplicate selectors, got %dx%d", r, c)
	}
}

func TestNewOptionalEdgeWeightsValidation(t *testing.T) {
	if vec, err := newOptionalEdgeWeights(nil, 5); err != nil || vec != nil {
		t.Errorf("expected nil vector for nil edge weights, got %v, %v", vec, err)
	}

	if _, err := newOptionalEdgeWeights([]float64{1.0}, 2); err == nil {
		t.Error("expected error for weight count mismatch")
	}

	if _, err := newOptionalEdgeWeights([]float64{math.NaN()}, 1); err == nil {
		t.Error("expected error for NaN edge weight")
	}
}

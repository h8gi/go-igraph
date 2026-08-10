package igraph

import (
	"math"
	"testing"
)

func TestCentralityInvalidDirectionMode(t *testing.T) {
	g, err := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	if _, err := g.Closeness(AllVertices(), DistanceCentralityOptions{Direction: DirectionMode(99)}); err == nil {
		t.Error("expected error for invalid DirectionMode in Closeness")
	}
	if _, err := g.HarmonicCentrality(AllVertices(), DistanceCentralityOptions{Direction: DirectionMode(99)}); err == nil {
		t.Error("expected error for invalid DirectionMode in HarmonicCentrality")
	}
}

func TestValidateCentralityCutoffAndWeights(t *testing.T) {
	nanCutoff := math.NaN()
	if _, _, err := validateCentralityCutoff(&nanCutoff); err == nil {
		t.Error("expected error for NaN cutoff")
	}

	infCutoff := math.Inf(1)
	if _, _, err := validateCentralityCutoff(&infCutoff); err == nil {
		t.Error("expected error for +Inf cutoff")
	}

	negCutoff := -1.0
	if _, _, err := validateCentralityCutoff(&negCutoff); err == nil {
		t.Error("expected error for negative cutoff")
	}

	if _, err := newOptionalPositiveEdgeWeights([]float64{0.0}, 1); err == nil {
		t.Error("expected error for 0 weight")
	}

	if _, err := newOptionalPositiveEdgeWeights([]float64{-1.0}, 1); err == nil {
		t.Error("expected error for negative weight")
	}
}

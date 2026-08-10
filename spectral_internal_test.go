package igraph

import (
	"testing"
)

func TestEigenvectorCentralityNegativeWeightError(t *testing.T) {
	g, err := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	if _, err := g.EigenvectorCentrality(EigenvectorCentralityOptions{Weights: []float64{-1.0}}); err == nil {
		t.Error("expected error for negative edge weight in EigenvectorCentrality")
	}

	if _, err := g.HITS(HITSOptions{Weights: []float64{-1.0}}); err == nil {
		t.Error("expected error for negative edge weight in HITS")
	}

	if _, err := g.PageRank(AllVertices(), PageRankOptions{Weights: []float64{-1.0}}); err == nil {
		t.Error("expected error for negative edge weight in PageRank")
	}
}

func TestPageRankMutuallyExclusiveResetOptions(t *testing.T) {
	g, err := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	resetV := AllVertices()
	_, err = g.PageRank(AllVertices(), PageRankOptions{
		ResetDistribution: []float64{0.5, 0.5},
		ResetVertices:     &resetV,
	})
	if err == nil {
		t.Error("expected error when both ResetDistribution and ResetVertices are specified in PageRank")
	}
}

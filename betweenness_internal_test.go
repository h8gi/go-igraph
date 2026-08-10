package igraph

import (
	"testing"
)

func TestBetweennessNonPositiveWeightsAndCutoffErrors(t *testing.T) {
	g, err := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	if _, err := g.VertexBetweenness(AllVertices(), BetweennessOptions{Weights: []float64{0.0}}); err == nil {
		t.Error("expected error for non-positive edge weights in VertexBetweenness")
	}

	cutoff := -1.0
	if _, err := g.VertexBetweenness(AllVertices(), BetweennessOptions{Cutoff: &cutoff}); err == nil {
		t.Error("expected error for negative cutoff in VertexBetweenness")
	}

	if _, err := g.EdgeBetweenness(AllEdges(), BetweennessOptions{Weights: []float64{0.0}}); err == nil {
		t.Error("expected error for non-positive edge weights in EdgeBetweenness")
	}

	if _, err := g.EdgeBetweenness(AllEdges(), BetweennessOptions{Cutoff: &cutoff}); err == nil {
		t.Error("expected error for negative cutoff in EdgeBetweenness")
	}
}

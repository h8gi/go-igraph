package igraph

import (
	"testing"
)

func TestCollectEdgeSubgraphFailures(t *testing.T) {
	// Fake query returns nil graph
	_, err := collectEdgeSubgraph(3, false, edgeSubgraphOperations{
		query: func() (*Graph, error) {
			return nil, nil
		},
		closeGraph: func(g *Graph) {},
	})
	if err == nil {
		t.Error("expected error when query returns nil graph")
	}

	// Fake query returns valid graph, but identityMapping returns mismatched OldToNew size
	g, err := NewGraphFromEdges(3, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}

	_, err = collectEdgeSubgraph(5, false, edgeSubgraphOperations{
		query: func() (*Graph, error) {
			return g, nil
		},
		identityMapping: func(int) (IDMapping, error) {
			return IDMapping{OldToNew: []int{0, 1}}, nil
		},
		closeGraph: func(graph *Graph) { _ = graph.Close() },
	})
	if err == nil {
		t.Error("expected error for mismatched mapping length in collectEdgeSubgraph")
	}
}

func TestDecomposeInvalidOptions(t *testing.T) {
	g, err := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	if _, err := g.Decompose(DecomposeOptions{Connectedness: ConnectednessMode(99)}); err == nil {
		t.Error("expected error for invalid ConnectednessMode in Decompose")
	}
	if _, err := g.Decompose(DecomposeOptions{MaximumComponents: -2}); err == nil {
		t.Error("expected error for negative MaximumComponents in Decompose")
	}
	if _, err := g.Decompose(DecomposeOptions{MinimumVertices: -2}); err == nil {
		t.Error("expected error for negative MinimumVertices in Decompose")
	}
}

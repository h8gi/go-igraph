package igraph

import (
	"testing"
)

func TestTransitivityLocalDuplicateSelectorIDs(t *testing.T) {
	g, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}, {2, 0}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	selector, err := VertexIDs(0, 0, 1)
	if err != nil {
		t.Fatal(err)
	}

	scores, err := g.LocalTransitivity(selector, TransitivityNaN)
	if err != nil {
		t.Fatalf("LocalTransitivity with duplicate selector IDs failed: %v", err)
	}
	if len(scores) != 3 {
		t.Errorf("expected 3 scores for duplicate selector IDs, got %d", len(scores))
	}
}

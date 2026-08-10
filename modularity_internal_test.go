package igraph

import (
	"testing"
)

func TestCorenessInvalidMode(t *testing.T) {
	g, err := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	if _, err := g.Coreness(NeiMode(99)); err == nil {
		t.Error("expected error for invalid NeiMode in Coreness")
	}
}

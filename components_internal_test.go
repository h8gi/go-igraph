package igraph

import (
	"testing"
)

func TestConnectedComponentsInvalidMode(t *testing.T) {
	g, err := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	if _, err := g.ConnectedComponents(ConnectednessMode(99)); err == nil {
		t.Error("expected error for invalid ConnectednessMode")
	}
}

package igraph

import (
	"testing"
)

func TestValidateEdgeRangeErrors(t *testing.T) {
	if err := validateEdge(Edge{From: -1, To: 0}, 2, 0); err == nil {
		t.Error("expected error for negative edge.From")
	}
	if err := validateEdge(Edge{From: 0, To: 2}, 2, 0); err == nil {
		t.Error("expected error for edge.To out of range")
	}
}

func TestWriteNilFileErrors(t *testing.T) {
	g, err := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	if err := g.WriteEdgeList(nil); err == nil {
		t.Error("expected error for WriteEdgeList(nil)")
	}
	if err := g.WriteGraphML(nil, false); err == nil {
		t.Error("expected error for WriteGraphML(nil)")
	}
}

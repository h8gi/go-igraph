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

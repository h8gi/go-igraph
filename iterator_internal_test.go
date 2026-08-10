package igraph

import (
	"testing"
)

func TestIteratorCreationFailureOnInvalidSelector(t *testing.T) {
	g, err := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	// Invalid vertex ID selector materialized against C graph
	invalidV, err := VertexIDs(99)
	if err == nil {
		_, errVit := materializeVertexIDs(&g.graph, invalidV)
		if errVit == nil {
			t.Error("expected error for invalid vertex ID iterator creation")
		}
	}

	// Invalid edge ID selector materialized against C graph
	invalidE, err := EdgeIDs(99)
	if err == nil {
		_, errEit := materializeEdgeIDs(&g.graph, invalidE)
		if errEit == nil {
			t.Error("expected error for invalid edge ID iterator creation")
		}
	}
}

func TestNewCSelectorErrorsOnInvalidKind(t *testing.T) {
	if _, err := newCVertexSelector(VertexSelector{kind: vertexSelectorKind(99)}); err == nil {
		t.Error("expected error for invalid vertex selector kind")
	}
	if _, err := newCEdgeSelector(EdgeSelector{kind: edgeSelectorKind(99)}); err == nil {
		t.Error("expected error for invalid edge selector kind")
	}
}

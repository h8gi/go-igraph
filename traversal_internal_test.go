package igraph

import (
	"testing"
)

func TestBFSValidationRootNotInRestriction(t *testing.T) {
	g, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	restriction, err := VertexIDs(0)
	if err != nil {
		t.Fatal(err)
	}

	_, err = g.BreadthFirstSearch(BFSOptions{
		Roots:       []int{0, 1},
		Restriction: restriction,
	})
	if err == nil {
		t.Error("expected error when root is not in restriction")
	}
}

func TestCloseIntVectorsDirect(t *testing.T) {
	vec, err := newIntVector([]int{0, 1})
	if err != nil {
		t.Fatal(err)
	}
	closeIntVectors([]*intVector{vec})
	closeIntVectors(nil)
}

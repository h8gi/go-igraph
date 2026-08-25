package igraph

import (
	"testing"
)

func TestIsBiconnectedPropagatesUpstreamError(t *testing.T) {
	g, err := NewGraphFromEdges(2, []Edge{{From: 0, To: 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	if _, err := g.isBiconnected(func() (bool, int) {
		return false, 4
	}); err == nil {
		t.Fatal("IsBiconnected upstream error not propagated")
	}
}

func TestValidateBiconnectedComponentsErrors(t *testing.T) {
	if err := validateBiconnectedComponents(BiconnectedComponents{ComponentEdges: nil}, 1, 1); err == nil {
		t.Error("expected error for nil ComponentEdges")
	}
	if err := validateBiconnectedComponents(BiconnectedComponents{ComponentEdges: [][]int{}, ComponentVertices: nil}, 1, 1); err == nil {
		t.Error("expected error for nil ComponentVertices")
	}
	if err := validateBiconnectedComponents(BiconnectedComponents{ComponentEdges: [][]int{}, ComponentVertices: [][]int{}, ArticulationPoints: nil}, 1, 1); err == nil {
		t.Error("expected error for nil ArticulationPoints")
	}

	// Mismatched count
	if err := validateBiconnectedComponents(BiconnectedComponents{
		Count:              2,
		ComponentEdges:     [][]int{{0}},
		ComponentVertices:  [][]int{{0, 1}},
		ArticulationPoints: []int{},
	}, 2, 1); err == nil {
		t.Error("expected error for mismatched count")
	}

	// Nil component edge or vertex collection inside slice
	if err := validateBiconnectedComponents(BiconnectedComponents{
		Count:              1,
		ComponentEdges:     [][]int{nil},
		ComponentVertices:  [][]int{{0, 1}},
		ArticulationPoints: []int{},
	}, 2, 1); err == nil {
		t.Error("expected error for nil inner component edges")
	}

	// Edge ID out of range
	if err := validateBiconnectedComponents(BiconnectedComponents{
		Count:              1,
		ComponentEdges:     [][]int{{-1}},
		ComponentVertices:  [][]int{{0, 1}},
		ArticulationPoints: []int{},
	}, 2, 1); err == nil {
		t.Error("expected error for edge ID out of range")
	}

	// Vertex ID out of range
	if err := validateBiconnectedComponents(BiconnectedComponents{
		Count:              1,
		ComponentEdges:     [][]int{{0}},
		ComponentVertices:  [][]int{{-1}},
		ArticulationPoints: []int{},
	}, 2, 1); err == nil {
		t.Error("expected error for vertex ID out of range")
	}

	// Articulation point out of range
	if err := validateBiconnectedComponents(BiconnectedComponents{
		Count:              1,
		ComponentEdges:     [][]int{{0}},
		ComponentVertices:  [][]int{{0, 1}},
		ArticulationPoints: []int{99},
	}, 2, 1); err == nil {
		t.Error("expected error for articulation point out of range")
	}
}

package igraph

import (
	"testing"
)

func TestDyadCensusInternal(t *testing.T) {
	g, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}, {2, 0}}, true)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}
	defer g.Close()

	result, err := g.DyadCensus()
	if err != nil {
		t.Fatalf("DyadCensus() error = %v", err)
	}

	// 3 vertices: C(3,2) = 3 dyads, all asymmetric
	if result.Asymmetric != 3 {
		t.Errorf("Asymmetric = %d, want 3", result.Asymmetric)
	}
}

func TestTriadCensusInternal(t *testing.T) {
	g, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}, {2, 0}}, false)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}
	defer g.Close()

	result, err := g.TriadCensus()
	if err != nil {
		t.Fatalf("TriadCensus() error = %v", err)
	}

	if len(result) != 16 {
		t.Fatalf("result length = %d, want 16", len(result))
	}

	sum := int64(0)
	for _, v := range result {
		sum += v
	}
	if sum != 1 {
		t.Errorf("total count = %d, want 1", sum)
	}
}

func TestTrianglesCountInternal(t *testing.T) {
	edges := []Edge{
		{0, 1}, {0, 2}, {0, 3},
		{1, 2}, {1, 3},
		{2, 3},
	}
	g, err := NewGraphFromEdges(4, edges, false)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}
	defer g.Close()

	count, err := g.TrianglesCount()
	if err != nil {
		t.Fatalf("TrianglesCount() error = %v", err)
	}

	if count != 4 {
		t.Errorf("count = %d, want 4", count)
	}
}

func TestTrianglesListInternal(t *testing.T) {
	g, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}, {2, 0}}, false)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}
	defer g.Close()

	triangles, err := g.TrianglesList()
	if err != nil {
		t.Fatalf("TrianglesList() error = %v", err)
	}

	if len(triangles) != 1 {
		t.Fatalf("expected 1 triangle, got %d", len(triangles))
	}

	if triangles[0] != [3]int{0, 1, 2} {
		t.Errorf("triangle = %v, want [0,1,2]", triangles[0])
	}
}

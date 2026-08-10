package igraph_test

import (
	"errors"
	"testing"

	"github.com/h8gi/go-igraph"
)

func TestConnectivityAndDisjointPaths(t *testing.T) {
	// Create two triangles connected by a bridge 2-3
	// Vertices: 0, 1, 2 (triangle 1), 3, 4, 5 (triangle 2)
	edges := []igraph.Edge{
		{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0},
		{From: 2, To: 3}, // bridge
		{From: 3, To: 4}, {From: 4, To: 5}, {From: 5, To: 3},
	}
	g, err := igraph.NewGraphFromEdges(6, edges, false)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}
	defer g.Close()

	t.Run("EdgeConnectivity and Adhesion", func(t *testing.T) {
		ec, err := g.EdgeConnectivity(true)
		if err != nil {
			t.Fatalf("EdgeConnectivity failed: %v", err)
		}
		if ec != 1 {
			t.Errorf("expected edge connectivity 1 (bridge), got %d", ec)
		}

		adh, err := g.Adhesion(true)
		if err != nil {
			t.Fatalf("Adhesion failed: %v", err)
		}
		if adh != 1 {
			t.Errorf("expected adhesion 1, got %d", adh)
		}
	})

	t.Run("VertexConnectivity and Cohesion", func(t *testing.T) {
		vc, err := g.VertexConnectivity(true)
		if err != nil {
			t.Fatalf("VertexConnectivity failed: %v", err)
		}
		if vc != 1 {
			t.Errorf("expected vertex connectivity 1, got %d", vc)
		}

		coh, err := g.Cohesion(true)
		if err != nil {
			t.Fatalf("Cohesion failed: %v", err)
		}
		if coh != 1 {
			t.Errorf("expected cohesion 1, got %d", coh)
		}
	})

	t.Run("STEdgeConnectivity and STVertexConnectivity", func(t *testing.T) {
		stEC, err := g.STEdgeConnectivity(0, 5)
		if err != nil {
			t.Fatalf("STEdgeConnectivity failed: %v", err)
		}
		if stEC != 1 {
			t.Errorf("expected STEdgeConnectivity 1, got %d", stEC)
		}

		stVC, err := g.STVertexConnectivity(0, 5, igraph.VConnNeigError)
		if err != nil {
			t.Fatalf("STVertexConnectivity failed: %v", err)
		}
		if stVC != 1 {
			t.Errorf("expected STVertexConnectivity 1, got %d", stVC)
		}

		// Neighbor mode: adjacent vertices 0 and 1
		stVCNeg, err := g.STVertexConnectivity(0, 1, igraph.VConnNeigIgnore)
		if err != nil {
			t.Fatalf("STVertexConnectivity neighbors ignore failed: %v", err)
		}
		if stVCNeg < 0 {
			t.Errorf("expected non-negative vertex connectivity, got %d", stVCNeg)
		}

		stVCNum, err := g.STVertexConnectivity(0, 1, igraph.VConnNeigNumberOfNodes)
		if err != nil {
			t.Fatalf("STVertexConnectivity neighbors number of nodes failed: %v", err)
		}
		vcCount, _ := g.VertexCount()
		if stVCNum != vcCount && stVCNum < 0 {
			t.Errorf("unexpected vertex connectivity result: %d", stVCNum)
		}
	})

	t.Run("DisjointPaths", func(t *testing.T) {
		edp, err := g.EdgeDisjointPaths(0, 5)
		if err != nil {
			t.Fatalf("EdgeDisjointPaths failed: %v", err)
		}
		if edp != 1 {
			t.Errorf("expected EdgeDisjointPaths 1, got %d", edp)
		}

		vdp, err := g.VertexDisjointPaths(0, 5)
		if err != nil {
			t.Fatalf("VertexDisjointPaths failed: %v", err)
		}
		if vdp != 1 {
			t.Errorf("expected VertexDisjointPaths 1, got %d", vdp)
		}

		// Within same triangle 0 to 2
		edp2, err := g.EdgeDisjointPaths(0, 2)
		if err != nil {
			t.Fatalf("EdgeDisjointPaths within triangle failed: %v", err)
		}
		if edp2 != 2 {
			t.Errorf("expected 2 edge disjoint paths in triangle, got %d", edp2)
		}
	})
}

func TestConnectivityValidationAndClosed(t *testing.T) {
	g, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 1}}, true)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}
	defer g.Close()

	t.Run("out of bounds vertex", func(t *testing.T) {
		if _, err := g.STEdgeConnectivity(-1, 1); err == nil {
			t.Errorf("expected error for negative source")
		}
		if _, err := g.STVertexConnectivity(0, 10, igraph.VConnNeigError); err == nil {
			t.Errorf("expected error for out of bounds target")
		}
		if _, err := g.EdgeDisjointPaths(0, 10); err == nil {
			t.Errorf("expected error for out of bounds target")
		}
		if _, err := g.VertexDisjointPaths(-1, 1); err == nil {
			t.Errorf("expected error for negative source")
		}
	})

	t.Run("source equals target", func(t *testing.T) {
		if _, err := g.STEdgeConnectivity(1, 1); err == nil {
			t.Errorf("expected error when source == target")
		}
		if _, err := g.STVertexConnectivity(1, 1, igraph.VConnNeigError); err == nil {
			t.Errorf("expected error when source == target")
		}
	})

	t.Run("invalid neighbor mode", func(t *testing.T) {
		if _, err := g.STVertexConnectivity(0, 1, igraph.VertexConnectivityNeighbors(255)); err == nil {
			t.Errorf("expected error for invalid neighbor mode")
		}
	})

	t.Run("closed graph and nil graph", func(t *testing.T) {
		var nilGraph *igraph.Graph
		if _, err := nilGraph.EdgeConnectivity(true); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for nil graph")
		}
		if _, err := nilGraph.STEdgeConnectivity(0, 1); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for nil graph")
		}
		if _, err := nilGraph.VertexConnectivity(true); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for nil graph")
		}
		if _, err := nilGraph.STVertexConnectivity(0, 1, igraph.VConnNeigError); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for nil graph")
		}
		if _, err := nilGraph.EdgeDisjointPaths(0, 1); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for nil graph")
		}
		if _, err := nilGraph.VertexDisjointPaths(0, 1); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for nil graph")
		}
		if _, err := nilGraph.Adhesion(true); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for nil graph")
		}
		if _, err := nilGraph.Cohesion(true); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for nil graph")
		}

		g2, _ := igraph.NewGraphFromEdges(2, []igraph.Edge{{From: 0, To: 1}}, true)
		g2.Close()

		if _, err := g2.EdgeConnectivity(true); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for closed graph")
		}
		if _, err := g2.STEdgeConnectivity(0, 1); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for closed graph")
		}
		if _, err := g2.VertexConnectivity(true); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for closed graph")
		}
		if _, err := g2.STVertexConnectivity(0, 1, igraph.VConnNeigError); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for closed graph")
		}
		if _, err := g2.EdgeDisjointPaths(0, 1); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for closed graph")
		}
		if _, err := g2.VertexDisjointPaths(0, 1); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for closed graph")
		}
		if _, err := g2.Adhesion(true); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for closed graph")
		}
		if _, err := g2.Cohesion(true); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for closed graph")
		}
	})
}

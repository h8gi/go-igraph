package igraph_test

import (
	"errors"
	"math"
	"testing"

	"github.com/h8gi/go-igraph"
)

func TestResidualGraphAndReverseResidualGraph(t *testing.T) {
	edges := []igraph.Edge{
		{From: 0, To: 1},
		{From: 1, To: 2},
	}
	g, err := igraph.NewGraphFromEdges(3, edges, true)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}

	capacities := []float64{10.0, 5.0}
	flows := []float64{4.0, 2.0}

	resResult, err := g.ResidualGraph(capacities, flows)
	if err != nil {
		t.Fatalf("ResidualGraph failed: %v", err)
	}
	if resResult.Graph == nil {
		t.Fatalf("expected non-nil residual graph")
	}

	resResultNil, err := g.ResidualGraph(nil, nil)
	if err != nil {
		t.Fatalf("ResidualGraph with nil options failed: %v", err)
	}
	_ = resResultNil.Graph.Close()

	revResGraph, err := g.ReverseResidualGraph(capacities, flows)
	if err != nil {
		t.Fatalf("ReverseResidualGraph failed: %v", err)
	}
	if revResGraph == nil {
		t.Fatalf("expected non-nil reverse residual graph")
	}

	revResNil, err := g.ReverseResidualGraph(nil, nil)
	if err != nil {
		t.Fatalf("ReverseResidualGraph with nil options failed: %v", err)
	}
	_ = revResNil.Close()

	// Close parent graph; derived graphs must survive and be independently closeable
	g.Close()

	resEdgeCount, err := resResult.Graph.EdgeCount()
	if err != nil {
		t.Errorf("residual graph failed after parent close: %v", err)
	}
	if resEdgeCount <= 0 {
		t.Errorf("expected positive residual edge count, got %d", resEdgeCount)
	}
	if err := resResult.Graph.Close(); err != nil {
		t.Errorf("closing residual graph failed: %v", err)
	}

	revResEdgeCount, err := revResGraph.EdgeCount()
	if err != nil {
		t.Errorf("reverse residual graph failed after parent close: %v", err)
	}
	if revResEdgeCount <= 0 {
		t.Errorf("expected positive reverse residual edge count, got %d", revResEdgeCount)
	}
	if err := revResGraph.Close(); err != nil {
		t.Errorf("closing reverse residual graph failed: %v", err)
	}
}

func TestGomoryHuTree(t *testing.T) {
	// Undirected graph forming two triangles connected by bridge 2-3
	edges := []igraph.Edge{
		{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0},
		{From: 2, To: 3},
		{From: 3, To: 4}, {From: 4, To: 5}, {From: 5, To: 3},
	}
	g, err := igraph.NewGraphFromEdges(6, edges, false)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}

	ghtResult, err := g.GomoryHuTree(nil)
	if err != nil {
		t.Fatalf("GomoryHuTree failed: %v", err)
	}
	if ghtResult.Tree == nil {
		t.Fatalf("expected non-nil GomoryHu tree")
	}

	// Test GomoryHuTree with non-nil capacities
	capacities := []float64{1.0, 2.0, 1.0, 3.0, 1.0, 2.0, 1.0}
	ghtResult2, err := g.GomoryHuTree(capacities)
	if err != nil {
		t.Fatalf("GomoryHuTree with capacities failed: %v", err)
	}
	if ghtResult2.Tree == nil {
		t.Fatalf("expected non-nil GomoryHu tree with capacities")
	}
	_ = ghtResult2.Tree.Close()

	// Parent graph close
	g.Close()

	treeVC, err := ghtResult.Tree.VertexCount()
	if err != nil {
		t.Fatalf("tree VertexCount failed after parent close: %v", err)
	}
	if treeVC != 6 {
		t.Errorf("expected 6 vertices in Gomory-Hu tree, got %d", treeVC)
	}
	if err := ghtResult.Tree.Close(); err != nil {
		t.Errorf("closing Gomory-Hu tree failed: %v", err)
	}
}

func TestDominatorTree(t *testing.T) {
	// Directed graph: 0 -> 1 -> 2, 0 -> 2
	edges := []igraph.Edge{
		{From: 0, To: 1},
		{From: 1, To: 2},
		{From: 0, To: 2},
	}
	g, err := igraph.NewGraphFromEdges(3, edges, true)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}

	domResult, err := g.DominatorTree(0, igraph.DirectionOut)
	if err != nil {
		t.Fatalf("DominatorTree failed: %v", err)
	}
	if domResult.Tree == nil || domResult.Dominators == nil {
		t.Fatalf("expected non-nil Tree and Dominators in DominatorTreeResult")
	}

	domIn, err := g.DominatorTree(2, igraph.DirectionIn)
	if err != nil {
		t.Fatalf("DominatorTree DirectionIn failed: %v", err)
	}
	_ = domIn.Tree.Close()

	g.Close()

	treeVC, err := domResult.Tree.VertexCount()
	if err != nil {
		t.Fatalf("dominator tree VertexCount failed: %v", err)
	}
	if treeVC != 3 {
		t.Errorf("expected 3 vertices, got %d", treeVC)
	}
	if err := domResult.Tree.Close(); err != nil {
		t.Errorf("closing dominator tree failed: %v", err)
	}
}

func TestEvenTarjanReduction(t *testing.T) {
	edges := []igraph.Edge{
		{From: 0, To: 1},
		{From: 1, To: 2},
	}
	g, err := igraph.NewGraphFromEdges(3, edges, true)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}

	tarjanResult, err := g.EvenTarjanReduction()
	if err != nil {
		t.Fatalf("EvenTarjanReduction failed: %v", err)
	}
	if tarjanResult.Graph == nil {
		t.Fatalf("expected non-nil reduced graph")
	}

	g.Close()

	if err := tarjanResult.Graph.Close(); err != nil {
		t.Errorf("closing reduced graph failed: %v", err)
	}
}

func TestDerivedGraphsValidationAndClosed(t *testing.T) {
	g, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}}, true)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}
	defer g.Close()

	t.Run("invalid capacities and flows", func(t *testing.T) {
		if _, err := g.ResidualGraph([]float64{1.0}, nil); err == nil {
			t.Errorf("expected error for invalid capacities length")
		}
		if _, err := g.ResidualGraph(nil, []float64{1.0}); err == nil {
			t.Errorf("expected error for invalid flows length")
		}
		if _, err := g.ResidualGraph([]float64{1.0, -1.0}, nil); err == nil {
			t.Errorf("expected error for negative capacity")
		}
		if _, err := g.ResidualGraph(nil, []float64{1.0, math.NaN()}); err == nil {
			t.Errorf("expected error for NaN flow")
		}

		if _, err := g.ReverseResidualGraph([]float64{1.0}, nil); err == nil {
			t.Errorf("expected error for invalid capacities length")
		}
		if _, err := g.GomoryHuTree([]float64{1.0}); err == nil {
			t.Errorf("expected error for invalid capacities length")
		}
	})

	t.Run("invalid root and direction in dominator tree", func(t *testing.T) {
		if _, err := g.DominatorTree(-1, igraph.DirectionOut); err == nil {
			t.Errorf("expected error for negative root")
		}
		if _, err := g.DominatorTree(10, igraph.DirectionOut); err == nil {
			t.Errorf("expected error for out of bounds root")
		}
		if _, err := g.DominatorTree(0, igraph.DirectionMode(255)); err == nil {
			t.Errorf("expected error for invalid direction mode")
		}
	})

	t.Run("nil and closed graph", func(t *testing.T) {
		var nilGraph *igraph.Graph
		if _, err := nilGraph.ResidualGraph(nil, nil); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for nil graph ResidualGraph")
		}
		if _, err := nilGraph.ReverseResidualGraph(nil, nil); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for nil graph ReverseResidualGraph")
		}
		if _, err := nilGraph.GomoryHuTree(nil); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for nil graph GomoryHuTree")
		}
		if _, err := nilGraph.DominatorTree(0, igraph.DirectionOut); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for nil graph DominatorTree")
		}
		if _, err := nilGraph.EvenTarjanReduction(); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for nil graph EvenTarjanReduction")
		}

		g2, _ := igraph.NewGraphFromEdges(2, []igraph.Edge{{From: 0, To: 1}}, true)
		g2.Close()

		if _, err := g2.ResidualGraph(nil, nil); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for closed graph ResidualGraph")
		}
		if _, err := g2.ReverseResidualGraph(nil, nil); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for closed graph ReverseResidualGraph")
		}
		if _, err := g2.GomoryHuTree(nil); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for closed graph GomoryHuTree")
		}
		if _, err := g2.DominatorTree(0, igraph.DirectionOut); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for closed graph DominatorTree")
		}
		if _, err := g2.EvenTarjanReduction(); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for closed graph EvenTarjanReduction")
		}
	})
}

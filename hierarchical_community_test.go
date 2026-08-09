package igraph_test

import (
	"math"
	"testing"

	"github.com/h8gi/go-igraph"
)

// createTwoTriangles returns a 6-vertex graph consisting of two triangles (0-1-2 and 3-4-5)
// connected by a bridge edge (2-3). Edges = 7.
func createTwoTriangles(directed bool) (*igraph.Graph, error) {
	edges := []igraph.Edge{
		{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0}, // Triangle 1
		{From: 2, To: 3},                                     // Bridge
		{From: 3, To: 4}, {From: 4, To: 5}, {From: 5, To: 3}, // Triangle 2
	}
	return igraph.NewGraphFromEdges(6, edges, directed)
}

func TestCommunityFastGreedy(t *testing.T) {
	t.Run("normal unweighted", func(t *testing.T) {
		g, err := createTwoTriangles(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		hc, err := g.CommunityFastGreedy(nil)
		if err != nil {
			t.Fatalf("CommunityFastGreedy failed: %v", err)
		}

		if hc.NodeCount != 6 {
			t.Errorf("expected NodeCount 6, got %d", hc.NodeCount)
		}
		if len(hc.Merges) == 0 {
			t.Errorf("expected non-empty Merges")
		}
		if len(hc.Modularity) == 0 {
			t.Errorf("expected non-empty Modularity")
		}

		opt, err := hc.OptimalMembership()
		if err != nil {
			t.Fatalf("OptimalMembership failed: %v", err)
		}
		if opt.CommunityCount != 2 {
			t.Errorf("expected 2 communities at optimal split, got %d", opt.CommunityCount)
		}
		if len(opt.Membership) != 6 {
			t.Errorf("expected membership length 6, got %d", len(opt.Membership))
		}
		// Nodes 0,1,2 in one community, 3,4,5 in another
		if opt.Membership[0] != opt.Membership[1] || opt.Membership[1] != opt.Membership[2] {
			t.Errorf("nodes 0, 1, 2 should belong to the same community: %v", opt.Membership)
		}
		if opt.Membership[3] != opt.Membership[4] || opt.Membership[4] != opt.Membership[5] {
			t.Errorf("nodes 3, 4, 5 should belong to the same community: %v", opt.Membership)
		}
		if opt.Membership[0] == opt.Membership[3] {
			t.Errorf("triangle 1 and triangle 2 should be in different communities: %v", opt.Membership)
		}

		// Test MembershipAt all steps
		for step := 0; step <= len(hc.Merges); step++ {
			part, err := hc.MembershipAt(step)
			if err != nil {
				t.Fatalf("MembershipAt(%d) failed: %v", step, err)
			}
			if len(part.Membership) != 6 {
				t.Errorf("step %d: expected membership length 6, got %d", step, len(part.Membership))
			}
		}
	})

	t.Run("weighted", func(t *testing.T) {
		g, err := createTwoTriangles(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		weights := []float64{1.0, 1.0, 1.0, 0.1, 1.0, 1.0, 1.0}
		hc, err := g.CommunityFastGreedy(weights)
		if err != nil {
			t.Fatalf("CommunityFastGreedy weighted failed: %v", err)
		}
		if hc.NodeCount != 6 {
			t.Errorf("expected NodeCount 6, got %d", hc.NodeCount)
		}
		opt, err := hc.OptimalMembership()
		if err != nil {
			t.Fatalf("OptimalMembership failed: %v", err)
		}
		if opt.CommunityCount != 2 {
			t.Errorf("expected 2 communities, got %d", opt.CommunityCount)
		}
	})

	t.Run("directed graph error", func(t *testing.T) {
		g, err := createTwoTriangles(true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		_, err = g.CommunityFastGreedy(nil)
		if err == nil {
			t.Errorf("expected error running FastGreedy on directed graph, got nil")
		}
	})

	t.Run("empty graph", func(t *testing.T) {
		g, err := igraph.NewGraph()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		hc, err := g.CommunityFastGreedy(nil)
		if err != nil {
			t.Fatalf("unexpected error on empty graph: %v", err)
		}
		if hc.NodeCount != 0 {
			t.Errorf("expected NodeCount 0, got %d", hc.NodeCount)
		}
		if len(hc.Merges) != 0 {
			t.Errorf("expected empty Merges, got %v", hc.Merges)
		}
		part, err := hc.OptimalMembership()
		if err != nil {
			t.Fatalf("unexpected error on OptimalMembership: %v", err)
		}
		if part.CommunityCount != 0 {
			t.Errorf("expected 0 communities, got %d", part.CommunityCount)
		}
	})

	t.Run("single vertex graph", func(t *testing.T) {
		g, err := igraph.NewGraphFromEdges(1, nil, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		hc, err := g.CommunityFastGreedy(nil)
		if err != nil {
			t.Fatalf("unexpected error on single vertex graph: %v", err)
		}
		if hc.NodeCount != 1 {
			t.Errorf("expected NodeCount 1, got %d", hc.NodeCount)
		}
		part, err := hc.OptimalMembership()
		if err != nil {
			t.Fatalf("unexpected error on OptimalMembership: %v", err)
		}
		if len(part.Membership) != 1 {
			t.Errorf("expected membership len 1, got %d", len(part.Membership))
		}
	})

	t.Run("disconnected graph", func(t *testing.T) {
		edges := []igraph.Edge{
			{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0},
			{From: 3, To: 4}, {From: 4, To: 5}, {From: 5, To: 3},
		}
		g, err := igraph.NewGraphFromEdges(6, edges, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		hc, err := g.CommunityFastGreedy(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if hc.NodeCount != 6 {
			t.Errorf("expected NodeCount 6, got %d", hc.NodeCount)
		}
		opt, err := hc.OptimalMembership()
		if err != nil {
			t.Fatalf("unexpected error on OptimalMembership: %v", err)
		}
		if opt.CommunityCount != 2 {
			t.Errorf("expected 2 communities for disconnected components, got %d", opt.CommunityCount)
		}
	})

	t.Run("invalid weights length", func(t *testing.T) {
		g, err := createTwoTriangles(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		_, err = g.CommunityFastGreedy([]float64{1.0, 2.0})
		if err == nil {
			t.Errorf("expected error for invalid weights length, got nil")
		}
	})

	t.Run("invalid weight values", func(t *testing.T) {
		g, err := createTwoTriangles(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		weights := []float64{1.0, 1.0, 1.0, math.NaN(), 1.0, 1.0, 1.0}
		_, err = g.CommunityFastGreedy(weights)
		if err == nil {
			t.Errorf("expected error for NaN weight, got nil")
		}
	})

	t.Run("closed graph", func(t *testing.T) {
		g, err := createTwoTriangles(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		g.Close()

		_, err = g.CommunityFastGreedy(nil)
		if err != igraph.ErrClosed {
			t.Errorf("expected ErrClosed, got %v", err)
		}
	})

	t.Run("nil graph", func(t *testing.T) {
		var g *igraph.Graph
		_, err := g.CommunityFastGreedy(nil)
		if err != igraph.ErrClosed {
			t.Errorf("expected ErrClosed, got %v", err)
		}
	})
}

func TestCommunityWalktrap(t *testing.T) {
	t.Run("normal unweighted", func(t *testing.T) {
		g, err := createTwoTriangles(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		for _, steps := range []int{1, 2, 4, 10} {
			hc, err := g.CommunityWalktrap(nil, steps)
			if err != nil {
				t.Fatalf("CommunityWalktrap(steps=%d) failed: %v", steps, err)
			}
			if hc.NodeCount != 6 {
				t.Errorf("expected NodeCount 6, got %d", hc.NodeCount)
			}
			if len(hc.Merges) == 0 {
				t.Errorf("expected non-empty Merges")
			}

			opt, err := hc.OptimalMembership()
			if err != nil {
				t.Fatalf("OptimalMembership failed: %v", err)
			}
			if len(opt.Membership) != 6 {
				t.Errorf("expected membership len 6, got %d", len(opt.Membership))
			}
		}
	})

	t.Run("weighted", func(t *testing.T) {
		g, err := createTwoTriangles(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		weights := []float64{1.0, 1.0, 1.0, 0.1, 1.0, 1.0, 1.0}
		hc, err := g.CommunityWalktrap(weights, 4)
		if err != nil {
			t.Fatalf("CommunityWalktrap weighted failed: %v", err)
		}
		if hc.NodeCount != 6 {
			t.Errorf("expected NodeCount 6, got %d", hc.NodeCount)
		}
		opt, err := hc.OptimalMembership()
		if err != nil {
			t.Fatalf("OptimalMembership failed: %v", err)
		}
		if len(opt.Membership) != 6 {
			t.Errorf("expected membership len 6, got %d", len(opt.Membership))
		}
	})

	t.Run("directed graph", func(t *testing.T) {
		g, err := createTwoTriangles(true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		hc, err := g.CommunityWalktrap(nil, 4)
		if err != nil {
			t.Fatalf("CommunityWalktrap on directed graph failed: %v", err)
		}
		if hc.NodeCount != 6 {
			t.Errorf("expected NodeCount 6, got %d", hc.NodeCount)
		}
	})

	t.Run("empty graph", func(t *testing.T) {
		g, err := igraph.NewGraph()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		hc, err := g.CommunityWalktrap(nil, 4)
		if err != nil {
			t.Fatalf("unexpected error on empty graph: %v", err)
		}
		if hc.NodeCount != 0 {
			t.Errorf("expected NodeCount 0, got %d", hc.NodeCount)
		}
	})

	t.Run("single vertex graph", func(t *testing.T) {
		g, err := igraph.NewGraphFromEdges(1, nil, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		hc, err := g.CommunityWalktrap(nil, 4)
		if err != nil {
			t.Fatalf("unexpected error on single vertex graph: %v", err)
		}
		if hc.NodeCount != 1 {
			t.Errorf("expected NodeCount 1, got %d", hc.NodeCount)
		}
	})

	t.Run("disconnected graph", func(t *testing.T) {
		edges := []igraph.Edge{
			{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0},
			{From: 3, To: 4}, {From: 4, To: 5}, {From: 5, To: 3},
		}
		g, err := igraph.NewGraphFromEdges(6, edges, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		hc, err := g.CommunityWalktrap(nil, 4)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if hc.NodeCount != 6 {
			t.Errorf("expected NodeCount 6, got %d", hc.NodeCount)
		}
	})

	t.Run("invalid steps", func(t *testing.T) {
		g, err := createTwoTriangles(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		_, err = g.CommunityWalktrap(nil, 0)
		if err == nil {
			t.Errorf("expected error for steps=0, got nil")
		}
	})

	t.Run("steps overflow", func(t *testing.T) {
		g, err := createTwoTriangles(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		_, err = g.CommunityWalktrap(nil, math.MaxInt64)
		if err == nil {
			t.Errorf("expected error for steps overflow, got nil")
		}
	})

	t.Run("invalid weights length", func(t *testing.T) {
		g, err := createTwoTriangles(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		_, err = g.CommunityWalktrap([]float64{1.0}, 4)
		if err == nil {
			t.Errorf("expected error for invalid weights length, got nil")
		}
	})

	t.Run("closed graph", func(t *testing.T) {
		g, err := createTwoTriangles(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		g.Close()

		_, err = g.CommunityWalktrap(nil, 4)
		if err != igraph.ErrClosed {
			t.Errorf("expected ErrClosed, got %v", err)
		}
	})

	t.Run("nil graph", func(t *testing.T) {
		var g *igraph.Graph
		_, err := g.CommunityWalktrap(nil, 4)
		if err != igraph.ErrClosed {
			t.Errorf("expected ErrClosed, got %v", err)
		}
	})
}

func TestCommunityEdgeBetweenness(t *testing.T) {
	t.Run("normal unweighted undirected", func(t *testing.T) {
		g, err := createTwoTriangles(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		hc, err := g.CommunityEdgeBetweenness(nil, false)
		if err != nil {
			t.Fatalf("CommunityEdgeBetweenness failed: %v", err)
		}

		if hc.NodeCount != 6 {
			t.Errorf("expected NodeCount 6, got %d", hc.NodeCount)
		}
		if len(hc.Merges) == 0 {
			t.Errorf("expected non-empty Merges")
		}

		opt, err := hc.OptimalMembership()
		if err != nil {
			t.Fatalf("OptimalMembership failed: %v", err)
		}
		if len(opt.Membership) != 6 {
			t.Errorf("expected membership len 6, got %d", len(opt.Membership))
		}
		if opt.CommunityCount != 2 {
			t.Errorf("expected 2 communities, got %d", opt.CommunityCount)
		}
	})

	t.Run("directed graph", func(t *testing.T) {
		g, err := createTwoTriangles(true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		hc, err := g.CommunityEdgeBetweenness(nil, true)
		if err != nil {
			t.Fatalf("CommunityEdgeBetweenness directed failed: %v", err)
		}
		if hc.NodeCount != 6 {
			t.Errorf("expected NodeCount 6, got %d", hc.NodeCount)
		}
	})

	t.Run("weighted", func(t *testing.T) {
		g, err := createTwoTriangles(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		weights := []float64{1.0, 1.0, 1.0, 0.1, 1.0, 1.0, 1.0}
		hc, err := g.CommunityEdgeBetweenness(weights, false)
		if err != nil {
			t.Fatalf("CommunityEdgeBetweenness weighted failed: %v", err)
		}
		if hc.NodeCount != 6 {
			t.Errorf("expected NodeCount 6, got %d", hc.NodeCount)
		}
	})

	t.Run("empty graph", func(t *testing.T) {
		g, err := igraph.NewGraph()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		hc, err := g.CommunityEdgeBetweenness(nil, false)
		if err != nil {
			t.Fatalf("unexpected error on empty graph: %v", err)
		}
		if hc.NodeCount != 0 {
			t.Errorf("expected NodeCount 0, got %d", hc.NodeCount)
		}
	})

	t.Run("single vertex graph", func(t *testing.T) {
		g, err := igraph.NewGraphFromEdges(1, nil, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		hc, err := g.CommunityEdgeBetweenness(nil, false)
		if err != nil {
			t.Fatalf("unexpected error on single vertex graph: %v", err)
		}
		if hc.NodeCount != 1 {
			t.Errorf("expected NodeCount 1, got %d", hc.NodeCount)
		}
	})

	t.Run("disconnected graph", func(t *testing.T) {
		edges := []igraph.Edge{
			{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0},
			{From: 3, To: 4}, {From: 4, To: 5}, {From: 5, To: 3},
		}
		g, err := igraph.NewGraphFromEdges(6, edges, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		hc, err := g.CommunityEdgeBetweenness(nil, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if hc.NodeCount != 6 {
			t.Errorf("expected NodeCount 6, got %d", hc.NodeCount)
		}
	})

	t.Run("invalid weights length", func(t *testing.T) {
		g, err := createTwoTriangles(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		_, err = g.CommunityEdgeBetweenness([]float64{1.0}, false)
		if err == nil {
			t.Errorf("expected error for invalid weights length, got nil")
		}
	})

	t.Run("invalid weight values", func(t *testing.T) {
		g, err := createTwoTriangles(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		weights := []float64{1.0, 1.0, 1.0, math.NaN(), 1.0, 1.0, 1.0}
		_, err = g.CommunityEdgeBetweenness(weights, false)
		if err == nil {
			t.Errorf("expected error for NaN weight, got nil")
		}
	})

	t.Run("closed graph", func(t *testing.T) {
		g, err := createTwoTriangles(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		g.Close()

		_, err = g.CommunityEdgeBetweenness(nil, false)
		if err != igraph.ErrClosed {
			t.Errorf("expected ErrClosed, got %v", err)
		}
	})

	t.Run("nil graph", func(t *testing.T) {
		var g *igraph.Graph
		_, err := g.CommunityEdgeBetweenness(nil, false)
		if err != igraph.ErrClosed {
			t.Errorf("expected ErrClosed, got %v", err)
		}
	})
}

func TestCommunityEBGetMerges(t *testing.T) {
	t.Run("normal unweighted", func(t *testing.T) {
		g, err := createTwoTriangles(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		// Edge 3 is the bridge between node 2 and 3
		removedEdges := []int{3, 0, 1, 2, 4, 5, 6}
		hc, err := g.CommunityEBGetMerges(removedEdges, nil, false)
		if err != nil {
			t.Fatalf("CommunityEBGetMerges failed: %v", err)
		}

		if hc.NodeCount != 6 {
			t.Errorf("expected NodeCount 6, got %d", hc.NodeCount)
		}
		if len(hc.Merges) == 0 {
			t.Errorf("expected non-empty Merges")
		}

		opt, err := hc.OptimalMembership()
		if err != nil {
			t.Fatalf("OptimalMembership failed: %v", err)
		}
		if len(opt.Membership) != 6 {
			t.Errorf("expected membership len 6, got %d", len(opt.Membership))
		}
	})

	t.Run("weighted directed", func(t *testing.T) {
		g, err := createTwoTriangles(true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		removedEdges := []int{3, 0, 1, 2, 4, 5, 6}
		weights := []float64{1.0, 1.0, 1.0, 0.1, 1.0, 1.0, 1.0}
		hc, err := g.CommunityEBGetMerges(removedEdges, weights, true)
		if err != nil {
			t.Fatalf("CommunityEBGetMerges weighted failed: %v", err)
		}
		if hc.NodeCount != 6 {
			t.Errorf("expected NodeCount 6, got %d", hc.NodeCount)
		}
	})

	t.Run("empty graph", func(t *testing.T) {
		g, err := igraph.NewGraph()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		hc, err := g.CommunityEBGetMerges([]int{}, nil, false)
		if err != nil {
			t.Fatalf("unexpected error on empty graph: %v", err)
		}
		if hc.NodeCount != 0 {
			t.Errorf("expected NodeCount 0, got %d", hc.NodeCount)
		}
	})

	t.Run("single vertex graph", func(t *testing.T) {
		g, err := igraph.NewGraphFromEdges(1, nil, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		hc, err := g.CommunityEBGetMerges([]int{}, nil, false)
		if err != nil {
			t.Fatalf("unexpected error on single vertex graph: %v", err)
		}
		if hc.NodeCount != 1 {
			t.Errorf("expected NodeCount 1, got %d", hc.NodeCount)
		}
	})

	t.Run("invalid edge id", func(t *testing.T) {
		g, err := createTwoTriangles(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		_, err = g.CommunityEBGetMerges([]int{999}, nil, false)
		if err == nil {
			t.Errorf("expected error for invalid edge id, got nil")
		}
	})

	t.Run("edges int overflow", func(t *testing.T) {
		g, err := createTwoTriangles(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		_, err = g.CommunityEBGetMerges([]int{math.MaxInt64}, nil, false)
		if err == nil {
			t.Errorf("expected error for edges overflow, got nil")
		}
	})

	t.Run("invalid weights length", func(t *testing.T) {
		g, err := createTwoTriangles(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		removedEdges := []int{3, 0, 1, 2, 4, 5, 6}
		_, err = g.CommunityEBGetMerges(removedEdges, []float64{1.0}, false)
		if err == nil {
			t.Errorf("expected error for invalid weights length, got nil")
		}
	})

	t.Run("invalid weight values", func(t *testing.T) {
		g, err := createTwoTriangles(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer g.Close()

		removedEdges := []int{3, 0, 1, 2, 4, 5, 6}
		weights := []float64{1.0, 1.0, 1.0, math.NaN(), 1.0, 1.0, 1.0}
		_, err = g.CommunityEBGetMerges(removedEdges, weights, false)
		if err == nil {
			t.Errorf("expected error for NaN weight, got nil")
		}
	})

	t.Run("closed graph", func(t *testing.T) {
		g, err := createTwoTriangles(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		g.Close()

		_, err = g.CommunityEBGetMerges([]int{0}, nil, false)
		if err != igraph.ErrClosed {
			t.Errorf("expected ErrClosed, got %v", err)
		}
	})

	t.Run("nil graph", func(t *testing.T) {
		var g *igraph.Graph
		_, err := g.CommunityEBGetMerges([]int{0}, nil, false)
		if err != igraph.ErrClosed {
			t.Errorf("expected ErrClosed, got %v", err)
		}
	})
}

func TestHierarchicalCommunityBounds(t *testing.T) {
	g, err := createTwoTriangles(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer g.Close()

	hc, err := g.CommunityFastGreedy(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("out of bounds step negative", func(t *testing.T) {
		_, err := hc.MembershipAt(-1)
		if err == nil {
			t.Errorf("expected error for negative steps, got nil")
		}
	})

	t.Run("out of bounds step max+1", func(t *testing.T) {
		_, err := hc.MembershipAt(len(hc.Merges) + 1)
		if err == nil {
			t.Errorf("expected error for steps > max, got nil")
		}
	})

	t.Run("OptimalMembership empty community", func(t *testing.T) {
		emptyHC := igraph.HierarchicalCommunity{
			Merges:     [][2]int{},
			Modularity: []float64{},
			NodeCount:  0,
		}
		part, err := emptyHC.OptimalMembership()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if part.CommunityCount != 0 {
			t.Errorf("expected 0 communities, got %d", part.CommunityCount)
		}
	})

	t.Run("OptimalMembership modularity with NaNs", func(t *testing.T) {
		nanHC := igraph.HierarchicalCommunity{
			Merges:     [][2]int{{0, 1}},
			Modularity: []float64{math.NaN(), math.NaN()},
			NodeCount:  2,
		}
		part, err := nanHC.OptimalMembership()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(part.Membership) != 2 {
			t.Errorf("expected membership len 2, got %d", len(part.Membership))
		}
	})
}

package igraph_test

import (
	"math"
	"reflect"
	"testing"

	"github.com/h8gi/go-igraph"
)

func TestCoreness(t *testing.T) {
	t.Run("triangle graph", func(t *testing.T) {
		// Triangle graph K3: 3 vertices, 3 edges forming a triangle.
		edges := []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0}}
		g, err := igraph.NewGraphFromEdges(3, edges, false)
		if err != nil {
			t.Fatalf("failed to create graph: %v", err)
		}
		defer g.Close()

		coreness, err := g.Coreness(igraph.NeiAll)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []int{2, 2, 2}
		if !reflect.DeepEqual(coreness, expected) {
			t.Errorf("got coreness %v, want %v", coreness, expected)
		}
	})

	t.Run("star graph", func(t *testing.T) {
		// Star graph: 4 vertices, center 0 connected to 1, 2, 3.
		edges := []igraph.Edge{{From: 0, To: 1}, {From: 0, To: 2}, {From: 0, To: 3}}
		g, err := igraph.NewGraphFromEdges(4, edges, false)
		if err != nil {
			t.Fatalf("failed to create graph: %v", err)
		}
		defer g.Close()

		coreness, err := g.Coreness(igraph.NeiAll)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []int{1, 1, 1, 1}
		if !reflect.DeepEqual(coreness, expected) {
			t.Errorf("got coreness %v, want %v", coreness, expected)
		}
	})

	t.Run("directed modes", func(t *testing.T) {
		// Directed graph: 0 -> 1 -> 2, 0 -> 2
		edges := []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 0, To: 2}}
		g, err := igraph.NewGraphFromEdges(3, edges, true)
		if err != nil {
			t.Fatalf("failed to create graph: %v", err)
		}
		defer g.Close()

		cOut, err := g.Coreness(igraph.NeiOut)
		if err != nil {
			t.Fatalf("Coreness(NeiOut) error: %v", err)
		}
		if len(cOut) != 3 {
			t.Errorf("got len %d, want 3", len(cOut))
		}

		cIn, err := g.Coreness(igraph.NeiIn)
		if err != nil {
			t.Fatalf("Coreness(NeiIn) error: %v", err)
		}
		if len(cIn) != 3 {
			t.Errorf("got len %d, want 3", len(cIn))
		}
	})

	t.Run("empty graph", func(t *testing.T) {
		g, err := igraph.NewGraph()
		if err != nil {
			t.Fatalf("failed to create empty graph: %v", err)
		}
		defer g.Close()

		coreness, err := g.Coreness(igraph.NeiAll)
		if err != nil {
			t.Fatalf("unexpected error on empty graph: %v", err)
		}
		if len(coreness) != 0 {
			t.Errorf("got len %d, want 0", len(coreness))
		}
	})

	t.Run("disconnected graph", func(t *testing.T) {
		// K3 triangle (0,1,2) plus isolated vertex 3.
		edges := []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0}}
		g, err := igraph.NewGraphFromEdges(4, edges, false)
		if err != nil {
			t.Fatalf("failed to create graph: %v", err)
		}
		defer g.Close()

		coreness, err := g.Coreness(igraph.NeiAll)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []int{2, 2, 2, 0}
		if !reflect.DeepEqual(coreness, expected) {
			t.Errorf("got coreness %v, want %v", coreness, expected)
		}
	})

	t.Run("invalid mode", func(t *testing.T) {
		g, err := igraph.NewGraphFromEdges(2, []igraph.Edge{{From: 0, To: 1}}, false)
		if err != nil {
			t.Fatalf("failed to create graph: %v", err)
		}
		defer g.Close()

		_, err = g.Coreness(igraph.NeiMode(99))
		if err == nil {
			t.Fatal("expected error for invalid NeiMode, got nil")
		}
	})

	t.Run("closed and nil graph", func(t *testing.T) {
		g, err := igraph.NewGraphFromEdges(2, []igraph.Edge{{From: 0, To: 1}}, false)
		if err != nil {
			t.Fatalf("failed to create graph: %v", err)
		}
		g.Close()

		if _, err := g.Coreness(igraph.NeiAll); err != igraph.ErrClosed {
			t.Errorf("got err %v, want ErrClosed on closed graph", err)
		}

		var nilG *igraph.Graph
		if _, err := nilG.Coreness(igraph.NeiAll); err != igraph.ErrClosed {
			t.Errorf("got err %v, want ErrClosed on nil graph", err)
		}
	})
}

func TestTrussness(t *testing.T) {
	t.Run("triangle graph", func(t *testing.T) {
		// Triangle graph: 3 vertices, 3 edges forming K3.
		edges := []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0}}
		g, err := igraph.NewGraphFromEdges(3, edges, false)
		if err != nil {
			t.Fatalf("failed to create graph: %v", err)
		}
		defer g.Close()

		trussness, err := g.Trussness()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []int{3, 3, 3}
		if !reflect.DeepEqual(trussness, expected) {
			t.Errorf("got trussness %v, want %v", trussness, expected)
		}
	})

	t.Run("path graph", func(t *testing.T) {
		// Path graph: 3 vertices, 2 edges 0-1, 1-2. No triangles.
		edges := []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}}
		g, err := igraph.NewGraphFromEdges(3, edges, false)
		if err != nil {
			t.Fatalf("failed to create graph: %v", err)
		}
		defer g.Close()

		trussness, err := g.Trussness()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []int{2, 2}
		if !reflect.DeepEqual(trussness, expected) {
			t.Errorf("got trussness %v, want %v", trussness, expected)
		}
	})

	t.Run("empty graph", func(t *testing.T) {
		g, err := igraph.NewGraph()
		if err != nil {
			t.Fatalf("failed to create empty graph: %v", err)
		}
		defer g.Close()

		trussness, err := g.Trussness()
		if err != nil {
			t.Fatalf("unexpected error on empty graph: %v", err)
		}
		if len(trussness) != 0 {
			t.Errorf("got len %d, want 0", len(trussness))
		}
	})

	t.Run("disconnected graph", func(t *testing.T) {
		// K3 triangle (0-1, 1-2, 2-0) and isolated edge (3-4).
		edges := []igraph.Edge{
			{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0},
			{From: 3, To: 4},
		}
		g, err := igraph.NewGraphFromEdges(5, edges, false)
		if err != nil {
			t.Fatalf("failed to create graph: %v", err)
		}
		defer g.Close()

		trussness, err := g.Trussness()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []int{3, 3, 3, 2}
		if !reflect.DeepEqual(trussness, expected) {
			t.Errorf("got trussness %v, want %v", trussness, expected)
		}
	})

	t.Run("closed and nil graph", func(t *testing.T) {
		g, err := igraph.NewGraphFromEdges(2, []igraph.Edge{{From: 0, To: 1}}, false)
		if err != nil {
			t.Fatalf("failed to create graph: %v", err)
		}
		g.Close()

		if _, err := g.Trussness(); err != igraph.ErrClosed {
			t.Errorf("got err %v, want ErrClosed on closed graph", err)
		}

		var nilG *igraph.Graph
		if _, err := nilG.Trussness(); err != igraph.ErrClosed {
			t.Errorf("got err %v, want ErrClosed on nil graph", err)
		}
	})
}

func TestModularity(t *testing.T) {
	t.Run("two disjoint triangles", func(t *testing.T) {
		// 6 vertices, 2 components: triangle {0,1,2} and triangle {3,4,5}.
		edges := []igraph.Edge{
			{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0},
			{From: 3, To: 4}, {From: 4, To: 5}, {From: 5, To: 3},
		}
		g, err := igraph.NewGraphFromEdges(6, edges, false)
		if err != nil {
			t.Fatalf("failed to create graph: %v", err)
		}
		defer g.Close()

		// Perfect partition matching the two components.
		membership := []int{0, 0, 0, 1, 1, 1}
		mod, err := g.Modularity(membership, nil, 1.0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mod <= 0.4 {
			t.Errorf("got modularity %f, expected > 0.4 for perfect 2-community split", mod)
		}

		// Single community partition.
		singleMem := []int{0, 0, 0, 0, 0, 0}
		modSingle, err := g.Modularity(singleMem, nil, 1.0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if math.Abs(modSingle) > 1e-9 {
			t.Errorf("got modularity %f for single community, expected ~0.0", modSingle)
		}
	})

	t.Run("weighted modularity", func(t *testing.T) {
		edges := []igraph.Edge{
			{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0},
			{From: 3, To: 4}, {From: 4, To: 5}, {From: 5, To: 3},
		}
		g, err := igraph.NewGraphFromEdges(6, edges, false)
		if err != nil {
			t.Fatalf("failed to create graph: %v", err)
		}
		defer g.Close()

		weights := []float64{1.0, 1.0, 1.0, 2.0, 2.0, 2.0}
		membership := []int{0, 0, 0, 1, 1, 1}
		mod, err := g.Modularity(membership, weights, 1.0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mod <= 0.4 {
			t.Errorf("got weighted modularity %f, expected > 0.4", mod)
		}
	})

	t.Run("directed graph", func(t *testing.T) {
		edges := []igraph.Edge{
			{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0},
			{From: 3, To: 4}, {From: 4, To: 5}, {From: 5, To: 3},
		}
		g, err := igraph.NewGraphFromEdges(6, edges, true)
		if err != nil {
			t.Fatalf("failed to create graph: %v", err)
		}
		defer g.Close()

		membership := []int{0, 0, 0, 1, 1, 1}
		mod, err := g.Modularity(membership, nil, 1.0)
		if err != nil {
			t.Fatalf("unexpected error on directed graph: %v", err)
		}
		if mod <= 0.4 {
			t.Errorf("got directed modularity %f, expected > 0.4", mod)
		}
	})

	t.Run("empty graph", func(t *testing.T) {
		g, err := igraph.NewGraph()
		if err != nil {
			t.Fatalf("failed to create empty graph: %v", err)
		}
		defer g.Close()

		mod, err := g.Modularity([]int{}, nil, 1.0)
		if err != nil {
			t.Fatalf("unexpected error on empty graph: %v", err)
		}
		_ = mod
	})

	t.Run("invalid membership length", func(t *testing.T) {
		g, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 1}}, false)
		if err != nil {
			t.Fatalf("failed to create graph: %v", err)
		}
		defer g.Close()

		_, err = g.Modularity([]int{0, 1}, nil, 1.0)
		if err == nil {
			t.Fatal("expected error for membership length mismatch, got nil")
		}
	})

	t.Run("invalid weights length", func(t *testing.T) {
		g, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 1}}, false)
		if err != nil {
			t.Fatalf("failed to create graph: %v", err)
		}
		defer g.Close()

		_, err = g.Modularity([]int{0, 0, 0}, []float64{1.0, 2.0}, 1.0)
		if err == nil {
			t.Fatal("expected error for weights length mismatch, got nil")
		}
	})

	t.Run("invalid non-finite weight", func(t *testing.T) {
		g, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 1}}, false)
		if err != nil {
			t.Fatalf("failed to create graph: %v", err)
		}
		defer g.Close()

		_, err = g.Modularity([]int{0, 0, 0}, []float64{math.NaN()}, 1.0)
		if err == nil {
			t.Fatal("expected error for NaN weight, got nil")
		}

		_, err = g.Modularity([]int{0, 0, 0}, []float64{math.Inf(1)}, 1.0)
		if err == nil {
			t.Fatal("expected error for Inf weight, got nil")
		}
	})

	t.Run("closed and nil graph", func(t *testing.T) {
		g, err := igraph.NewGraphFromEdges(2, []igraph.Edge{{From: 0, To: 1}}, false)
		if err != nil {
			t.Fatalf("failed to create graph: %v", err)
		}
		g.Close()

		if _, err := g.Modularity([]int{0, 1}, nil, 1.0); err != igraph.ErrClosed {
			t.Errorf("got err %v, want ErrClosed on closed graph", err)
		}

		var nilG *igraph.Graph
		if _, err := nilG.Modularity([]int{0, 1}, nil, 1.0); err != igraph.ErrClosed {
			t.Errorf("got err %v, want ErrClosed on nil graph", err)
		}
	})
}

func TestModularityMatrix(t *testing.T) {
	t.Run("cycle graph", func(t *testing.T) {
		// Cycle C4: 4 vertices, edges (0,1),(1,2),(2,3),(3,0)
		edges := []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 3}, {From: 3, To: 0}}
		g, err := igraph.NewGraphFromEdges(4, edges, false)
		if err != nil {
			t.Fatalf("failed to create graph: %v", err)
		}
		defer g.Close()

		mat, err := g.ModularityMatrix(nil, 1.0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mat == nil {
			t.Fatal("got nil matrix")
		}
		r, c := mat.Dims()
		if r != 4 || c != 4 {
			t.Errorf("got dimensions %dx%d, want 4x4", r, c)
		}
	})

	t.Run("weighted modularity matrix", func(t *testing.T) {
		edges := []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0}}
		g, err := igraph.NewGraphFromEdges(3, edges, false)
		if err != nil {
			t.Fatalf("failed to create graph: %v", err)
		}
		defer g.Close()

		weights := []float64{1.0, 2.0, 3.0}
		mat, err := g.ModularityMatrix(weights, 1.0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mat == nil {
			t.Fatal("got nil matrix")
		}
		r, c := mat.Dims()
		if r != 3 || c != 3 {
			t.Errorf("got dimensions %dx%d, want 3x3", r, c)
		}
	})

	t.Run("directed modularity matrix", func(t *testing.T) {
		edges := []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0}}
		g, err := igraph.NewGraphFromEdges(3, edges, true)
		if err != nil {
			t.Fatalf("failed to create graph: %v", err)
		}
		defer g.Close()

		mat, err := g.ModularityMatrix(nil, 1.0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mat == nil {
			t.Fatal("got nil matrix")
		}
		r, c := mat.Dims()
		if r != 3 || c != 3 {
			t.Errorf("got dimensions %dx%d, want 3x3", r, c)
		}
	})

	t.Run("empty graph", func(t *testing.T) {
		g, err := igraph.NewGraph()
		if err != nil {
			t.Fatalf("failed to create empty graph: %v", err)
		}
		defer g.Close()

		mat, err := g.ModularityMatrix(nil, 1.0)
		if err != nil {
			t.Fatalf("unexpected error on empty graph: %v", err)
		}
		if mat == nil {
			t.Fatal("got nil matrix for empty graph")
		}
		r, c := mat.Dims()
		if r != 0 || c != 0 {
			t.Errorf("got dimensions %dx%d, want 0x0", r, c)
		}
	})

	t.Run("invalid weights length", func(t *testing.T) {
		g, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 1}}, false)
		if err != nil {
			t.Fatalf("failed to create graph: %v", err)
		}
		defer g.Close()

		_, err = g.ModularityMatrix([]float64{1.0, 2.0}, 1.0)
		if err == nil {
			t.Fatal("expected error for weights length mismatch, got nil")
		}
	})

	t.Run("invalid non-finite weight", func(t *testing.T) {
		g, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 1}}, false)
		if err != nil {
			t.Fatalf("failed to create graph: %v", err)
		}
		defer g.Close()

		_, err = g.ModularityMatrix([]float64{math.NaN()}, 1.0)
		if err == nil {
			t.Fatal("expected error for NaN weight, got nil")
		}
	})

	t.Run("closed and nil graph", func(t *testing.T) {
		g, err := igraph.NewGraphFromEdges(2, []igraph.Edge{{From: 0, To: 1}}, false)
		if err != nil {
			t.Fatalf("failed to create graph: %v", err)
		}
		g.Close()

		if _, err := g.ModularityMatrix(nil, 1.0); err != igraph.ErrClosed {
			t.Errorf("got err %v, want ErrClosed on closed graph", err)
		}

		var nilG *igraph.Graph
		if _, err := nilG.ModularityMatrix(nil, 1.0); err != igraph.ErrClosed {
			t.Errorf("got err %v, want ErrClosed on nil graph", err)
		}
	})
}

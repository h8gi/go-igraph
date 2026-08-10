package igraph_test

import (
	"errors"
	"math"
	"testing"

	"github.com/h8gi/go-igraph"
)

func TestAllSTCutsAndMincuts(t *testing.T) {
	// Diamond graph: 0 -> 1 -> 3, 0 -> 2 -> 3
	edges := []igraph.Edge{
		{From: 0, To: 1},
		{From: 0, To: 2},
		{From: 1, To: 3},
		{From: 2, To: 3},
	}
	g, err := igraph.NewGraphFromEdges(4, edges, true)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}
	defer g.Close()

	t.Run("AllSTCuts", func(t *testing.T) {
		cuts, err := g.AllSTCuts(0, 3)
		if err != nil {
			t.Fatalf("AllSTCuts failed: %v", err)
		}
		if len(cuts) == 0 {
			t.Errorf("expected at least one s-t cut, got 0")
		}
		for _, c := range cuts {
			if c.Cut == nil || c.Partition == nil {
				t.Errorf("expected non-nil Cut and Partition in STCut")
			}
		}
	})

	t.Run("AllSTMincuts unit capacities", func(t *testing.T) {
		minVal, cuts, err := g.AllSTMincuts(0, 3, nil)
		if err != nil {
			t.Fatalf("AllSTMincuts failed: %v", err)
		}
		if minVal != 2.0 {
			t.Errorf("expected min cut value 2.0, got %v", minVal)
		}
		if len(cuts) == 0 {
			t.Errorf("expected at least one mincut, got 0")
		}
	})

	t.Run("AllSTMincuts custom capacities", func(t *testing.T) {
		capacities := []float64{10.0, 5.0, 10.0, 5.0}
		minVal, cuts, err := g.AllSTMincuts(0, 3, capacities)
		if err != nil {
			t.Fatalf("AllSTMincuts with custom capacities failed: %v", err)
		}
		if minVal != 15.0 {
			t.Errorf("expected min cut value 15.0, got %v", minVal)
		}
		if len(cuts) == 0 {
			t.Errorf("expected at least one mincut, got 0")
		}
	})
}

func TestAllSTCutsValidationAndClosed(t *testing.T) {
	edges := []igraph.Edge{
		{From: 0, To: 1},
		{From: 1, To: 2},
	}
	g, err := igraph.NewGraphFromEdges(3, edges, true)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}

	t.Run("out of bounds vertex", func(t *testing.T) {
		if _, err := g.AllSTCuts(-1, 2); err == nil {
			t.Errorf("expected error for negative source")
		}
		if _, err := g.AllSTCuts(0, 10); err == nil {
			t.Errorf("expected error for out of bounds target")
		}
		if _, _, err := g.AllSTMincuts(-1, 2, nil); err == nil {
			t.Errorf("expected error for negative source")
		}
		if _, _, err := g.AllSTMincuts(0, 10, nil); err == nil {
			t.Errorf("expected error for out of bounds target")
		}
	})

	t.Run("source equals target", func(t *testing.T) {
		if _, err := g.AllSTCuts(1, 1); err == nil {
			t.Errorf("expected error when source == target")
		}
		if _, _, err := g.AllSTMincuts(1, 1, nil); err == nil {
			t.Errorf("expected error when source == target")
		}
	})

	t.Run("invalid capacities", func(t *testing.T) {
		if _, _, err := g.AllSTMincuts(0, 2, []float64{1.0}); err == nil {
			t.Errorf("expected error for invalid capacity length")
		}
		if _, _, err := g.AllSTMincuts(0, 2, []float64{1.0, -1.0}); err == nil {
			t.Errorf("expected error for negative capacity")
		}
		if _, _, err := g.AllSTMincuts(0, 2, []float64{1.0, math.NaN()}); err == nil {
			t.Errorf("expected error for NaN capacity")
		}
	})

	t.Run("closed and nil graph", func(t *testing.T) {
		var nilGraph *igraph.Graph
		if _, err := nilGraph.AllSTCuts(0, 1); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for nil graph AllSTCuts")
		}
		if _, _, err := nilGraph.AllSTMincuts(0, 1, nil); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for nil graph AllSTMincuts")
		}

		g.Close()
		if _, err := g.AllSTCuts(0, 1); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for closed graph AllSTCuts")
		}
		if _, _, err := g.AllSTMincuts(0, 1, nil); !errors.Is(err, igraph.ErrClosed) {
			t.Errorf("expected ErrClosed for closed graph AllSTMincuts")
		}
	})
}

package igraph_test

import (
	"math"
	"reflect"
	"sort"
	"testing"

	"github.com/h8gi/go-igraph"
)

func TestMaxFlowAndMinCut(t *testing.T) {
	// Create a simple directed flow network graph:
	// Vertices: 0 (source), 1, 2, 3 (target)
	// Edges:
	// 0 -> 1 (cap: 10.0)
	// 0 -> 2 (cap: 5.0)
	// 1 -> 2 (cap: 15.0)
	// 1 -> 3 (cap: 10.0)
	// 2 -> 3 (cap: 10.0)
	edges := []igraph.Edge{
		{From: 0, To: 1},
		{From: 0, To: 2},
		{From: 1, To: 2},
		{From: 1, To: 3},
		{From: 2, To: 3},
	}
	g, err := igraph.NewGraphFromEdges(4, edges, true)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}
	defer g.Close()

	capacities := []float64{10.0, 5.0, 15.0, 10.0, 10.0}

	t.Run("MaxFlow with custom capacities", func(t *testing.T) {
		res, err := g.MaxFlow(0, 3, capacities)
		if err != nil {
			t.Fatalf("MaxFlow failed: %v", err)
		}
		if res.Value != 15.0 {
			t.Errorf("expected max flow 15.0, got %v", res.Value)
		}
		numEdges, err := g.EdgeCount()
		if err != nil {
			t.Fatalf("EdgeCount failed: %v", err)
		}
		if len(res.Flow) != numEdges {
			t.Errorf("expected flow length %d, got %d", numEdges, len(res.Flow))
		}
		if res.Cut == nil || res.Partition == nil || res.Partition2 == nil {
			t.Errorf("expected non-nil result slices in MaxFlowResult")
		}

		val, err := g.MaxFlowValue(0, 3, capacities)
		if err != nil {
			t.Fatalf("MaxFlowValue failed: %v", err)
		}
		if val != 15.0 {
			t.Errorf("expected max flow value 15.0, got %v", val)
		}
	})

	t.Run("MaxFlow with nil capacities (unit capacities)", func(t *testing.T) {
		res, err := g.MaxFlow(0, 3, nil)
		if err != nil {
			t.Fatalf("MaxFlow failed: %v", err)
		}
		if res.Value != 2.0 {
			t.Errorf("expected max flow 2.0, got %v", res.Value)
		}
	})

	t.Run("MaxFlow reverse direction zero flow", func(t *testing.T) {
		res, err := g.MaxFlow(3, 0, capacities)
		if err != nil {
			t.Fatalf("MaxFlow reverse failed: %v", err)
		}
		if res.Value != 0 {
			t.Errorf("expected reverse max flow 0, got %v", res.Value)
		}
	})

	t.Run("STMinCut with custom and nil capacities", func(t *testing.T) {
		res, err := g.STMinCut(0, 3, capacities)
		if err != nil {
			t.Fatalf("STMinCut failed: %v", err)
		}
		if res.Value != 15.0 {
			t.Errorf("expected STMinCut value 15.0, got %v", res.Value)
		}

		val, err := g.STMinCutValue(0, 3, capacities)
		if err != nil {
			t.Fatalf("STMinCutValue failed: %v", err)
		}
		if val != 15.0 {
			t.Errorf("expected STMinCutValue 15.0, got %v", val)
		}

		// nil capacities
		resNil, err := g.STMinCut(0, 3, nil)
		if err != nil {
			t.Fatalf("STMinCut nil failed: %v", err)
		}
		if resNil.Value != 2.0 {
			t.Errorf("expected STMinCut nil value 2.0, got %v", resNil.Value)
		}

		valNil, err := g.STMinCutValue(0, 3, nil)
		if err != nil {
			t.Fatalf("STMinCutValue nil failed: %v", err)
		}
		if valNil != 2.0 {
			t.Errorf("expected STMinCutValue nil 2.0, got %v", valNil)
		}
	})

	t.Run("MinCut global on directed graph", func(t *testing.T) {
		res, err := g.MinCut(capacities)
		if err != nil {
			t.Fatalf("MinCut failed: %v", err)
		}
		if res.Value != 0 {
			t.Errorf("expected mincut 0 for sink vertex, got %v", res.Value)
		}

		val, err := g.MinCutValue(capacities)
		if err != nil {
			t.Fatalf("MinCutValue failed: %v", err)
		}
		if val != 0 {
			t.Errorf("expected MinCutValue 0, got %v", val)
		}
	})

	t.Run("MinCut global on undirected graph", func(t *testing.T) {
		undirectedEdges := []igraph.Edge{
			{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 0},
			{From: 2, To: 3}, // bridge
			{From: 3, To: 4}, {From: 4, To: 5}, {From: 5, To: 3},
		}
		ug, err := igraph.NewGraphFromEdges(6, undirectedEdges, false)
		if err != nil {
			t.Fatalf("failed to create undirected graph: %v", err)
		}
		defer ug.Close()

		res, err := ug.MinCut(nil)
		if err != nil {
			t.Fatalf("MinCut failed: %v", err)
		}
		if res.Value != 1.0 {
			t.Errorf("expected global mincut value 1.0 (bridge), got %v", res.Value)
		}

		// Verify partition sets cover all 6 vertices
		allVertices := append([]int(nil), res.Partition...)
		allVertices = append(allVertices, res.Partition2...)
		sort.Ints(allVertices)
		if !reflect.DeepEqual(allVertices, []int{0, 1, 2, 3, 4, 5}) {
			t.Errorf("partition sets do not partition graph vertices: got %v", allVertices)
		}

		val, err := ug.MinCutValue(nil)
		if err != nil {
			t.Fatalf("MinCutValue failed: %v", err)
		}
		if val != 1.0 {
			t.Errorf("MinCutValue %v mismatch, expected 1.0", val)
		}

		// Custom capacities on undirected graph
		uCaps := []float64{10, 10, 10, 3.5, 10, 10, 10}
		resCaps, err := ug.MinCut(uCaps)
		if err != nil {
			t.Fatalf("MinCut custom capacities failed: %v", err)
		}
		if resCaps.Value != 3.5 {
			t.Errorf("expected custom mincut 3.5, got %v", resCaps.Value)
		}

		valCaps, err := ug.MinCutValue(uCaps)
		if err != nil {
			t.Fatalf("MinCutValue custom capacities failed: %v", err)
		}
		if valCaps != 3.5 {
			t.Errorf("expected custom mincut value 3.5, got %v", valCaps)
		}
	})
}

func TestMaxFlowFullGraph(t *testing.T) {
	edges := []igraph.Edge{
		{From: 0, To: 1}, {From: 0, To: 2}, {From: 0, To: 3},
		{From: 1, To: 0}, {From: 1, To: 2}, {From: 1, To: 3},
		{From: 2, To: 0}, {From: 2, To: 1}, {From: 2, To: 3},
		{From: 3, To: 0}, {From: 3, To: 1}, {From: 3, To: 2},
	}
	g, err := igraph.NewGraphFromEdges(4, edges, true)
	if err != nil {
		t.Fatalf("failed to create full graph: %v", err)
	}
	defer g.Close()

	flow, err := g.MaxFlow(0, 3, nil)
	if err != nil {
		t.Fatalf("MaxFlow on full graph failed: %v", err)
	}
	if flow.Value != 3.0 {
		t.Errorf("expected max flow 3.0, got %v", flow.Value)
	}

	mc, err := g.MinCut(nil)
	if err != nil {
		t.Fatalf("MinCut on full graph failed: %v", err)
	}
	if mc.Value != 3.0 {
		t.Errorf("expected mincut 3.0, got %v", mc.Value)
	}
}

func TestMaxFlowDisconnectedGraph(t *testing.T) {
	edges := []igraph.Edge{
		{From: 0, To: 1},
		{From: 2, To: 3},
	}
	g, err := igraph.NewGraphFromEdges(4, edges, true)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}
	defer g.Close()

	res, err := g.MaxFlow(0, 3, nil)
	if err != nil {
		t.Fatalf("MaxFlow on disconnected graph failed: %v", err)
	}
	if res.Value != 0 {
		t.Errorf("expected max flow 0, got %v", res.Value)
	}

	val, err := g.MaxFlowValue(0, 3, nil)
	if err != nil {
		t.Fatalf("MaxFlowValue failed: %v", err)
	}
	if val != 0 {
		t.Errorf("expected 0, got %v", val)
	}

	stRes, err := g.STMinCut(0, 3, nil)
	if err != nil {
		t.Fatalf("STMinCut failed: %v", err)
	}
	if stRes.Value != 0 {
		t.Errorf("expected STMinCut 0, got %v", stRes.Value)
	}

	stVal, err := g.STMinCutValue(0, 3, nil)
	if err != nil {
		t.Fatalf("STMinCutValue failed: %v", err)
	}
	if stVal != 0 {
		t.Errorf("expected 0, got %v", stVal)
	}

	mcRes, err := g.MinCut(nil)
	if err != nil {
		t.Fatalf("MinCut failed: %v", err)
	}
	if mcRes.Value != 0 {
		t.Errorf("expected MinCut 0, got %v", mcRes.Value)
	}

	mcVal, err := g.MinCutValue(nil)
	if err != nil {
		t.Fatalf("MinCutValue failed: %v", err)
	}
	if mcVal != 0 {
		t.Errorf("expected 0, got %v", mcVal)
	}
}

func TestMaxFlowAndMinCutValidation(t *testing.T) {
	edges := []igraph.Edge{
		{From: 0, To: 1},
		{From: 1, To: 2},
	}
	g, err := igraph.NewGraphFromEdges(3, edges, true)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}
	defer g.Close()

	t.Run("invalid capacity length", func(t *testing.T) {
		funcs := []func() error{
			func() error { _, err := g.MaxFlow(0, 2, []float64{1.0}); return err },
			func() error { _, err := g.MaxFlowValue(0, 2, []float64{1.0}); return err },
			func() error { _, err := g.STMinCut(0, 2, []float64{1.0}); return err },
			func() error { _, err := g.STMinCutValue(0, 2, []float64{1.0}); return err },
			func() error { _, err := g.MinCut([]float64{1.0}); return err },
			func() error { _, err := g.MinCutValue([]float64{1.0}); return err },
		}
		for i, fn := range funcs {
			if err := fn(); err == nil {
				t.Errorf("func %d: expected error for invalid capacity length", i)
			}
		}
	})

	t.Run("negative capacity", func(t *testing.T) {
		funcs := []func() error{
			func() error { _, err := g.MaxFlow(0, 2, []float64{1.0, -2.0}); return err },
			func() error { _, err := g.MaxFlowValue(0, 2, []float64{1.0, -2.0}); return err },
			func() error { _, err := g.STMinCut(0, 2, []float64{1.0, -2.0}); return err },
			func() error { _, err := g.STMinCutValue(0, 2, []float64{1.0, -2.0}); return err },
			func() error { _, err := g.MinCut([]float64{1.0, -2.0}); return err },
			func() error { _, err := g.MinCutValue([]float64{1.0, -2.0}); return err },
		}
		for i, fn := range funcs {
			if err := fn(); err == nil {
				t.Errorf("func %d: expected error for negative capacity", i)
			}
		}
	})

	t.Run("NaN capacity", func(t *testing.T) {
		funcs := []func() error{
			func() error { _, err := g.MaxFlow(0, 2, []float64{1.0, math.NaN()}); return err },
			func() error { _, err := g.MaxFlowValue(0, 2, []float64{1.0, math.NaN()}); return err },
			func() error { _, err := g.STMinCut(0, 2, []float64{1.0, math.NaN()}); return err },
			func() error { _, err := g.STMinCutValue(0, 2, []float64{1.0, math.NaN()}); return err },
			func() error { _, err := g.MinCut([]float64{1.0, math.NaN()}); return err },
			func() error { _, err := g.MinCutValue([]float64{1.0, math.NaN()}); return err },
		}
		for i, fn := range funcs {
			if err := fn(); err == nil {
				t.Errorf("func %d: expected error for NaN capacity", i)
			}
		}
	})

	t.Run("out of bounds source", func(t *testing.T) {
		funcs := []func() error{
			func() error { _, err := g.MaxFlow(-1, 2, nil); return err },
			func() error { _, err := g.MaxFlow(10, 2, nil); return err },
			func() error { _, err := g.MaxFlowValue(-1, 2, nil); return err },
			func() error { _, err := g.MaxFlowValue(10, 2, nil); return err },
			func() error { _, err := g.STMinCut(-1, 2, nil); return err },
			func() error { _, err := g.STMinCut(10, 2, nil); return err },
			func() error { _, err := g.STMinCutValue(-1, 2, nil); return err },
			func() error { _, err := g.STMinCutValue(10, 2, nil); return err },
		}
		for i, fn := range funcs {
			if err := fn(); err == nil {
				t.Errorf("func %d: expected error for out of bounds source", i)
			}
		}
	})

	t.Run("out of bounds target", func(t *testing.T) {
		funcs := []func() error{
			func() error { _, err := g.MaxFlow(0, -1, nil); return err },
			func() error { _, err := g.MaxFlow(0, 10, nil); return err },
			func() error { _, err := g.MaxFlowValue(0, -1, nil); return err },
			func() error { _, err := g.MaxFlowValue(0, 10, nil); return err },
			func() error { _, err := g.STMinCut(0, -1, nil); return err },
			func() error { _, err := g.STMinCut(0, 10, nil); return err },
			func() error { _, err := g.STMinCutValue(0, -1, nil); return err },
			func() error { _, err := g.STMinCutValue(0, 10, nil); return err },
		}
		for i, fn := range funcs {
			if err := fn(); err == nil {
				t.Errorf("func %d: expected error for out of bounds target", i)
			}
		}
	})

	t.Run("source equals target", func(t *testing.T) {
		funcs := []func() error{
			func() error { _, err := g.MaxFlow(1, 1, nil); return err },
			func() error { _, err := g.MaxFlowValue(1, 1, nil); return err },
			func() error { _, err := g.STMinCut(1, 1, nil); return err },
			func() error { _, err := g.STMinCutValue(1, 1, nil); return err },
		}
		for i, fn := range funcs {
			if err := fn(); err == nil {
				t.Errorf("func %d: expected error when source == target", i)
			}
		}
	})
}

func TestMaxFlowNilAndClosedGraph(t *testing.T) {
	var nilGraph *igraph.Graph
	if _, err := nilGraph.MaxFlow(0, 1, nil); err != igraph.ErrClosed {
		t.Errorf("expected ErrClosed for nil Graph.MaxFlow, got %v", err)
	}
	if _, err := nilGraph.MaxFlowValue(0, 1, nil); err != igraph.ErrClosed {
		t.Errorf("expected ErrClosed for nil Graph.MaxFlowValue, got %v", err)
	}
	if _, err := nilGraph.STMinCut(0, 1, nil); err != igraph.ErrClosed {
		t.Errorf("expected ErrClosed for nil Graph.STMinCut, got %v", err)
	}
	if _, err := nilGraph.STMinCutValue(0, 1, nil); err != igraph.ErrClosed {
		t.Errorf("expected ErrClosed for nil Graph.STMinCutValue, got %v", err)
	}
	if _, err := nilGraph.MinCut(nil); err != igraph.ErrClosed {
		t.Errorf("expected ErrClosed for nil Graph.MinCut, got %v", err)
	}
	if _, err := nilGraph.MinCutValue(nil); err != igraph.ErrClosed {
		t.Errorf("expected ErrClosed for nil Graph.MinCutValue, got %v", err)
	}

	edges := []igraph.Edge{
		{From: 0, To: 1},
	}
	g, err := igraph.NewGraphFromEdges(2, edges, true)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}

	res, err := g.MaxFlow(0, 1, nil)
	if err != nil {
		t.Fatalf("MaxFlow failed: %v", err)
	}

	g.Close()

	// Operations on closed graph should fail
	if _, err := g.MaxFlow(0, 1, nil); err != igraph.ErrClosed {
		t.Errorf("expected ErrClosed for MaxFlow after Close, got %v", err)
	}
	if _, err := g.MaxFlowValue(0, 1, nil); err != igraph.ErrClosed {
		t.Errorf("expected ErrClosed for MaxFlowValue after Close, got %v", err)
	}
	if _, err := g.MinCut(nil); err != igraph.ErrClosed {
		t.Errorf("expected ErrClosed for MinCut after Close, got %v", err)
	}
	if _, err := g.MinCutValue(nil); err != igraph.ErrClosed {
		t.Errorf("expected ErrClosed for MinCutValue after Close, got %v", err)
	}
	if _, err := g.STMinCut(0, 1, nil); err != igraph.ErrClosed {
		t.Errorf("expected ErrClosed for STMinCut after Close, got %v", err)
	}
	if _, err := g.STMinCutValue(0, 1, nil); err != igraph.ErrClosed {
		t.Errorf("expected ErrClosed for STMinCutValue after Close, got %v", err)
	}

	// Slices obtained prior to Close should remain valid
	if len(res.Flow) != 1 {
		t.Errorf("flow slice should remain valid after graph Close")
	}
}

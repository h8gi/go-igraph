package igraph

import (
	"errors"
	"math"
	"strings"
	"sync"
	"testing"
)

func TestEdgeConvergenceDegreeDirectedAndOwnership(t *testing.T) {
	graph, err := NewPath(3, true, false)
	if err != nil {
		t.Fatal(err)
	}
	result, err := graph.EdgeConvergenceDegree()
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, result.InputSetSizes, []float64{1, 2})
	assertFloatSlice(t, result.OutputSetSizes, []float64{2, 1})
	assertFloatSlice(t, result.Convergence, []float64{-1.0 / 3, 1.0 / 3})
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, result.Convergence, []float64{-1.0 / 3, 1.0 / 3})
}

func TestEdgeConvergenceDegreeUndirectedEmptyLoopsAndParallelEdges(t *testing.T) {
	t.Run("undirected", func(t *testing.T) {
		graph, err := NewPath(4, false, false)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = graph.Close() })
		result, err := graph.EdgeConvergenceDegree()
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSlice(t, result.InputSetSizes, []float64{1, 2, 3})
		assertFloatSlice(t, result.OutputSetSizes, []float64{3, 2, 1})
		assertFloatSlice(t, result.Convergence, []float64{0.5, 0, 0.5})
	})

	t.Run("empty and edgeless", func(t *testing.T) {
		for _, vertexCount := range []int{0, 3} {
			graph, err := NewGraphFromEdges(vertexCount, nil, false)
			if err != nil {
				t.Fatal(err)
			}
			result, err := graph.EdgeConvergenceDegree()
			_ = graph.Close()
			if err != nil {
				t.Fatal(err)
			}
			if result.Convergence == nil || result.InputSetSizes == nil || result.OutputSetSizes == nil ||
				len(result.Convergence) != 0 || len(result.InputSetSizes) != 0 || len(result.OutputSetSizes) != 0 {
				t.Errorf("edgeless result = %#v", result)
			}
		}
	})

	t.Run("loop and parallel", func(t *testing.T) {
		graph, err := NewGraphFromEdges(2, []Edge{{0, 0}, {0, 1}, {0, 1}}, false)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = graph.Close() })
		result, err := graph.EdgeConvergenceDegree()
		if err != nil {
			t.Fatal(err)
		}
		if !math.IsNaN(result.Convergence[0]) || result.InputSetSizes[0] != 0 || result.OutputSetSizes[0] != 0 {
			t.Errorf("loop result = %#v", result)
		}
		if len(result.Convergence) != 3 {
			t.Fatalf("parallel result length = %d", len(result.Convergence))
		}
	})
}

func TestEdgeConvergenceDegreeErrorsAndConcurrency(t *testing.T) {
	if _, err := collectConvergenceDegree(1, func(result, ins, outs *realVector) int {
		return 2
	}); err == nil || !strings.Contains(err.Error(), "edge convergence") {
		t.Errorf("upstream error = %v", err)
	}
	if _, err := collectConvergenceDegree(1, func(result, ins, outs *realVector) int {
		return 0
	}); err == nil || !strings.Contains(err.Error(), "lengths") {
		t.Errorf("alignment error = %v", err)
	}

	graph, err := NewPath(20, false, false)
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for index := 0; index < 8; index++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if _, err := graph.EdgeConvergenceDegree(); err != nil {
				t.Errorf("concurrent call: %v", err)
			}
		}()
	}
	group.Wait()
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := graph.EdgeConvergenceDegree(); !errors.Is(err, ErrClosed) {
		t.Errorf("call after Close error = %v", err)
	}
	var nilGraph *Graph
	if _, err := nilGraph.EdgeConvergenceDegree(); !errors.Is(err, ErrClosed) {
		t.Errorf("nil graph error = %v", err)
	}
}

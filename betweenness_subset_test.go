package igraph

import (
	"errors"
	"math"
	"sync"
	"testing"
)

func TestBetweennessSubsetKnownAnswersSetSemanticsAndResultOrder(t *testing.T) {
	graph, err := NewPath(4, false, false)
	if err != nil {
		t.Fatal(err)
	}
	vertices, _ := VertexIDs(2, 0, 2, 1)
	edges, _ := EdgeIDs(2, 0, 2, 1)
	sources, _ := VertexIDs(0, 0)
	targets, _ := VertexIDs(3, 3)

	vertexScores, err := graph.VertexBetweennessSubset(vertices, sources, targets, SubsetBetweennessOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, vertexScores, []float64{0.5, 0, 0.5, 0.5})
	edgeScores, err := graph.EdgeBetweennessSubset(edges, sources, targets, SubsetBetweennessOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, edgeScores, []float64{0.5, 0.5, 0.5, 0.5})

	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	assertFloatSlice(t, vertexScores, []float64{0.5, 0, 0.5, 0.5})
	assertFloatSlice(t, edgeScores, []float64{0.5, 0.5, 0.5, 0.5})
}

func TestBetweennessSubsetDirectedWeightedEmptyAndDisconnected(t *testing.T) {
	t.Run("directed", func(t *testing.T) {
		graph, err := NewGraphFromEdges(3, []Edge{{0, 1}, {2, 1}}, true)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = graph.Close() })
		sources, _ := VertexIDs(0)
		targets, _ := VertexIDs(2)
		directed, err := graph.VertexBetweennessSubset(AllVertices(), sources, targets, SubsetBetweennessOptions{DirectedPaths: true})
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSlice(t, directed, []float64{0, 0, 0})
		undirected, err := graph.VertexBetweennessSubset(AllVertices(), sources, targets, SubsetBetweennessOptions{})
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSlice(t, undirected, []float64{0, 0.5, 0})
	})

	t.Run("weighted", func(t *testing.T) {
		graph, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}, {0, 2}}, false)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = graph.Close() })
		sources, _ := VertexIDs(0)
		targets, _ := VertexIDs(2)
		options := SubsetBetweennessOptions{Weights: []float64{1, 1, 3}}
		vertices, err := graph.VertexBetweennessSubset(AllVertices(), sources, targets, options)
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSlice(t, vertices, []float64{0, 0.5, 0})
		edges, err := graph.EdgeBetweennessSubset(AllEdges(), sources, targets, options)
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSlice(t, edges, []float64{0.5, 0.5, 0})
	})

	t.Run("empty and disconnected", func(t *testing.T) {
		graph, err := NewGraphFromEdges(4, []Edge{{0, 1}, {2, 3}}, false)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = graph.Close() })
		targets, _ := VertexIDs(3)
		vertices, err := graph.VertexBetweennessSubset(NoVertices(), NoVertices(), targets, SubsetBetweennessOptions{})
		if err != nil || vertices == nil || len(vertices) != 0 {
			t.Fatalf("empty result = %#v, %v", vertices, err)
		}
		edges, err := graph.EdgeBetweennessSubset(AllEdges(), NoVertices(), targets, SubsetBetweennessOptions{})
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSlice(t, edges, []float64{0, 0})

		empty, err := NewGraph()
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = empty.Close() })
		vertices, err = empty.VertexBetweennessSubset(AllVertices(), AllVertices(), AllVertices(), SubsetBetweennessOptions{})
		if err != nil || vertices == nil || len(vertices) != 0 {
			t.Fatalf("empty graph result = %#v, %v", vertices, err)
		}
	})
}

func TestBetweennessSubsetRejectsInvalidInputsAndClosedGraph(t *testing.T) {
	graph, err := NewPath(3, false, false)
	if err != nil {
		t.Fatal(err)
	}
	badVertex := VertexSelector{kind: vertexSelectorIDs, ids: []int{3}}
	badEdge := EdgeSelector{kind: edgeSelectorIDs, ids: []int{2}}
	if _, err := graph.VertexBetweennessSubset(AllVertices(), badVertex, AllVertices(), SubsetBetweennessOptions{}); err == nil {
		t.Error("invalid source selector error = nil")
	}
	if _, err := graph.VertexBetweennessSubset(badVertex, AllVertices(), AllVertices(), SubsetBetweennessOptions{}); err == nil {
		t.Error("invalid result selector error = nil")
	}
	if _, err := graph.EdgeBetweennessSubset(badEdge, AllVertices(), AllVertices(), SubsetBetweennessOptions{}); err == nil {
		t.Error("invalid edge selector error = nil")
	}
	for _, weights := range [][]float64{{}, {1}, {1, 0}, {1, math.NaN()}, {1, math.Inf(1)}} {
		options := SubsetBetweennessOptions{Weights: weights}
		if _, err := graph.VertexBetweennessSubset(AllVertices(), AllVertices(), AllVertices(), options); err == nil {
			t.Errorf("invalid weights %#v error = nil", weights)
		}
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := graph.VertexBetweennessSubset(AllVertices(), AllVertices(), AllVertices(), SubsetBetweennessOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("vertex call after Close error = %v", err)
	}
	if _, err := graph.EdgeBetweennessSubset(AllEdges(), AllVertices(), AllVertices(), SubsetBetweennessOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("edge call after Close error = %v", err)
	}
	var nilGraph *Graph
	if _, err := nilGraph.VertexBetweennessSubset(AllVertices(), AllVertices(), AllVertices(), SubsetBetweennessOptions{}); !errors.Is(err, ErrClosed) {
		t.Errorf("nil graph error = %v", err)
	}
}

func TestBetweennessSubsetConcurrentReads(t *testing.T) {
	graph, err := NewPath(20, false, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	var group sync.WaitGroup
	for index := 0; index < 8; index++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if _, err := graph.VertexBetweennessSubset(AllVertices(), AllVertices(), AllVertices(), SubsetBetweennessOptions{}); err != nil {
				t.Errorf("concurrent vertex call: %v", err)
			}
			if _, err := graph.EdgeBetweennessSubset(AllEdges(), AllVertices(), AllVertices(), SubsetBetweennessOptions{}); err != nil {
				t.Errorf("concurrent edge call: %v", err)
			}
		}()
	}
	group.Wait()
}

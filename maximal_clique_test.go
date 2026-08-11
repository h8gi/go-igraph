package igraph

import (
	"errors"
	"reflect"
	"testing"
)

func TestMaximalCliquesCountsAndHistogram(t *testing.T) {
	graph := testCliqueGraph(t)
	defer graph.Close()
	result, err := graph.MaximalCliques(VertexSetEnumerationOptions{MaxResults: 2})
	if err != nil || result.Truncated || !reflect.DeepEqual(sortedVertexSets(result.Sets), [][]int{{3, 4}, {0, 1, 2}}) {
		t.Fatalf("MaximalCliques = %#v, %v", result, err)
	}
	limited, err := graph.MaximalCliques(VertexSetEnumerationOptions{MaxResults: 1})
	if err != nil || !limited.Truncated || len(limited.Sets) != 1 {
		t.Errorf("limited MaximalCliques = %#v, %v", limited, err)
	}
	count, err := graph.MaximalCliqueCount(VertexSetRange{})
	if err != nil || count != 2 {
		t.Errorf("MaximalCliqueCount = %d, %v", count, err)
	}
	histogram, err := graph.MaximalCliqueSizeHistogram(VertexSetRange{})
	if err != nil || !reflect.DeepEqual(histogram, []int{0, 1, 1}) {
		t.Errorf("MaximalCliqueSizeHistogram = %v, %v", histogram, err)
	}
	minimum, maximum := 3, 3
	count, err = graph.MaximalCliqueCount(VertexSetRange{Minimum: &minimum, Maximum: &maximum})
	if err != nil || count != 1 {
		t.Errorf("bounded MaximalCliqueCount = %d, %v", count, err)
	}
}

func TestMaximalCliquesFromInitialVertices(t *testing.T) {
	graph := testCliqueGraph(t)
	defer graph.Close()
	result, err := graph.MaximalCliquesFromVertices([]int{0}, VertexSetEnumerationOptions{MaxResults: 1})
	if err != nil || result.Truncated || !reflect.DeepEqual(result.Sets, [][]int{{3, 4}}) {
		t.Fatalf("initial vertex result = %#v, %v", result, err)
	}
	// The result need not contain the initial vertex, proving that this is an
	// internal search-root partition rather than an induced-subgraph filter or
	// a request for cliques containing the supplied IDs.
	empty, err := graph.MaximalCliquesFromVertices([]int{}, VertexSetEnumerationOptions{MaxResults: 1})
	if err != nil || empty.Sets == nil || len(empty.Sets) != 0 || empty.Truncated {
		t.Errorf("empty initial vertices = %#v, %v", empty, err)
	}
	if _, err := graph.MaximalCliquesFromVertices([]int{0, 0}, VertexSetEnumerationOptions{MaxResults: 1}); err == nil {
		t.Error("expected duplicate initial vertex error")
	}
	if _, err := graph.MaximalCliquesFromVertices([]int{5}, VertexSetEnumerationOptions{MaxResults: 1}); err == nil {
		t.Error("expected invalid initial vertex error")
	}
}

func TestMaximalCliqueGraphShapesAndClosure(t *testing.T) {
	var nilGraph *Graph
	nilQueries := []func() error{
		func() error {
			_, err := nilGraph.MaximalCliques(VertexSetEnumerationOptions{MaxResults: 1})
			return err
		},
		func() error {
			_, err := nilGraph.MaximalCliquesFromVertices([]int{0}, VertexSetEnumerationOptions{MaxResults: 1})
			return err
		},
		func() error { _, err := nilGraph.MaximalCliqueCount(VertexSetRange{}); return err },
		func() error { _, err := nilGraph.MaximalCliqueSizeHistogram(VertexSetRange{}); return err },
	}
	for index, query := range nilQueries {
		if err := query(); !errors.Is(err, ErrClosed) {
			t.Errorf("nil query %d error = %v", index, err)
		}
	}

	graph, err := NewGraphFromEdges(3, []Edge{{0, 0}, {0, 1}, {0, 1}, {1, 2}, {2, 0}}, true)
	if err != nil {
		t.Fatal(err)
	}
	result, err := graph.MaximalCliques(VertexSetEnumerationOptions{MaxResults: 1})
	if err != nil || result.Truncated || !reflect.DeepEqual(result.Sets, [][]int{{0, 1, 2}}) {
		t.Errorf("directed multigraph maximal cliques = %#v, %v", result, err)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := graph.MaximalCliques(VertexSetEnumerationOptions{MaxResults: 1}); !errors.Is(err, ErrClosed) {
		t.Errorf("closed MaximalCliques error = %v", err)
	}
	if _, err := graph.MaximalCliquesFromVertices([]int{0}, VertexSetEnumerationOptions{MaxResults: 1}); !errors.Is(err, ErrClosed) {
		t.Errorf("closed subset error = %v", err)
	}
	if _, err := graph.MaximalCliqueCount(VertexSetRange{}); !errors.Is(err, ErrClosed) {
		t.Errorf("closed count error = %v", err)
	}
	if _, err := graph.MaximalCliqueSizeHistogram(VertexSetRange{}); !errors.Is(err, ErrClosed) {
		t.Errorf("closed histogram error = %v", err)
	}
}

func TestMaximalCliqueValidation(t *testing.T) {
	graph := testCliqueGraph(t)
	defer graph.Close()
	if _, err := graph.MaximalCliques(VertexSetEnumerationOptions{}); err == nil {
		t.Error("expected invalid maximal-clique limit")
	}
	minimum, maximum := 3, 2
	invalid := VertexSetRange{Minimum: &minimum, Maximum: &maximum}
	if _, err := graph.MaximalCliqueCount(invalid); err == nil {
		t.Error("expected invalid count range")
	}
	if _, err := graph.MaximalCliqueSizeHistogram(invalid); err == nil {
		t.Error("expected invalid histogram range")
	}
	if _, err := graph.MaximalCliquesFromVertices([]int{-1}, VertexSetEnumerationOptions{MaxResults: 1}); err == nil {
		t.Error("expected negative initial vertex error")
	}

	empty, err := NewGraphFromEdges(0, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	defer empty.Close()
	result, err := empty.MaximalCliques(VertexSetEnumerationOptions{MaxResults: 1})
	if err != nil || result.Sets == nil || len(result.Sets) != 0 || result.Truncated {
		t.Errorf("empty maximal cliques = %#v, %v", result, err)
	}
	count, err := empty.MaximalCliqueCount(VertexSetRange{})
	if err != nil || count != 0 {
		t.Errorf("empty maximal count = %d, %v", count, err)
	}
	histogram, err := empty.MaximalCliqueSizeHistogram(VertexSetRange{})
	if err != nil || histogram == nil || len(histogram) != 0 {
		t.Errorf("empty maximal histogram = %v, %v", histogram, err)
	}
}

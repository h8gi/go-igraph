package igraph

import (
	"errors"
	"reflect"
	"testing"
)

func TestIndependentVertexSetsBoundsAndTruncation(t *testing.T) {
	graph, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()

	all, err := graph.IndependentVertexSets(VertexSetEnumerationOptions{MaxResults: 4})
	want := [][]int{{0}, {1}, {2}, {0, 2}}
	if err != nil || all.Truncated || !reflect.DeepEqual(sortedVertexSets(all.Sets), want) {
		t.Fatalf("IndependentVertexSets = %#v, %v", all, err)
	}
	limited, err := graph.IndependentVertexSets(VertexSetEnumerationOptions{MaxResults: 3})
	if err != nil || !limited.Truncated || len(limited.Sets) != 3 {
		t.Errorf("limited independent sets = %#v, %v", limited, err)
	}
	minimum, maximum := 2, 2
	pairs, err := graph.IndependentVertexSets(VertexSetEnumerationOptions{
		Range: VertexSetRange{Minimum: &minimum, Maximum: &maximum}, MaxResults: 1,
	})
	if err != nil || pairs.Truncated || !reflect.DeepEqual(pairs.Sets, [][]int{{0, 2}}) {
		t.Errorf("bounded independent sets = %#v, %v", pairs, err)
	}
	minimum, maximum = 3, 3
	none, err := graph.IndependentVertexSets(VertexSetEnumerationOptions{
		Range: VertexSetRange{Minimum: &minimum, Maximum: &maximum}, MaxResults: 1,
	})
	if err != nil || none.Sets == nil || len(none.Sets) != 0 || none.Truncated {
		t.Errorf("empty independent sets = %#v, %v", none, err)
	}
}

func TestMaximumVersusMaximalIndependentVertexSets(t *testing.T) {
	graph, err := NewGraphFromEdges(3, []Edge{{0, 1}, {1, 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()

	maximal, err := graph.MaximalIndependentVertexSets(VertexSetEnumerationOptions{MaxResults: 2})
	if err != nil || maximal.Truncated || !reflect.DeepEqual(sortedVertexSets(maximal.Sets), [][]int{{1}, {0, 2}}) {
		t.Fatalf("maximal independent sets = %#v, %v", maximal, err)
	}
	largest, err := graph.LargestIndependentVertexSets(1)
	if err != nil || largest.Truncated || !reflect.DeepEqual(largest.Sets, [][]int{{0, 2}}) {
		t.Errorf("largest independent sets = %#v, %v", largest, err)
	}
	minimum := 2
	filtered, err := graph.MaximalIndependentVertexSets(VertexSetEnumerationOptions{
		Range: VertexSetRange{Minimum: &minimum}, MaxResults: 1,
	})
	if err != nil || filtered.Truncated || !reflect.DeepEqual(filtered.Sets, [][]int{{0, 2}}) {
		t.Errorf("filtered maximal independent sets = %#v, %v", filtered, err)
	}
}

func TestLargestIndependentVertexSetTies(t *testing.T) {
	graph := testCliqueGraph(t)
	defer graph.Close()
	limited, err := graph.LargestIndependentVertexSets(5)
	if err != nil || !limited.Truncated || len(limited.Sets) != 5 {
		t.Fatalf("limited largest independent sets = %#v, %v", limited, err)
	}
	all, err := graph.LargestIndependentVertexSets(6)
	want := [][]int{{0, 3}, {0, 4}, {1, 3}, {1, 4}, {2, 3}, {2, 4}}
	if err != nil || all.Truncated || !reflect.DeepEqual(sortedVertexSets(all.Sets), want) {
		t.Errorf("largest independent sets = %#v, %v", all, err)
	}
}

func TestIndependentVertexSetGraphShapes(t *testing.T) {
	tests := []struct {
		name     string
		vertices int
		edges    []Edge
		directed bool
		want     [][]int
	}{
		{name: "complete", vertices: 3, edges: []Edge{{0, 1}, {0, 2}, {1, 2}}, want: [][]int{{0}, {1}, {2}}},
		{name: "edgeless", vertices: 3, want: [][]int{{0, 1, 2}}},
		{name: "disconnected", vertices: 4, edges: []Edge{{0, 1}, {2, 3}}, want: [][]int{{0, 2}, {0, 3}, {1, 2}, {1, 3}}},
		{name: "directed loops and parallels", vertices: 3, directed: true, edges: []Edge{{0, 0}, {0, 1}, {0, 1}}, want: [][]int{{0, 2}, {1, 2}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph, err := NewGraphFromEdges(tt.vertices, tt.edges, tt.directed)
			if err != nil {
				t.Fatal(err)
			}
			defer graph.Close()
			got, err := graph.LargestIndependentVertexSets(len(tt.want))
			if err != nil || got.Truncated || !reflect.DeepEqual(sortedVertexSets(got.Sets), tt.want) {
				t.Errorf("LargestIndependentVertexSets = %#v, %v", got, err)
			}
		})
	}

	empty, err := NewGraphFromEdges(0, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	defer empty.Close()
	for name, result := range map[string]func() (VertexSetEnumeration, error){
		"all": func() (VertexSetEnumeration, error) {
			return empty.IndependentVertexSets(VertexSetEnumerationOptions{MaxResults: 1})
		},
		"maximal": func() (VertexSetEnumeration, error) {
			return empty.MaximalIndependentVertexSets(VertexSetEnumerationOptions{MaxResults: 1})
		},
		"largest": func() (VertexSetEnumeration, error) { return empty.LargestIndependentVertexSets(1) },
	} {
		got, err := result()
		if err != nil || got.Sets == nil || len(got.Sets) != 0 || got.Truncated {
			t.Errorf("empty %s = %#v, %v", name, got, err)
		}
	}
}

func TestIndependentVertexSetsMatchComplementCliques(t *testing.T) {
	graph := testCliqueGraph(t)
	defer graph.Close()
	complement, err := graph.Complement(false)
	if err != nil {
		t.Fatal(err)
	}
	defer complement.Graph.Close()
	independent, err := graph.IndependentVertexSets(VertexSetEnumerationOptions{MaxResults: 21})
	if err != nil {
		t.Fatal(err)
	}
	cliques, err := complement.Graph.Cliques(VertexSetEnumerationOptions{MaxResults: 21})
	if err != nil {
		t.Fatal(err)
	}
	if independent.Truncated || cliques.Truncated || !reflect.DeepEqual(sortedVertexSets(independent.Sets), sortedVertexSets(cliques.Sets)) {
		t.Errorf("independent sets %#v != complement cliques %#v", independent, cliques)
	}
}

func TestIndependentVertexSetEnumerationValidationAndClosure(t *testing.T) {
	var nilGraph *Graph
	queries := []func(*Graph) error{
		func(graph *Graph) error {
			_, err := graph.IndependentVertexSets(VertexSetEnumerationOptions{MaxResults: 1})
			return err
		},
		func(graph *Graph) error {
			_, err := graph.MaximalIndependentVertexSets(VertexSetEnumerationOptions{MaxResults: 1})
			return err
		},
		func(graph *Graph) error { _, err := graph.LargestIndependentVertexSets(1); return err },
	}
	for index, query := range queries {
		if err := query(nilGraph); !errors.Is(err, ErrClosed) {
			t.Errorf("nil query %d error = %v", index, err)
		}
	}
	graph := testCliqueGraph(t)
	if _, err := graph.IndependentVertexSets(VertexSetEnumerationOptions{}); err == nil {
		t.Error("expected invalid result limit")
	}
	if _, err := graph.MaximalIndependentVertexSets(VertexSetEnumerationOptions{}); err == nil {
		t.Error("expected invalid maximal result limit")
	}
	if _, err := graph.LargestIndependentVertexSets(0); err == nil {
		t.Error("expected invalid largest result limit")
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	for index, query := range queries {
		if err := query(graph); !errors.Is(err, ErrClosed) {
			t.Errorf("closed query %d error = %v", index, err)
		}
	}
}

package igraph

import (
	"errors"
	"math"
	"reflect"
	"sort"
	"testing"
)

func TestCliquesBoundsAndTruncation(t *testing.T) {
	graph := testCliqueGraph(t)
	defer graph.Close()

	all, err := graph.Cliques(VertexSetEnumerationOptions{MaxResults: 10})
	if err != nil {
		t.Fatalf("Cliques failed: %v", err)
	}
	if all.Truncated || len(all.Sets) != 10 {
		t.Fatalf("all cliques = %#v", all)
	}
	want := [][]int{{0}, {1}, {2}, {3}, {4}, {0, 1}, {0, 2}, {1, 2}, {3, 4}, {0, 1, 2}}
	if got := sortedVertexSets(all.Sets); !reflect.DeepEqual(got, sortedVertexSets(want)) {
		t.Errorf("Cliques = %v; want %v", got, want)
	}

	limited, err := graph.Cliques(VertexSetEnumerationOptions{MaxResults: 9})
	if err != nil || !limited.Truncated || len(limited.Sets) != 9 {
		t.Errorf("limited cliques = %#v, %v", limited, err)
	}
	exact, err := graph.Cliques(VertexSetEnumerationOptions{MaxResults: 10})
	if err != nil || exact.Truncated {
		t.Errorf("exact bound = %#v, %v", exact, err)
	}

	minimum, maximum := 2, 2
	edges, err := graph.Cliques(VertexSetEnumerationOptions{
		Range: VertexSetRange{Minimum: &minimum, Maximum: &maximum}, MaxResults: 4,
	})
	if err != nil || edges.Truncated || len(edges.Sets) != 4 {
		t.Errorf("size-two cliques = %#v, %v", edges, err)
	}
	for _, set := range edges.Sets {
		if len(set) != 2 || !sort.IntsAreSorted(set) {
			t.Errorf("non-canonical size-two clique: %v", set)
		}
	}
	minimum, maximum = 4, 4
	none, err := graph.Cliques(VertexSetEnumerationOptions{
		Range: VertexSetRange{Minimum: &minimum, Maximum: &maximum}, MaxResults: 1,
	})
	if err != nil || none.Truncated || none.Sets == nil || len(none.Sets) != 0 {
		t.Errorf("empty result = %#v, %v", none, err)
	}
}

func TestLargestCliquesAndOwnership(t *testing.T) {
	graph, err := NewGraphFromEdges(6, []Edge{
		{0, 1}, {0, 2}, {1, 2},
		{3, 4}, {3, 5}, {4, 5},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	limited, err := graph.LargestCliques(1)
	if err != nil || !limited.Truncated || len(limited.Sets) != 1 {
		t.Fatalf("limited largest cliques = %#v, %v", limited, err)
	}
	all, err := graph.LargestCliques(2)
	if err != nil || all.Truncated || len(all.Sets) != 2 {
		t.Fatalf("largest cliques = %#v, %v", all, err)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	all.Sets[0][0] = 99
	if all.Sets[1][0] == 99 {
		t.Error("largest clique results share backing storage")
	}
}

func TestCliqueEnumerationGraphShapes(t *testing.T) {
	empty, err := NewGraphFromEdges(0, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	defer empty.Close()
	got, err := empty.Cliques(VertexSetEnumerationOptions{MaxResults: 1})
	if err != nil || got.Sets == nil || len(got.Sets) != 0 || got.Truncated {
		t.Errorf("empty Cliques = %#v, %v", got, err)
	}
	largest, err := empty.LargestCliques(1)
	if err != nil || largest.Sets == nil || len(largest.Sets) != 0 || largest.Truncated {
		t.Errorf("empty LargestCliques = %#v, %v", largest, err)
	}

	multigraph, err := NewGraphFromEdges(3, []Edge{{0, 0}, {0, 1}, {0, 1}, {1, 2}, {2, 0}}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer multigraph.Close()
	triangles, err := multigraph.LargestCliques(2)
	if err != nil || triangles.Truncated || !reflect.DeepEqual(triangles.Sets, [][]int{{0, 1, 2}}) {
		t.Errorf("directed multigraph largest cliques = %#v, %v", triangles, err)
	}
}

func TestCliqueSizeHistogram(t *testing.T) {
	graph := testCliqueGraph(t)
	defer graph.Close()

	histogram, err := graph.CliqueSizeHistogram(VertexSetRange{})
	if err != nil || !reflect.DeepEqual(histogram, []int{5, 4, 1}) {
		t.Fatalf("histogram = %v, %v", histogram, err)
	}
	minimum, maximum := 2, 2
	histogram, err = graph.CliqueSizeHistogram(VertexSetRange{Minimum: &minimum, Maximum: &maximum})
	if err != nil || !reflect.DeepEqual(histogram, []int{0, 4}) {
		t.Errorf("bounded histogram = %v, %v", histogram, err)
	}

	empty, err := NewGraphFromEdges(0, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	defer empty.Close()
	histogram, err = empty.CliqueSizeHistogram(VertexSetRange{})
	if err != nil || histogram == nil || len(histogram) != 0 {
		t.Errorf("empty histogram = %v, %v", histogram, err)
	}
}

func TestCliqueEnumerationValidationAndClosure(t *testing.T) {
	graph := testCliqueGraph(t)
	if _, err := graph.Cliques(VertexSetEnumerationOptions{}); err == nil {
		t.Error("expected invalid result limit")
	}
	if _, err := graph.LargestCliques(0); err == nil {
		t.Error("expected invalid largest-clique limit")
	}
	minimum, maximum := 3, 2
	if _, err := graph.CliqueSizeHistogram(VertexSetRange{Minimum: &minimum, Maximum: &maximum}); err == nil {
		t.Error("expected invalid histogram range")
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := graph.Cliques(VertexSetEnumerationOptions{MaxResults: 1}); !errors.Is(err, ErrClosed) {
		t.Errorf("closed Cliques error = %v", err)
	}
	if _, err := graph.LargestCliques(1); !errors.Is(err, ErrClosed) {
		t.Errorf("closed LargestCliques error = %v", err)
	}
	if _, err := graph.CliqueSizeHistogram(VertexSetRange{}); !errors.Is(err, ErrClosed) {
		t.Errorf("closed CliqueSizeHistogram error = %v", err)
	}
}

func TestCliqueHistogramCountConversion(t *testing.T) {
	if got, err := cliqueHistogramCounts([]float64{0, 2, 3}); err != nil || !reflect.DeepEqual(got, []int{0, 2, 3}) {
		t.Errorf("valid counts = %v, %v", got, err)
	}
	for _, value := range []float64{-1, 0.5, math.NaN(), math.Inf(1), math.Inf(-1), math.MaxFloat64} {
		if _, err := cliqueHistogramCounts([]float64{value}); err == nil {
			t.Errorf("count %g unexpectedly valid", value)
		}
	}
}

func testCliqueGraph(t *testing.T) *Graph {
	t.Helper()
	graph, err := NewGraphFromEdges(5, []Edge{{0, 1}, {0, 2}, {1, 2}, {3, 4}}, false)
	if err != nil {
		t.Fatal(err)
	}
	return graph
}

func sortedVertexSets(sets [][]int) [][]int {
	result := make([][]int, len(sets))
	for i, set := range sets {
		result[i] = append([]int{}, set...)
		sort.Ints(result[i])
	}
	sort.Slice(result, func(i, j int) bool {
		if len(result[i]) != len(result[j]) {
			return len(result[i]) < len(result[j])
		}
		for k := range result[i] {
			if result[i][k] != result[j][k] {
				return result[i][k] < result[j][k]
			}
		}
		return false
	})
	return result
}

package igraph

import (
	"errors"
	"math/big"
	"reflect"
	"sort"
	"testing"
)

func TestCanonicalGraphsOfIsomorphicInputsMatch(t *testing.T) {
	left := testGraphFromEdges(t, 5, []Edge{{0, 1}, {1, 2}, {2, 3}, {3, 0}, {1, 4}}, false)
	right := testGraphFromEdges(t, 5, []Edge{{2, 4}, {4, 1}, {1, 3}, {3, 2}, {4, 0}}, false)
	leftCanonical, err := left.CanonicalGraph(nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = leftCanonical.Graph.Close() })
	rightCanonical, err := right.CanonicalGraph(nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rightCanonical.Graph.Close() })
	if !reflect.DeepEqual(canonicalEdgeKeys(t, leftCanonical.Graph), canonicalEdgeKeys(t, rightCanonical.Graph)) {
		t.Fatalf("canonical graphs differ: %v vs %v", canonicalEdgeKeys(t, leftCanonical.Graph), canonicalEdgeKeys(t, rightCanonical.Graph))
	}
	assertSourceToCanonicalEdges(t, left, leftCanonical.Graph, leftCanonical.SourceToCanonical)
	if err := left.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := leftCanonical.Graph.VertexCount(); err != nil {
		t.Fatalf("canonical graph after source close: %v", err)
	}
}

func TestColoredCanonicalPermutation(t *testing.T) {
	graph := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, false)
	plain, err := graph.CanonicalPermutation(nil)
	if err != nil {
		t.Fatal(err)
	}
	colored, err := graph.CanonicalPermutation([]int{1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}
	if len(plain) != 3 || len(colored) != 3 {
		t.Fatalf("permutations = %v, %v", plain, colored)
	}
	if !isPermutation(colored) {
		t.Fatalf("colored permutation = %v", colored)
	}
	plain[0] = 99
	if colored[0] == 99 {
		t.Fatal("canonical permutations share storage")
	}
}

func TestAutomorphismGroupSizeAndGenerators(t *testing.T) {
	triangle := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}, {2, 0}}, false)
	size, err := triangle.AutomorphismGroupSize(nil)
	if err != nil || size.Cmp(big.NewInt(6)) != 0 {
		t.Fatalf("triangle group size = %v, %v; want 6", size, err)
	}
	generators, err := triangle.AutomorphismGenerators(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(generators) == 0 {
		t.Fatal("triangle generators are empty")
	}
	for _, generator := range generators {
		if len(generator) != 3 || !isPermutation(generator) {
			t.Fatalf("invalid generator %v", generator)
		}
	}

	coloredSize, err := triangle.AutomorphismGroupSize([]int{1, 2, 3})
	if err != nil || coloredSize.Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("colored group size = %v, %v; want 1", coloredSize, err)
	}
	if err := triangle.Close(); err != nil {
		t.Fatal(err)
	}
	if size.Cmp(big.NewInt(6)) != 0 || len(generators[0]) != 3 {
		t.Fatal("automorphism results did not survive graph closure")
	}
}

func TestCanonicalEmptyLoopsAndAsymmetricGraph(t *testing.T) {
	empty := testGraphFromEdges(t, 0, nil, false)
	permutation, err := empty.CanonicalPermutation([]int{})
	if err != nil || permutation == nil || len(permutation) != 0 {
		t.Fatalf("empty canonical permutation = %#v, %v", permutation, err)
	}
	generators, err := empty.AutomorphismGenerators([]int{})
	if err != nil || generators == nil {
		t.Fatalf("empty generators = %#v, %v", generators, err)
	}
	size, err := empty.AutomorphismGroupSize([]int{})
	if err != nil || size.Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("empty group size = %v, %v", size, err)
	}

	loop := testGraphFromEdges(t, 3, []Edge{{0, 0}, {0, 1}, {1, 2}}, false)
	loopSize, err := loop.AutomorphismGroupSize(nil)
	if err != nil || loopSize.Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("loop graph group size = %v, %v", loopSize, err)
	}
}

func TestCanonicalValidationAndClosedGraph(t *testing.T) {
	graph := testGraphFromEdges(t, 2, []Edge{{0, 1}}, false)
	parallel := testGraphFromEdges(t, 2, []Edge{{0, 1}, {0, 1}}, false)
	for name, call := range map[string]func() error{
		"canonical parallel":  func() error { _, err := parallel.CanonicalPermutation(nil); return err },
		"graph parallel":      func() error { _, err := parallel.CanonicalGraph(nil); return err },
		"generators parallel": func() error { _, err := parallel.AutomorphismGenerators(nil); return err },
		"count parallel":      func() error { _, err := parallel.AutomorphismGroupSize(nil); return err },
		"colors short":        func() error { _, err := graph.CanonicalPermutation([]int{1}); return err },
		"graph colors short":  func() error { _, err := graph.CanonicalGraph([]int{1}); return err },
		"generator colors short": func() error {
			_, err := graph.AutomorphismGenerators([]int{1})
			return err
		},
		"count colors short": func() error { _, err := graph.AutomorphismGroupSize([]int{1}); return err },
	} {
		t.Run(name, func(t *testing.T) {
			if err := call(); err == nil {
				t.Fatal("error = nil")
			}
		})
	}
	closed := testGraphFromEdges(t, 0, nil, false)
	if err := closed.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := closed.AutomorphismGroupSize(nil); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed graph error = %v", err)
	}
	if _, err := (*Graph)(nil).AutomorphismGenerators(nil); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil graph error = %v", err)
	}
}

func TestInvertPermutationValidation(t *testing.T) {
	if _, err := invertPermutation([]int{2, 0}); err == nil {
		t.Fatal("out-of-range permutation error = nil")
	}
	if _, err := invertPermutation([]int{0, 0}); err == nil {
		t.Fatal("duplicate permutation error = nil")
	}
}

func assertSourceToCanonicalEdges(t *testing.T, source, canonical *Graph, permutation []int) {
	t.Helper()
	edges, err := source.Edges()
	if err != nil {
		t.Fatal(err)
	}
	mapped := make([]Edge, len(edges))
	for index, edge := range edges {
		mapped[index] = Edge{From: permutation[edge.From], To: permutation[edge.To]}
	}
	if !reflect.DeepEqual(edgeKeys(mapped), canonicalEdgeKeys(t, canonical)) {
		t.Fatalf("source-to-canonical permutation %v does not map edges", permutation)
	}
}

func canonicalEdgeKeys(t *testing.T, graph *Graph) [][2]int {
	t.Helper()
	edges, err := graph.Edges()
	if err != nil {
		t.Fatal(err)
	}
	return edgeKeys(edges)
}

func edgeKeys(edges []Edge) [][2]int {
	keys := make([][2]int, len(edges))
	for index, edge := range edges {
		from, to := edge.From, edge.To
		if from > to {
			from, to = to, from
		}
		keys[index] = [2]int{from, to}
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i][0] != keys[j][0] {
			return keys[i][0] < keys[j][0]
		}
		return keys[i][1] < keys[j][1]
	})
	return keys
}

func isPermutation(values []int) bool {
	seen := make([]bool, len(values))
	for _, value := range values {
		if value < 0 || value >= len(values) || seen[value] {
			return false
		}
		seen[value] = true
	}
	return true
}

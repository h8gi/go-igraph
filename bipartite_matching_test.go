package igraph_test

import (
	"errors"
	"math"
	"reflect"
	"sync"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func matchingFixture(t *testing.T, directed bool) igraph.BipartiteGraphResult {
	t.Helper()
	result, err := igraph.NewBipartite(
		igraph.BipartitePartition{false, false, false, true, true, true},
		[]igraph.Edge{{0, 3}, {0, 4}, {1, 4}, {1, 5}, {2, 5}}, directed,
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = result.Graph.Close() })
	return result
}

func TestBipartiteMatchingValidationAndMaximality(t *testing.T) {
	fixture := matchingFixture(t, true)
	valid, err := fixture.Graph.IsBipartiteMatching(fixture.Partition, []igraph.MatchedPair{{FalseVertex: 0, TrueVertex: 3}, {FalseVertex: 1, TrueVertex: 4}})
	if err != nil || !valid {
		t.Fatalf("valid matching = %v, %v", valid, err)
	}
	maximal, err := fixture.Graph.IsMaximalBipartiteMatching(fixture.Partition, []igraph.MatchedPair{{FalseVertex: 0, TrueVertex: 3}, {FalseVertex: 1, TrueVertex: 4}})
	if err != nil || maximal {
		t.Fatalf("non-maximal matching = %v, %v", maximal, err)
	}
	maximal, err = fixture.Graph.IsMaximalBipartiteMatching(fixture.Partition, []igraph.MatchedPair{{FalseVertex: 0, TrueVertex: 3}, {FalseVertex: 1, TrueVertex: 4}, {FalseVertex: 2, TrueVertex: 5}})
	if err != nil || !maximal {
		t.Fatalf("maximal matching = %v, %v", maximal, err)
	}
	valid, err = fixture.Graph.IsBipartiteMatching(fixture.Partition, []igraph.MatchedPair{{FalseVertex: 0, TrueVertex: 5}})
	if err != nil || valid {
		t.Fatalf("missing-edge matching = %v, %v", valid, err)
	}
}

func TestBipartiteMatchingMalformedPairs(t *testing.T) {
	fixture := matchingFixture(t, false)
	invalid := [][]igraph.MatchedPair{
		{{FalseVertex: -1, TrueVertex: 3}},
		{{FalseVertex: 0, TrueVertex: 6}},
		{{FalseVertex: 3, TrueVertex: 0}},
		{{FalseVertex: 0, TrueVertex: 3}, {FalseVertex: 0, TrueVertex: 4}},
		{{FalseVertex: 0, TrueVertex: 3}, {FalseVertex: 1, TrueVertex: 3}},
	}
	for _, pairs := range invalid {
		if _, err := fixture.Graph.IsBipartiteMatching(fixture.Partition, pairs); err == nil {
			t.Fatalf("pairs %v error nil", pairs)
		}
	}
}

func TestMaximumBipartiteMatchingUnweightedAndWeighted(t *testing.T) {
	fixture := matchingFixture(t, true)
	unweighted, err := fixture.Graph.MaximumBipartiteMatching(fixture.Partition, igraph.BipartiteMatchingOptions{})
	if err != nil {
		t.Fatal(err)
	}
	want := []igraph.MatchedPair{{FalseVertex: 0, TrueVertex: 3}, {FalseVertex: 1, TrueVertex: 4}, {FalseVertex: 2, TrueVertex: 5}}
	if unweighted.Size != 3 || unweighted.Weight != 3 || !reflect.DeepEqual(unweighted.Pairs, want) {
		t.Fatalf("unweighted = %#v", unweighted)
	}
	valid, err := fixture.Graph.IsBipartiteMatching(fixture.Partition, unweighted.Pairs)
	if err != nil || !valid {
		t.Fatalf("result validity = %v, %v", valid, err)
	}

	// Edge order follows the constructor: choosing 0-4 and 1-5 yields the
	// larger total weight while 2 remains unmatched.
	weighted, err := fixture.Graph.MaximumBipartiteMatching(fixture.Partition, igraph.BipartiteMatchingOptions{Weights: []float64{1, 10, 1, 10, -100}})
	if err != nil {
		t.Fatal(err)
	}
	if weighted.Size != 2 || weighted.Weight != 20 || !reflect.DeepEqual(weighted.Pairs, []igraph.MatchedPair{{FalseVertex: 0, TrueVertex: 4}, {FalseVertex: 1, TrueVertex: 5}}) {
		t.Fatalf("weighted = %#v", weighted)
	}
	_ = fixture.Graph.Close()
	weighted.Pairs[0].FalseVertex = 99
}

func TestMaximumBipartiteMatchingEmptyAndOptions(t *testing.T) {
	empty, err := igraph.NewBipartite(igraph.BipartitePartition{}, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	defer empty.Graph.Close()
	result, err := empty.Graph.MaximumBipartiteMatching(empty.Partition, igraph.BipartiteMatchingOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Size != 0 || result.Weight != 0 || result.Pairs == nil {
		t.Fatalf("empty result = %#v", result)
	}

	fixture := matchingFixture(t, false)
	badEpsilon := []float64{0, -1, math.NaN(), math.Inf(1)}
	for _, epsilon := range badEpsilon {
		if _, err := fixture.Graph.MaximumBipartiteMatching(fixture.Partition, igraph.BipartiteMatchingOptions{Weights: make([]float64, 5), Epsilon: &epsilon}); err == nil {
			t.Fatalf("epsilon %v error nil", epsilon)
		}
	}
	if _, err := fixture.Graph.MaximumBipartiteMatching(fixture.Partition, igraph.BipartiteMatchingOptions{Weights: []float64{1}}); err == nil {
		t.Fatal("short weights error nil")
	}
	weights := make([]float64, 5)
	weights[2] = math.NaN()
	if _, err := fixture.Graph.MaximumBipartiteMatching(fixture.Partition, igraph.BipartiteMatchingOptions{Weights: weights}); err == nil {
		t.Fatal("NaN weights error nil")
	}
}

func TestBipartiteMatchingPartitionClosedAndConcurrent(t *testing.T) {
	fixture := matchingFixture(t, false)
	bad := append(igraph.BipartitePartition{}, fixture.Partition...)
	bad[3] = false
	if _, err := fixture.Graph.MaximumBipartiteMatching(bad, igraph.BipartiteMatchingOptions{}); err == nil {
		t.Fatal("invalid partition error nil")
	}
	var group sync.WaitGroup
	for range 6 {
		group.Add(1)
		go func() {
			defer group.Done()
			result, err := fixture.Graph.MaximumBipartiteMatching(fixture.Partition, igraph.BipartiteMatchingOptions{})
			if err != nil || result.Size != 3 {
				t.Errorf("matching = %#v, %v", result, err)
			}
		}()
	}
	group.Wait()
	_ = fixture.Graph.Close()
	if _, err := fixture.Graph.MaximumBipartiteMatching(fixture.Partition, igraph.BipartiteMatchingOptions{}); !errors.Is(err, igraph.ErrClosed) {
		t.Fatalf("closed error = %v", err)
	}
}

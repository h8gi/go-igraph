package igraph

import (
	"errors"
	"testing"
)

func TestMatchingVectorConversion(t *testing.T) {
	partition := BipartitePartition{false, false, true, true}
	mates, err := matchingVector(partition, []MatchedPair{{FalseVertex: 1, TrueVertex: 2}})
	if err != nil {
		t.Fatal(err)
	}
	pairs, err := matchingPairs(partition, mates)
	if err != nil {
		t.Fatal(err)
	}
	if len(pairs) != 1 || pairs[0] != (MatchedPair{FalseVertex: 1, TrueVertex: 2}) {
		t.Fatalf("pairs = %v", pairs)
	}
	if _, err := matchingPairs(partition, []int{-1}); err == nil {
		t.Fatal("short mates error nil")
	}
	if _, err := matchingPairs(partition, []int{2, -1, -1, -1}); err == nil {
		t.Fatal("asymmetric mates error nil")
	}
}

func TestBipartiteMatchingFailureAdapters(t *testing.T) {
	fixture, _ := NewBipartite(BipartitePartition{false, true}, []Edge{{0, 1}}, false)
	t.Cleanup(func() { _ = fixture.Graph.Close() })
	failure := errors.New("injected failure")

	boolInit := defaultMatchingAdapters()
	boolInit.newBool = func([]bool) (*boolVector, error) { return nil, failure }
	if _, err := fixture.Graph.validateBipartiteMatching(fixture.Partition, nil, false, &boolInit); !errors.Is(err, failure) {
		t.Fatalf("bool init = %v", err)
	}
	intInit := defaultMatchingAdapters()
	intInit.newInt = func([]int) (*intVector, error) { return nil, failure }
	if _, err := fixture.Graph.validateBipartiteMatching(fixture.Partition, nil, false, &intInit); !errors.Is(err, failure) {
		t.Fatalf("int init = %v", err)
	}
	upstream := defaultMatchingAdapters()
	upstream.validate = func(*Graph, *boolVector, *intVector, bool) (bool, int) { return false, 1 }
	if _, err := fixture.Graph.validateBipartiteMatching(fixture.Partition, nil, false, &upstream); err == nil {
		t.Fatal("validation upstream error nil")
	}

	weightInit := defaultMatchingAdapters()
	weightInit.newReal = func([]float64) (*realVector, error) { return nil, failure }
	if _, err := fixture.Graph.maximumBipartiteMatching(fixture.Partition, BipartiteMatchingOptions{Weights: []float64{1}}, &weightInit); !errors.Is(err, failure) {
		t.Fatalf("weight init = %v", err)
	}
	maximumUpstream := defaultMatchingAdapters()
	maximumUpstream.maximum = func(*Graph, *boolVector, *realVector, float64, *intVector) (int64, float64, int) { return 0, 0, 1 }
	if _, err := fixture.Graph.maximumBipartiteMatching(fixture.Partition, BipartiteMatchingOptions{}, &maximumUpstream); err == nil {
		t.Fatal("maximum upstream error nil")
	}
	conversion := defaultMatchingAdapters()
	conversion.convertInt = func(*intVector) ([]int, error) { return nil, failure }
	if _, err := fixture.Graph.maximumBipartiteMatching(fixture.Partition, BipartiteMatchingOptions{}, &conversion); !errors.Is(err, failure) {
		t.Fatalf("conversion = %v", err)
	}
	malformed := defaultMatchingAdapters()
	malformed.convertInt = func(*intVector) ([]int, error) { return []int{1, -1}, nil }
	malformed.maximum = func(*Graph, *boolVector, *realVector, float64, *intVector) (int64, float64, int) { return 1, 1, 0 }
	if _, err := fixture.Graph.maximumBipartiteMatching(fixture.Partition, BipartiteMatchingOptions{}, &malformed); err == nil {
		t.Fatal("malformed result error nil")
	}
}

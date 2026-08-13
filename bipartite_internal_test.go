package igraph

import (
	"errors"
	"testing"
)

func TestBipartiteFailureAdapters(t *testing.T) {
	graph, err := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	failure := errors.New("injected failure")

	initialization := defaultBipartiteAdapters()
	initialization.newBool = func([]bool) (*boolVector, error) { return nil, failure }
	if _, err := graph.bipartite(&initialization); !errors.Is(err, failure) {
		t.Fatalf("initialization error = %v", err)
	}

	upstream := defaultBipartiteAdapters()
	upstream.check = func(*Graph, *boolVector) (bool, int) { return false, 1 }
	if _, err := graph.bipartite(&upstream); err == nil {
		t.Fatal("upstream error = nil")
	}

	conversion := defaultBipartiteAdapters()
	conversion.convertBool = func(*boolVector) ([]bool, error) { return nil, failure }
	if _, err := graph.bipartite(&conversion); !errors.Is(err, failure) {
		t.Fatalf("conversion error = %v", err)
	}

	createInitialization := defaultBipartiteAdapters()
	createInitialization.newBool = func([]bool) (*boolVector, error) { return nil, failure }
	if _, err := newBipartite(BipartitePartition{false, true}, []Edge{{0, 1}}, false, &createInitialization); !errors.Is(err, failure) {
		t.Fatalf("create initialization error = %v", err)
	}

	createUpstream := defaultBipartiteAdapters()
	createUpstream.create = func(*boolVector, *intVector, bool) bipartiteGraphCallResult {
		return bipartiteGraphCallResult{code: 1}
	}
	if _, err := newBipartite(BipartitePartition{false, true}, []Edge{{0, 1}}, false, &createUpstream); err == nil {
		t.Fatal("create upstream error = nil")
	}

	fullConversion := defaultBipartiteAdapters()
	fullConversion.convertBool = func(*boolVector) ([]bool, error) { return nil, failure }
	if _, err := newFullBipartite(1, 1, false, DirectionOut, &fullConversion); !errors.Is(err, failure) {
		t.Fatalf("full conversion error = %v", err)
	}
}

func TestBipartiteTemporaryVectorsClose(t *testing.T) {
	graph, err := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })

	adapters := defaultBipartiteAdapters()
	closed := 0
	adapters.closeBool = func(vector *boolVector) { closed++; vector.close() }
	if _, err := graph.bipartite(&adapters); err != nil {
		t.Fatal(err)
	}
	if closed != 1 {
		t.Fatalf("closed vectors = %d, want 1", closed)
	}
}

package igraph

import (
	"errors"
	"testing"
)

func TestBipartiteProjectionFailureAdapters(t *testing.T) {
	fixture, _ := NewBipartite(BipartitePartition{false, false, true}, []Edge{{0, 2}, {1, 2}}, false)
	t.Cleanup(func() { _ = fixture.Graph.Close() })
	failure := errors.New("injected failure")

	boolInit := defaultProjectionAdapters()
	boolInit.newBool = func([]bool) (*boolVector, error) { return nil, failure }
	if _, err := fixture.Graph.bipartiteProjections(fixture.Partition, &boolInit); !errors.Is(err, failure) {
		t.Fatalf("bool init = %v", err)
	}

	intInit := defaultProjectionAdapters()
	intInit.newInt = func([]int) (*intVector, error) { return nil, failure }
	if _, err := fixture.Graph.bipartiteProjections(fixture.Partition, &intInit); !errors.Is(err, failure) {
		t.Fatalf("int init = %v", err)
	}

	upstream := defaultProjectionAdapters()
	upstream.call = func(*Graph, *boolVector, *intVector, *intVector, bool, bool, int) projectionCallResult {
		return projectionCallResult{code: 1}
	}
	if _, err := fixture.Graph.bipartiteProjections(fixture.Partition, &upstream); err == nil {
		t.Fatal("upstream error nil")
	}

	conversion := defaultProjectionAdapters()
	conversion.convertInt = func(*intVector) ([]int, error) { return nil, failure }
	closed := 0
	conversion.closeGraph = func(graph *Graph) error { closed++; return graph.Close() }
	if _, err := fixture.Graph.bipartiteProjections(fixture.Partition, &conversion); !errors.Is(err, failure) {
		t.Fatalf("conversion error = %v", err)
	}
	if closed != 2 {
		t.Fatalf("closed graphs = %d, want 2", closed)
	}

	single := defaultProjectionAdapters()
	single.convertInt = func(*intVector) ([]int, error) { return nil, failure }
	closed = 0
	single.closeGraph = func(graph *Graph) error { closed++; return graph.Close() }
	if _, err := fixture.Graph.bipartiteProjection(fixture.Partition, BipartiteModeFalse, &single); !errors.Is(err, failure) {
		t.Fatalf("single conversion = %v", err)
	}
	if closed != 1 {
		t.Fatalf("single closed graphs = %d", closed)
	}
}

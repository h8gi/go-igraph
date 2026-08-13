package igraph_test

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestBipartiteDetectionAndSuppliedPartition(t *testing.T) {
	graph, err := igraph.NewGraphFromEdges(6, []igraph.Edge{{0, 3}, {1, 3}, {1, 4}}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })

	result, err := graph.Bipartite()
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsBipartite || len(result.Partition) != 6 {
		t.Fatalf("Bipartite() = %#v", result)
	}
	valid, err := graph.IsBipartitePartition(result.Partition)
	if err != nil || !valid {
		t.Fatalf("computed partition valid = %v, error = %v", valid, err)
	}
	valid, err = graph.IsBipartitePartition(igraph.BipartitePartition{false, false, false, true, true, true})
	if err != nil || !valid {
		t.Fatalf("explicit partition valid = %v, error = %v", valid, err)
	}
	valid, err = graph.IsBipartitePartition(igraph.BipartitePartition{false, false, false, false, true, true})
	if err != nil || valid {
		t.Fatalf("invalid partition valid = %v, error = %v", valid, err)
	}
	if _, err := graph.IsBipartitePartition(igraph.BipartitePartition{false}); err == nil {
		t.Fatal("short partition error = nil")
	}

	cycle, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{0, 1}, {1, 2}, {2, 0}}, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cycle.Close() })
	nonBipartite, err := cycle.Bipartite()
	if err != nil {
		t.Fatal(err)
	}
	if nonBipartite.IsBipartite || nonBipartite.Partition == nil || len(nonBipartite.Partition) != 0 {
		t.Fatalf("odd cycle result = %#v", nonBipartite)
	}
}

func TestBipartiteEmptyDisconnectedAndOwnership(t *testing.T) {
	graph, err := igraph.NewGraphFromEdges(4, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	result, err := graph.Bipartite()
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsBipartite || len(result.Partition) != 4 {
		t.Fatalf("edgeless result = %#v", result)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	result.Partition[0] = !result.Partition[0]
	if _, err := graph.Bipartite(); !errors.Is(err, igraph.ErrClosed) {
		t.Fatalf("closed Bipartite error = %v", err)
	}
	if _, err := graph.IsBipartitePartition(result.Partition); !errors.Is(err, igraph.ErrClosed) {
		t.Fatalf("closed partition error = %v", err)
	}
	var nilGraph *igraph.Graph
	if _, err := nilGraph.Bipartite(); !errors.Is(err, igraph.ErrClosed) {
		t.Fatalf("nil Bipartite error = %v", err)
	}
}

func TestBipartiteConcurrentReads(t *testing.T) {
	graph, err := igraph.NewGraphFromEdges(4, []igraph.Edge{{0, 2}, {1, 3}}, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = graph.Close() })
	partition := igraph.BipartitePartition{false, false, true, true}
	var group sync.WaitGroup
	for range 8 {
		group.Add(1)
		go func() {
			defer group.Done()
			result, err := graph.Bipartite()
			if err != nil || !result.IsBipartite {
				t.Errorf("Bipartite() = %#v, %v", result, err)
			}
			valid, err := graph.IsBipartitePartition(partition)
			if err != nil || !valid {
				t.Errorf("IsBipartitePartition() = %v, %v", valid, err)
			}
		}()
	}
	group.Wait()
}

func TestNewBipartiteConstructionValidationAndOwnership(t *testing.T) {
	partition := igraph.BipartitePartition{false, false, true, true}
	edges := []igraph.Edge{{0, 2}, {1, 2}, {1, 2}, {1, 3}}
	result, err := igraph.NewBipartite(partition, edges, true)
	if err != nil {
		t.Fatal(err)
	}
	if result.Graph == nil || !reflect.DeepEqual(result.Partition, partition) {
		t.Fatalf("result = %#v", result)
	}
	partition[0] = true
	if result.Partition[0] {
		t.Fatal("returned partition aliases input")
	}
	count, err := result.Graph.EdgeCount()
	if err != nil || count != 4 {
		t.Fatalf("edge count = %d, error = %v", count, err)
	}
	if err := result.Graph.Close(); err != nil {
		t.Fatal(err)
	}
	result.Partition[0] = true

	invalid := []struct {
		partition igraph.BipartitePartition
		edges     []igraph.Edge
	}{
		{igraph.BipartitePartition{false, true}, []igraph.Edge{{0, 0}}},
		{igraph.BipartitePartition{false, false, true}, []igraph.Edge{{0, 1}}},
		{igraph.BipartitePartition{false, true}, []igraph.Edge{{0, 2}}},
	}
	for _, test := range invalid {
		if _, err := igraph.NewBipartite(test.partition, test.edges, false); err == nil {
			t.Fatalf("NewBipartite(%v, %v) error = nil", test.partition, test.edges)
		}
	}
}

func TestNewFullBipartiteDirectionsAndEmptyModes(t *testing.T) {
	for _, test := range []struct {
		direction igraph.DirectionMode
		edges     int
		first     igraph.Edge
	}{
		{igraph.DirectionOut, 6, igraph.Edge{From: 0, To: 2}},
		{igraph.DirectionIn, 6, igraph.Edge{From: 2, To: 0}},
		{igraph.DirectionAll, 12, igraph.Edge{From: 0, To: 2}},
	} {
		result, err := igraph.NewFullBipartite(2, 3, true, test.direction)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(result.Partition, igraph.BipartitePartition{false, false, true, true, true}) {
			t.Errorf("partition = %v", result.Partition)
		}
		edges, err := result.Graph.Edges()
		if err != nil {
			t.Fatal(err)
		}
		if len(edges) != test.edges || (len(edges) != 0 && edges[0] != test.first) {
			t.Errorf("direction %d edges = %v", test.direction, edges)
		}
		_ = result.Graph.Close()
	}

	empty, err := igraph.NewFullBipartite(0, 3, false, igraph.DirectionOut)
	if err != nil {
		t.Fatal(err)
	}
	defer empty.Graph.Close()
	if !reflect.DeepEqual(empty.Partition, igraph.BipartitePartition{true, true, true}) {
		t.Fatalf("empty-mode partition = %v", empty.Partition)
	}
	if _, err := igraph.NewFullBipartite(-1, 2, false, igraph.DirectionOut); err == nil {
		t.Fatal("negative mode size error = nil")
	}
	if _, err := igraph.NewFullBipartite(1, 2, false, igraph.DirectionMode(99)); err == nil {
		t.Fatal("invalid direction error = nil")
	}
}

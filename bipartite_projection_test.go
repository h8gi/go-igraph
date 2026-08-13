package igraph_test

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func projectionFixture(t *testing.T) igraph.BipartiteGraphResult {
	t.Helper()
	result, err := igraph.NewBipartite(
		igraph.BipartitePartition{false, false, false, true, true},
		[]igraph.Edge{{0, 3}, {0, 4}, {1, 3}, {1, 4}, {2, 4}}, false,
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = result.Graph.Close() })
	return result
}

func TestBipartiteProjectionSizesAndBothModes(t *testing.T) {
	fixture := projectionFixture(t)
	sizes, err := fixture.Graph.BipartiteProjectionSizes(fixture.Partition)
	if err != nil {
		t.Fatal(err)
	}
	if sizes.False != (igraph.BipartiteProjectionSize{Vertices: 3, Edges: 3}) || sizes.True != (igraph.BipartiteProjectionSize{Vertices: 2, Edges: 1}) {
		t.Fatalf("sizes = %#v", sizes)
	}
	result, err := fixture.Graph.BipartiteProjections(fixture.Partition)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = result.False.Graph.Close(); _ = result.True.Graph.Close() })
	if !reflect.DeepEqual(result.False.SourceVertexIDs, []int{0, 1, 2}) || !reflect.DeepEqual(result.True.SourceVertexIDs, []int{3, 4}) {
		t.Fatalf("source IDs = %v / %v", result.False.SourceVertexIDs, result.True.SourceVertexIDs)
	}
	if !reflect.DeepEqual(result.False.Multiplicities, []int{2, 1, 1}) || !reflect.DeepEqual(result.True.Multiplicities, []int{2}) {
		t.Fatalf("multiplicities = %v / %v", result.False.Multiplicities, result.True.Multiplicities)
	}
	falseEdges, _ := result.False.Graph.Edges()
	trueEdges, _ := result.True.Graph.Edges()
	if len(falseEdges) != 3 || len(trueEdges) != 1 {
		t.Fatalf("edges = %v / %v", falseEdges, trueEdges)
	}

	_ = fixture.Graph.Close()
	_ = result.False.Graph.Close()
	if count, err := result.True.Graph.EdgeCount(); err != nil || count != 1 {
		t.Fatalf("independent true graph = %d, %v", count, err)
	}
	result.False.SourceVertexIDs[0] = 99
	if result.True.SourceVertexIDs[0] != 3 {
		t.Fatal("source ID slices alias")
	}
}

func TestBipartiteProjectionSingleModes(t *testing.T) {
	fixture := projectionFixture(t)
	for _, test := range []struct {
		mode           igraph.BipartiteMode
		ids            []int
		multiplicities []int
		edges          int
	}{
		{igraph.BipartiteModeFalse, []int{0, 1, 2}, []int{2, 1, 1}, 3},
		{igraph.BipartiteModeTrue, []int{3, 4}, []int{2}, 1},
	} {
		result, err := fixture.Graph.BipartiteProjection(fixture.Partition, test.mode)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(result.SourceVertexIDs, test.ids) || !reflect.DeepEqual(result.Multiplicities, test.multiplicities) {
			t.Errorf("mode %v result = %#v", test.mode, result)
		}
		count, err := result.Graph.EdgeCount()
		if err != nil || count != test.edges {
			t.Errorf("mode %v edges = %d, %v", test.mode, count, err)
		}
		_ = result.Graph.Close()
	}
}

func TestBipartiteProjectionEmptyModesAndIsolates(t *testing.T) {
	for _, partition := range []igraph.BipartitePartition{{true, true}, {false, false}, {}} {
		graph, err := igraph.NewBipartite(partition, nil, false)
		if err != nil {
			t.Fatal(err)
		}
		both, err := graph.Graph.BipartiteProjections(graph.Partition)
		if err != nil {
			t.Fatal(err)
		}
		if both.False.SourceVertexIDs == nil || both.True.SourceVertexIDs == nil || both.False.Multiplicities == nil || both.True.Multiplicities == nil {
			t.Fatal("nil projection values")
		}
		for _, result := range []igraph.BipartiteProjectionResult{both.False, both.True} {
			count, err := result.Graph.EdgeCount()
			if err != nil || count != 0 {
				t.Fatalf("empty projection edges = %d, %v", count, err)
			}
			_ = result.Graph.Close()
		}
		for _, mode := range []igraph.BipartiteMode{igraph.BipartiteModeFalse, igraph.BipartiteModeTrue} {
			single, err := graph.Graph.BipartiteProjection(graph.Partition, mode)
			if err != nil {
				t.Fatal(err)
			}
			_ = single.Graph.Close()
		}
		_ = graph.Graph.Close()
	}
}

func TestBipartiteProjectionValidationClosedAndConcurrent(t *testing.T) {
	fixture := projectionFixture(t)
	if _, err := fixture.Graph.BipartiteProjections(igraph.BipartitePartition{false}); err == nil {
		t.Fatal("short partition error nil")
	}
	bad := append(igraph.BipartitePartition{}, fixture.Partition...)
	bad[3] = false
	if _, err := fixture.Graph.BipartiteProjections(bad); err == nil {
		t.Fatal("invalid partition error nil")
	}
	var group sync.WaitGroup
	for range 6 {
		group.Add(1)
		go func() {
			defer group.Done()
			result, err := fixture.Graph.BipartiteProjection(fixture.Partition, igraph.BipartiteModeFalse)
			if err != nil {
				t.Error(err)
				return
			}
			_ = result.Graph.Close()
		}()
	}
	group.Wait()
	_ = fixture.Graph.Close()
	if _, err := fixture.Graph.BipartiteProjectionSizes(fixture.Partition); !errors.Is(err, igraph.ErrClosed) {
		t.Fatalf("closed sizes error = %v", err)
	}
	if _, err := fixture.Graph.BipartiteProjections(fixture.Partition); !errors.Is(err, igraph.ErrClosed) {
		t.Fatalf("closed projections error = %v", err)
	}
}

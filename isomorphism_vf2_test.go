package igraph

import (
	"errors"
	"reflect"
	"sync"
	"testing"
)

func TestIsomorphicVF2MappingsAndColors(t *testing.T) {
	source := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, false)
	target := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, false)

	result, err := source.IsomorphicVF2(target, VF2IsomorphismOptions{
		SourceVertexColors: []int{1, 2, 3},
		TargetVertexColors: []int{3, 2, 1},
		SourceEdgeColors:   []int{4, 5},
		TargetEdgeColors:   []int{5, 4},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Found {
		t.Fatal("IsomorphicVF2() Found = false")
	}
	if want := []int{2, 1, 0}; !reflect.DeepEqual(result.SourceToTarget, want) {
		t.Errorf("SourceToTarget = %v, want %v", result.SourceToTarget, want)
	}
	if want := []int{2, 1, 0}; !reflect.DeepEqual(result.TargetToSource, want) {
		t.Errorf("TargetToSource = %v, want %v", result.TargetToSource, want)
	}

	nonmatch, err := source.IsomorphicVF2(target, VF2IsomorphismOptions{
		SourceVertexColors: []int{1, 2, 3},
		TargetVertexColors: []int{1, 2, 4},
	})
	if err != nil {
		t.Fatal(err)
	}
	if nonmatch.Found || nonmatch.SourceToTarget == nil || nonmatch.TargetToSource == nil || len(nonmatch.SourceToTarget) != 0 || len(nonmatch.TargetToSource) != 0 {
		t.Fatalf("IsomorphicVF2(non-match) = %+v, want false and non-nil empty mappings", nonmatch)
	}
}

func TestContainsSubgraphIsomorphicToVF2Mappings(t *testing.T) {
	target := testGraphFromEdges(t, 4, []Edge{{0, 1}, {1, 2}, {2, 3}}, false)
	pattern := testGraphFromEdges(t, 2, []Edge{{0, 1}}, false)
	result, err := target.ContainsSubgraphIsomorphicToVF2(pattern, VF2SubgraphOptions{
		TargetVertexColors:  []int{1, 2, 3, 4},
		PatternVertexColors: []int{2, 3},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Found || !reflect.DeepEqual(result.PatternToTarget, []int{1, 2}) {
		t.Fatalf("subgraph result = %+v", result)
	}
	if want := []int{RemovedID, 0, 1, RemovedID}; !reflect.DeepEqual(result.TargetToPattern, want) {
		t.Errorf("TargetToPattern = %v, want %v", result.TargetToPattern, want)
	}

	nonmatch, err := target.ContainsSubgraphIsomorphicToVF2(pattern, VF2SubgraphOptions{
		TargetVertexColors:  []int{1, 2, 3, 4},
		PatternVertexColors: []int{8, 9},
	})
	if err != nil {
		t.Fatal(err)
	}
	if nonmatch.Found || nonmatch.PatternToTarget == nil || nonmatch.TargetToPattern == nil {
		t.Fatalf("subgraph non-match = %+v", nonmatch)
	}
}

func TestVF2ValidationAndClosedGraphs(t *testing.T) {
	graph := testGraphFromEdges(t, 2, []Edge{{0, 1}}, false)
	loop := testGraphFromEdges(t, 2, []Edge{{0, 0}}, false)
	closed := testGraphFromEdges(t, 2, []Edge{{0, 1}}, false)
	if err := closed.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := graph.IsomorphicVF2(graph, VF2IsomorphismOptions{SourceVertexColors: []int{1, 2}}); err == nil {
		t.Fatal("one-sided vertex colors error = nil")
	}
	if _, err := graph.IsomorphicVF2(graph, VF2IsomorphismOptions{SourceEdgeColors: []int{1}}); err == nil {
		t.Fatal("one-sided edge colors error = nil")
	}
	if _, err := graph.IsomorphicVF2(graph, VF2IsomorphismOptions{SourceVertexColors: []int{1}, TargetVertexColors: []int{1, 2}}); err == nil {
		t.Fatal("invalid vertex color length error = nil")
	}
	if _, err := graph.IsomorphicVF2(graph, VF2IsomorphismOptions{SourceEdgeColors: []int{}, TargetEdgeColors: []int{1}}); err == nil {
		t.Fatal("invalid edge color length error = nil")
	}
	if _, err := graph.ContainsSubgraphIsomorphicToVF2(graph, VF2SubgraphOptions{TargetVertexColors: []int{1, 2}}); err == nil {
		t.Fatal("one-sided subgraph vertex colors error = nil")
	}
	if _, err := graph.ContainsSubgraphIsomorphicToVF2(graph, VF2SubgraphOptions{PatternEdgeColors: []int{1}}); err == nil {
		t.Fatal("one-sided subgraph edge colors error = nil")
	}
	if _, err := graph.IsomorphicVF2(loop, VF2IsomorphismOptions{}); err == nil {
		t.Fatal("non-simple graph error = nil")
	}
	directed := testGraphFromEdges(t, 2, []Edge{{0, 1}}, true)
	if _, err := graph.IsomorphicVF2(directed, VF2IsomorphismOptions{}); err == nil {
		t.Fatal("directedness mismatch error = nil")
	}
	if _, err := graph.ContainsSubgraphIsomorphicToVF2(directed, VF2SubgraphOptions{}); err == nil {
		t.Fatal("subgraph directedness mismatch error = nil")
	}
	if _, err := graph.IsomorphicVF2(closed, VF2IsomorphismOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed graph error = %v, want %v", err, ErrClosed)
	}
	if _, err := (*Graph)(nil).ContainsSubgraphIsomorphicToVF2(graph, VF2SubgraphOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil target error = %v, want %v", err, ErrClosed)
	}
}

func TestVF2EmptyAndConcurrentReversedOperands(t *testing.T) {
	(*vf2Colors)(nil).close()
	left := testGraphFromEdges(t, 0, nil, false)
	right := testGraphFromEdges(t, 0, nil, false)
	result, err := left.IsomorphicVF2(right, VF2IsomorphismOptions{
		SourceVertexColors: []int{}, TargetVertexColors: []int{},
		SourceEdgeColors: []int{}, TargetEdgeColors: []int{},
	})
	if err != nil || !result.Found || result.SourceToTarget == nil || result.TargetToSource == nil {
		t.Fatalf("empty IsomorphicVF2 = %+v, %v", result, err)
	}

	var wg sync.WaitGroup
	errorsCh := make(chan error, 40)
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); _, err := left.IsomorphicVF2(right, VF2IsomorphismOptions{}); errorsCh <- err }()
		go func() { defer wg.Done(); _, err := right.IsomorphicVF2(left, VF2IsomorphismOptions{}); errorsCh <- err }()
	}
	wg.Wait()
	close(errorsCh)
	for err := range errorsCh {
		if err != nil {
			t.Fatalf("concurrent VF2 error = %v", err)
		}
	}
}

func TestVF2ResultsRemainGoOwned(t *testing.T) {
	left := testGraphFromEdges(t, 2, []Edge{{0, 1}}, true)
	right := testGraphFromEdges(t, 2, []Edge{{0, 1}}, true)
	result, err := left.IsomorphicVF2(right, VF2IsomorphismOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := left.Close(); err != nil {
		t.Fatal(err)
	}
	if err := right.Close(); err != nil {
		t.Fatal(err)
	}
	if !result.Found || !reflect.DeepEqual(result.SourceToTarget, []int{0, 1}) {
		t.Fatalf("result after graph closure = %+v", result)
	}
}

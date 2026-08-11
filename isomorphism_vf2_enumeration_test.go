package igraph

import (
	"errors"
	"reflect"
	"sync"
	"testing"
)

func TestCountIsomorphismsVF2(t *testing.T) {
	left := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}, {2, 0}}, false)
	right := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}, {2, 0}}, false)
	count, err := left.CountIsomorphismsVF2(right, VF2IsomorphismOptions{})
	if err != nil || count != 6 {
		t.Fatalf("CountIsomorphismsVF2() = %d, %v; want 6, nil", count, err)
	}
	count, err = left.CountIsomorphismsVF2(right, VF2IsomorphismOptions{
		SourceVertexColors: []int{1, 2, 2},
		TargetVertexColors: []int{1, 2, 2},
	})
	if err != nil || count != 2 {
		t.Fatalf("colored CountIsomorphismsVF2() = %d, %v; want 2, nil", count, err)
	}
}

func TestCountSubgraphIsomorphismsVF2(t *testing.T) {
	target := testGraphFromEdges(t, 4, []Edge{{0, 1}, {1, 2}, {2, 3}}, false)
	pattern := testGraphFromEdges(t, 2, []Edge{{0, 1}}, false)
	count, err := target.CountSubgraphIsomorphismsVF2(pattern, VF2SubgraphOptions{})
	if err != nil || count != 6 {
		t.Fatalf("CountSubgraphIsomorphismsVF2() = %d, %v; want 6, nil", count, err)
	}

	nonmatch := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}, {2, 0}}, false)
	count, err = target.CountSubgraphIsomorphismsVF2(nonmatch, VF2SubgraphOptions{})
	if err != nil || count != 0 {
		t.Fatalf("CountSubgraphIsomorphismsVF2(non-match) = %d, %v", count, err)
	}
}

func TestEnumerateIsomorphismsVF2BoundAndOwnership(t *testing.T) {
	left := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}, {2, 0}}, false)
	right := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}, {2, 0}}, false)
	result, err := left.EnumerateIsomorphismsVF2(right, VF2IsomorphismEnumerationOptions{MaxMappings: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Mappings) != 2 || !result.Truncated {
		t.Fatalf("enumeration = %+v, want two mappings and truncation", result)
	}
	for index, mapping := range result.Mappings {
		if len(mapping) != 3 {
			t.Fatalf("mapping %d = %v", index, mapping)
		}
	}
	if err := left.Close(); err != nil {
		t.Fatal(err)
	}
	if err := right.Close(); err != nil {
		t.Fatal(err)
	}
	if len(result.Mappings[0]) != 3 {
		t.Fatal("mapping did not survive graph closure")
	}
	result.Mappings[0][0] = 99
	if result.Mappings[1][0] == 99 {
		t.Fatal("mapping slices share storage")
	}
}

func TestEnumerateSubgraphIsomorphismsVF2Directions(t *testing.T) {
	target := testGraphFromEdges(t, 4, []Edge{{0, 1}, {1, 2}, {2, 3}}, false)
	pattern := testGraphFromEdges(t, 2, []Edge{{0, 1}}, false)
	result, err := target.EnumerateSubgraphIsomorphismsVF2(pattern, VF2SubgraphEnumerationOptions{MaxMappings: 6})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Mappings) != 6 || result.Truncated {
		t.Fatalf("subgraph enumeration = %+v, want six complete mappings", result)
	}
	for _, mapping := range result.Mappings {
		if len(mapping) != 2 {
			t.Fatalf("pattern-to-target mapping = %v, want length 2", mapping)
		}
	}
}

func TestVF2EnumerationZeroEmptyAndInvalid(t *testing.T) {
	graph := testGraphFromEdges(t, 2, []Edge{{0, 1}}, false)
	nonmatch := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}, {2, 0}}, false)
	result, err := graph.EnumerateIsomorphismsVF2(nonmatch, VF2IsomorphismEnumerationOptions{MaxMappings: 1})
	if err != nil || result.Mappings == nil || len(result.Mappings) != 0 || result.Truncated {
		t.Fatalf("zero enumeration = %+v, %v", result, err)
	}
	if _, err := graph.EnumerateIsomorphismsVF2(graph, VF2IsomorphismEnumerationOptions{}); err == nil {
		t.Fatal("MaxMappings zero error = nil")
	}
	if _, err := graph.EnumerateSubgraphIsomorphismsVF2(graph, VF2SubgraphEnumerationOptions{MaxMappings: -1}); err == nil {
		t.Fatal("negative MaxMappings error = nil")
	}

	empty1 := testGraphFromEdges(t, 0, nil, false)
	empty2 := testGraphFromEdges(t, 0, nil, false)
	empty, err := empty1.EnumerateIsomorphismsVF2(empty2, VF2IsomorphismEnumerationOptions{MaxMappings: 1})
	if err != nil || empty.Truncated || !reflect.DeepEqual(empty.Mappings, [][]int{{}}) {
		t.Fatalf("empty enumeration = %#v, %v", empty, err)
	}
}

func TestVF2CountAndEnumerationErrorsAndConcurrency(t *testing.T) {
	left := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, false)
	right := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, false)
	directed := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, true)
	closed := testGraphFromEdges(t, 0, nil, false)
	if err := closed.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := left.CountIsomorphismsVF2(directed, VF2IsomorphismOptions{}); err == nil {
		t.Fatal("count directedness mismatch error = nil")
	}
	if _, err := left.CountSubgraphIsomorphismsVF2(closed, VF2SubgraphOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed count error = %v", err)
	}
	if _, err := (*Graph)(nil).EnumerateIsomorphismsVF2(right, VF2IsomorphismEnumerationOptions{MaxMappings: 1}); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil enumeration error = %v", err)
	}

	var wg sync.WaitGroup
	errorsCh := make(chan error, 40)
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, err := left.CountIsomorphismsVF2(right, VF2IsomorphismOptions{})
			errorsCh <- err
		}()
		go func() {
			defer wg.Done()
			_, err := right.EnumerateIsomorphismsVF2(left, VF2IsomorphismEnumerationOptions{MaxMappings: 1})
			errorsCh <- err
		}()
	}
	wg.Wait()
	close(errorsCh)
	for err := range errorsCh {
		if err != nil {
			t.Fatalf("concurrent VF2 error = %v", err)
		}
	}
}

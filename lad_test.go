package igraph

import (
	"errors"
	"reflect"
	"sync"
	"testing"
)

func TestLADInducedAndNonInduced(t *testing.T) {
	target := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}, {2, 0}}, false)
	pattern := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, false)
	nonInduced, err := target.ContainsSubgraphIsomorphicToLAD(pattern, LADOptions{})
	if err != nil || !nonInduced.Found || len(nonInduced.PatternToTarget) != 3 {
		t.Fatalf("non-induced LAD = %+v, %v", nonInduced, err)
	}
	induced, err := target.ContainsSubgraphIsomorphicToLAD(pattern, LADOptions{Induced: true})
	if err != nil || induced.Found || induced.PatternToTarget == nil || induced.TargetToPattern == nil {
		t.Fatalf("induced LAD = %+v, %v; want non-match", induced, err)
	}
}

func TestLADDomainsAndMappingDirections(t *testing.T) {
	target := testGraphFromEdges(t, 4, []Edge{{0, 1}, {1, 2}, {2, 3}}, false)
	pattern := testGraphFromEdges(t, 2, []Edge{{0, 1}}, false)
	result, err := target.ContainsSubgraphIsomorphicToLAD(pattern, LADOptions{
		Domains: [][]int{{1}, {2}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Found || !reflect.DeepEqual(result.PatternToTarget, []int{1, 2}) {
		t.Fatalf("PatternToTarget = %v", result.PatternToTarget)
	}
	if want := []int{RemovedID, 0, 1, RemovedID}; !reflect.DeepEqual(result.TargetToPattern, want) {
		t.Fatalf("TargetToPattern = %v, want %v", result.TargetToPattern, want)
	}

	impossible, err := target.ContainsSubgraphIsomorphicToLAD(pattern, LADOptions{
		Domains: [][]int{{0}, {}},
	})
	if err != nil || impossible.Found || impossible.PatternToTarget == nil || impossible.TargetToPattern == nil {
		t.Fatalf("impossible domain result = %+v, %v", impossible, err)
	}
}

func TestEnumerateSubgraphIsomorphismsLAD(t *testing.T) {
	target := testGraphFromEdges(t, 4, []Edge{{0, 1}, {1, 2}, {2, 3}}, false)
	pattern := testGraphFromEdges(t, 2, []Edge{{0, 1}}, false)
	limited, err := target.EnumerateSubgraphIsomorphismsLAD(pattern, LADEnumerationOptions{MaxMappings: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(limited.Mappings) != 2 || !limited.Truncated {
		t.Fatalf("limited LAD = %+v", limited)
	}
	complete, err := target.EnumerateSubgraphIsomorphismsLAD(pattern, LADEnumerationOptions{MaxMappings: 6})
	if err != nil {
		t.Fatal(err)
	}
	if len(complete.Mappings) != 6 || complete.Truncated {
		t.Fatalf("complete LAD = %+v", complete)
	}
	seen := make(map[[2]int]bool)
	for _, mapping := range complete.Mappings {
		if len(mapping) != 2 {
			t.Fatalf("mapping = %v", mapping)
		}
		key := [2]int{mapping[0], mapping[1]}
		if seen[key] {
			t.Fatalf("duplicate mapping %v", mapping)
		}
		seen[key] = true
	}
}

func TestLADDirectedLoopsEmptyAndOwnership(t *testing.T) {
	directedTarget := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, true)
	directedPattern := testGraphFromEdges(t, 2, []Edge{{0, 1}}, true)
	result, err := directedTarget.ContainsSubgraphIsomorphicToLAD(directedPattern, LADOptions{})
	if err != nil || !result.Found {
		t.Fatalf("directed LAD = %+v, %v", result, err)
	}

	loopTarget := testGraphFromEdges(t, 2, []Edge{{1, 1}}, false)
	loopPattern := testGraphFromEdges(t, 1, []Edge{{0, 0}}, false)
	loop, err := loopTarget.ContainsSubgraphIsomorphicToLAD(loopPattern, LADOptions{})
	if err != nil || !loop.Found || !reflect.DeepEqual(loop.PatternToTarget, []int{1}) {
		t.Fatalf("loop LAD = %+v, %v", loop, err)
	}

	emptyPattern := testGraphFromEdges(t, 0, nil, false)
	empty, err := loopTarget.EnumerateSubgraphIsomorphismsLAD(emptyPattern, LADEnumerationOptions{MaxMappings: 1})
	if err != nil || empty.Truncated || !reflect.DeepEqual(empty.Mappings, [][]int{{}}) {
		t.Fatalf("empty LAD = %#v, %v", empty, err)
	}
	if err := loopTarget.Close(); err != nil {
		t.Fatal(err)
	}
	if err := loopPattern.Close(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(loop.PatternToTarget, []int{1}) {
		t.Fatal("LAD result did not survive graph closure")
	}
}

func TestLADValidation(t *testing.T) {
	target := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, false)
	pattern := testGraphFromEdges(t, 2, []Edge{{0, 1}}, false)
	for name, domains := range map[string][][]int{
		"outer length": {{0}},
		"negative ID":  {{-1}, {1}},
		"large ID":     {{3}, {1}},
		"duplicate":    {{0, 0}, {1}},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := target.ContainsSubgraphIsomorphicToLAD(pattern, LADOptions{Domains: domains}); err == nil {
				t.Fatal("error = nil")
			}
		})
	}
	if _, err := target.EnumerateSubgraphIsomorphismsLAD(pattern, LADEnumerationOptions{}); err == nil {
		t.Fatal("MaxMappings zero error = nil")
	}
	if _, err := target.EnumerateSubgraphIsomorphismsLAD(pattern, LADEnumerationOptions{
		LADOptions: LADOptions{Domains: [][]int{{0}}}, MaxMappings: 1,
	}); err == nil {
		t.Fatal("enumeration malformed domains error = nil")
	}
	impossible, err := target.EnumerateSubgraphIsomorphismsLAD(pattern, LADEnumerationOptions{
		LADOptions: LADOptions{Domains: [][]int{{0}, {}}}, MaxMappings: 1,
	})
	if err != nil || impossible.Mappings == nil || len(impossible.Mappings) != 0 {
		t.Fatalf("impossible enumeration = %+v, %v", impossible, err)
	}
	directed := testGraphFromEdges(t, 2, []Edge{{0, 1}}, true)
	if _, err := target.ContainsSubgraphIsomorphicToLAD(directed, LADOptions{}); err == nil {
		t.Fatal("directedness mismatch error = nil")
	}
	parallel := testGraphFromEdges(t, 2, []Edge{{0, 1}, {0, 1}}, false)
	if _, err := target.ContainsSubgraphIsomorphicToLAD(parallel, LADOptions{}); err == nil {
		t.Fatal("parallel pattern error = nil")
	}
	closed := testGraphFromEdges(t, 0, nil, false)
	if err := closed.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := target.ContainsSubgraphIsomorphicToLAD(closed, LADOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed pattern error = %v", err)
	}
}

func TestSubgraphResultMappingValidation(t *testing.T) {
	if _, err := subgraphResultFromPatternMapping([]int{2}, 2); err == nil {
		t.Fatal("out-of-range mapping error = nil")
	}
	if _, err := subgraphResultFromPatternMapping([]int{0, 0}, 2); err == nil {
		t.Fatal("duplicate mapping error = nil")
	}
}

func TestLADConcurrentReversedOperands(t *testing.T) {
	left := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, false)
	right := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, false)
	var wg sync.WaitGroup
	errorsCh := make(chan error, 40)
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, err := left.ContainsSubgraphIsomorphicToLAD(right, LADOptions{})
			errorsCh <- err
		}()
		go func() {
			defer wg.Done()
			_, err := right.EnumerateSubgraphIsomorphismsLAD(left, LADEnumerationOptions{MaxMappings: 1})
			errorsCh <- err
		}()
	}
	wg.Wait()
	close(errorsCh)
	for err := range errorsCh {
		if err != nil {
			t.Fatalf("concurrent LAD error = %v", err)
		}
	}
}

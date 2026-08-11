package igraph

import (
	"errors"
	"strings"
	"testing"
)

func TestSimpleCyclesInitializationAndCleanupFailures(t *testing.T) {
	g := newCycleTestGraph(t, 0, nil, true)
	defer g.Close()
	failure := errors.New("simulated initialization failure")

	first := defaultSimpleCycleAdapters()
	first.initialize = func() (*intVectorList, error) { return nil, failure }
	if _, err := g.simpleCycles(SimpleCycleOptions{MaxResults: 1}, &first); !errors.Is(err, failure) {
		t.Errorf("first initialization error = %v", err)
	}

	second := defaultSimpleCycleAdapters()
	initializations, closes := 0, 0
	second.initialize = func() (*intVectorList, error) {
		initializations++
		if initializations == 2 {
			return nil, failure
		}
		return newIntVectorList()
	}
	second.close = func(list *intVectorList) {
		closes++
		list.close()
	}
	if _, err := g.simpleCycles(SimpleCycleOptions{MaxResults: 1}, &second); !errors.Is(err, failure) {
		t.Errorf("second initialization error = %v", err)
	}
	if closes != 1 {
		t.Errorf("second initialization failure closed %d lists, want 1", closes)
	}

	upstream := defaultSimpleCycleAdapters()
	closes = 0
	upstream.close = func(list *intVectorList) {
		closes++
		list.close()
	}
	upstream.call = func(*Graph, *intVectorList, *intVectorList, simpleCycleParameters) int { return 1 }
	if _, err := g.simpleCycles(SimpleCycleOptions{MaxResults: 1}, &upstream); err == nil || !strings.Contains(err.Error(), "enumerate simple cycles") {
		t.Errorf("upstream error = %v", err)
	}
	if closes != 2 {
		t.Errorf("upstream failure closed %d lists, want 2", closes)
	}
}

func TestSimpleCyclesConversionFailuresAreAtomic(t *testing.T) {
	g := newCycleTestGraph(t, 3, []Edge{{0, 1}, {1, 2}, {2, 0}}, true)
	defer g.Close()
	failure := errors.New("simulated conversion failure")
	for _, failCall := range []int{1, 2} {
		adapters := defaultSimpleCycleAdapters()
		calls, closes := 0, 0
		adapters.close = func(list *intVectorList) {
			closes++
			list.close()
		}
		adapters.convert = func(list *intVectorList) ([][]int, error) {
			calls++
			if calls == failCall {
				return nil, failure
			}
			return list.slices()
		}
		result, err := g.simpleCycles(SimpleCycleOptions{MaxResults: 10}, &adapters)
		if !errors.Is(err, failure) || result.Cycles != nil {
			t.Errorf("conversion %d = %#v, %v", failCall, result, err)
		}
		if closes != 2 {
			t.Errorf("conversion %d failure closed %d lists, want 2", failCall, closes)
		}
	}
}

func TestSimpleCyclesRejectsMismatchedConvertedLists(t *testing.T) {
	g := newCycleTestGraph(t, 0, nil, true)
	defer g.Close()
	for _, converted := range [][][][]int{
		{{{0}}, {}},
		{{{0}}, {{0, 1}}},
	} {
		adapters := defaultSimpleCycleAdapters()
		calls := 0
		adapters.call = func(*Graph, *intVectorList, *intVectorList, simpleCycleParameters) int { return 0 }
		adapters.convert = func(*intVectorList) ([][]int, error) {
			result := converted[calls]
			calls++
			return result, nil
		}
		if _, err := g.simpleCycles(SimpleCycleOptions{MaxResults: 10}, &adapters); err == nil {
			t.Errorf("mismatched converted lists %#v succeeded", converted)
		}
	}
}

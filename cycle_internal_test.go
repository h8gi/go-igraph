package igraph

import (
	"errors"
	"strings"
	"testing"
)

func TestCycleAnalysisInitializationFailures(t *testing.T) {
	g := newCycleTestGraph(t, 3, []Edge{{0, 1}, {1, 2}}, true)
	defer g.Close()
	failure := errors.New("simulated initialization failure")

	topological := defaultTopologicalSortAdapters()
	topological.initialize = func() (*intVector, error) { return nil, failure }
	if _, err := g.topologicalSort(DirectionOut, &topological); !errors.Is(err, failure) {
		t.Errorf("topologicalSort initialization error = %v", err)
	}

	girth := defaultGirthAdapters()
	girth.initialize = func() (*intVector, error) { return nil, failure }
	if _, err := g.girth(&girth); !errors.Is(err, failure) {
		t.Errorf("girth initialization error = %v", err)
	}

	find := defaultFindCycleAdapters()
	find.initialize = func() (*intVector, error) { return nil, failure }
	if _, err := g.findCycle(DirectionOut, &find); !errors.Is(err, failure) {
		t.Errorf("findCycle first initialization error = %v", err)
	}
}

func TestFindCyclePartialInitializationCleansFirstVector(t *testing.T) {
	g := newCycleTestGraph(t, 3, []Edge{{0, 1}, {1, 2}}, true)
	defer g.Close()
	failure := errors.New("simulated second initialization failure")
	initializations := 0
	closes := 0
	adapters := defaultFindCycleAdapters()
	adapters.initialize = func() (*intVector, error) {
		initializations++
		if initializations == 2 {
			return nil, failure
		}
		return newIntVector(nil)
	}
	adapters.close = func(vector *intVector) {
		closes++
		vector.close()
	}
	if _, err := g.findCycle(DirectionOut, &adapters); !errors.Is(err, failure) {
		t.Fatalf("findCycle second initialization error = %v", err)
	}
	if closes != 1 {
		t.Errorf("partial initialization closed %d vectors, want 1", closes)
	}
}

func TestCycleAnalysisUpstreamErrorAdapters(t *testing.T) {
	g := newCycleTestGraph(t, 3, []Edge{{0, 1}, {1, 2}}, true)
	defer g.Close()

	topological := defaultTopologicalSortAdapters()
	topological.call = func(*Graph, *intVector, DirectionMode) int { return 1 }
	if _, err := g.topologicalSort(DirectionOut, &topological); err == nil || !strings.Contains(err.Error(), "topologically sort graph") {
		t.Errorf("topologicalSort upstream error = %v", err)
	}

	find := defaultFindCycleAdapters()
	find.call = func(*Graph, *intVector, *intVector, DirectionMode) int { return 1 }
	if _, err := g.findCycle(DirectionOut, &find); err == nil || !strings.Contains(err.Error(), "find cycle") {
		t.Errorf("findCycle upstream error = %v", err)
	}

	girth := defaultGirthAdapters()
	girth.call = func(*Graph, *intVector) (float64, int) { return 0, 1 }
	if _, err := g.girth(&girth); err == nil || !strings.Contains(err.Error(), "calculate girth") {
		t.Errorf("girth upstream error = %v", err)
	}
}

func TestCycleAnalysisConversionErrorAdapters(t *testing.T) {
	g := newCycleTestGraph(t, 3, []Edge{{0, 1}, {1, 2}, {2, 0}}, true)
	defer g.Close()
	dag := newCycleTestGraph(t, 3, []Edge{{0, 1}, {1, 2}}, true)
	defer dag.Close()
	failure := errors.New("simulated checked conversion failure")

	topological := defaultTopologicalSortAdapters()
	topological.convert = func(*intVector) ([]int, error) { return nil, failure }
	if _, err := dag.topologicalSort(DirectionOut, &topological); !errors.Is(err, failure) {
		t.Errorf("topologicalSort conversion error = %v", err)
	}

	girth := defaultGirthAdapters()
	girth.convert = func(*intVector) ([]int, error) { return nil, failure }
	if _, err := g.girth(&girth); !errors.Is(err, failure) {
		t.Errorf("girth conversion error = %v", err)
	}

	for _, failCall := range []int{1, 2} {
		adapters := defaultFindCycleAdapters()
		calls := 0
		adapters.convert = func(vector *intVector) ([]int, error) {
			calls++
			if calls == failCall {
				return nil, failure
			}
			return vector.slice()
		}
		if _, err := g.findCycle(DirectionOut, &adapters); !errors.Is(err, failure) {
			t.Errorf("findCycle conversion %d error = %v", failCall, err)
		}
	}
}

func TestCyclePredicateInvalidKind(t *testing.T) {
	g := newCycleTestGraph(t, 0, nil, true)
	defer g.Close()
	if _, err := g.cyclePredicate("invalid", cyclePredicateKind(99)); err == nil {
		t.Error("cyclePredicate(invalid) succeeded")
	}
}

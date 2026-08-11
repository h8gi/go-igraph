package igraph

import (
	"errors"
	"testing"
)

func TestCycleAnalysisFailureSeams(t *testing.T) {
	g := newCycleTestGraph(t, 3, []Edge{{0, 1}, {1, 2}, {2, 0}}, true)
	defer g.Close()
	dag := newCycleTestGraph(t, 3, []Edge{{0, 1}, {1, 2}}, true)
	defer dag.Close()
	failure := errors.New("simulated failure")
	failAt := func(target cycleAnalysisStage) cycleAnalysisFailureHook {
		return func(stage cycleAnalysisStage) error {
			if stage == target {
				return failure
			}
			return nil
		}
	}

	for _, stage := range []cycleAnalysisStage{
		cycleAnalysisAfterFirstVectorInit,
		cycleAnalysisBeforeUpstream,
		cycleAnalysisBeforeFirstConversion,
	} {
		if _, err := dag.topologicalSort(DirectionOut, failAt(stage)); !errors.Is(err, failure) {
			t.Errorf("topologicalSort stage %d error = %v", stage, err)
		}
		if _, err := g.girth(failAt(stage)); !errors.Is(err, failure) {
			t.Errorf("girth stage %d error = %v", stage, err)
		}
	}
	for _, stage := range []cycleAnalysisStage{
		cycleAnalysisAfterFirstVectorInit,
		cycleAnalysisAfterSecondVectorInit,
		cycleAnalysisBeforeUpstream,
		cycleAnalysisBeforeFirstConversion,
		cycleAnalysisBeforeSecondConversion,
	} {
		if _, err := g.findCycle(DirectionOut, failAt(stage)); !errors.Is(err, failure) {
			t.Errorf("findCycle stage %d error = %v", stage, err)
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

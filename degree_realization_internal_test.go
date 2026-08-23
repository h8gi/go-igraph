package igraph

import (
	"errors"
	"strings"
	"testing"
)

func TestDegreeSequenceRealizationAdapterFailures(t *testing.T) {
	failure := errors.New("simulated vector initialization failure")

	first := defaultDegreeRealizationAdapters()
	first.newInt = func([]int) (*intVector, error) { return nil, failure }
	if _, err := realizeDegreeSequence([]int{0}, nil, DegreeSequenceRealizationOptions{}, &first); !errors.Is(err, failure) {
		t.Errorf("first initialization error = %v", err)
	}

	second := defaultDegreeRealizationAdapters()
	calls := 0
	second.newInt = func(values []int) (*intVector, error) {
		calls++
		if calls == 2 {
			return nil, failure
		}
		return newIntVector(values)
	}
	if _, err := realizeDegreeSequence([]int{1, 0}, []int{0, 1}, DegreeSequenceRealizationOptions{}, &second); !errors.Is(err, failure) {
		t.Errorf("second initialization error = %v", err)
	}

	bipartite := defaultDegreeRealizationAdapters()
	calls = 0
	bipartite.newInt = second.newInt
	if _, err := realizeBipartiteDegreeSequence([]int{1}, []int{1}, DegreeSequenceRealizationOptions{}, &bipartite); !errors.Is(err, failure) {
		t.Errorf("bipartite second initialization error = %v", err)
	}

	upstream := defaultDegreeRealizationAdapters()
	upstream.realize = func(*intVector, *intVector, EdgeType, DegreeSequenceRealizationMethod) degreeRealizationCallResult {
		return degreeRealizationCallResult{code: 1}
	}
	if _, err := realizeDegreeSequence([]int{0}, nil, DegreeSequenceRealizationOptions{}, &upstream); err == nil || !strings.Contains(err.Error(), "realize degree sequence") {
		t.Errorf("ordinary upstream error = %v", err)
	}

	upstream.bipartite = func(*intVector, *intVector, EdgeType, DegreeSequenceRealizationMethod) degreeRealizationCallResult {
		return degreeRealizationCallResult{code: 1}
	}
	if _, err := realizeBipartiteDegreeSequence(nil, nil, DegreeSequenceRealizationOptions{}, &upstream); err == nil || !strings.Contains(err.Error(), "realize bipartite degree sequence") {
		t.Errorf("bipartite upstream error = %v", err)
	}
}

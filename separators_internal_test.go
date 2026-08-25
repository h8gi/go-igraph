package igraph

import (
	"errors"
	"testing"
)

func TestEvaluateSeparatorPredicateFailuresAndCleanup(t *testing.T) {
	forced := errors.New("forced failure")
	tests := []struct {
		name            string
		materialized    []int
		failMaterialize bool
		failSelector    bool
		queryCode       int
		wantClosed      int
		wantQueried     bool
	}{
		{name: "materialization", failMaterialize: true},
		{name: "duplicate", materialized: []int{1, 1}},
		{name: "selector initialization", materialized: []int{1}, failSelector: true},
		{name: "upstream", materialized: []int{1}, queryCode: 4, wantClosed: 1, wantQueried: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			closed := 0
			queried := false
			_, err := evaluateSeparatorPredicate(NoVertices(), "test separator", separatorPredicateAdapters{
				materialize: func(VertexSelector) ([]int, error) {
					if tt.failMaterialize {
						return nil, forced
					}
					return tt.materialized, nil
				},
				newSelector: func(VertexSelector) (*cVertexSelector, error) {
					if tt.failSelector {
						return nil, forced
					}
					return &cVertexSelector{}, nil
				},
				close: func(*cVertexSelector) { closed++ },
				query: func(*cVertexSelector) (bool, int) {
					queried = true
					return false, tt.queryCode
				},
			})
			if err == nil {
				t.Fatal("evaluateSeparatorPredicate() error = nil")
			}
			if tt.failMaterialize || tt.failSelector {
				if !errors.Is(err, forced) {
					t.Errorf("error = %v, want %v", err, forced)
				}
			}
			if closed != tt.wantClosed || queried != tt.wantQueried {
				t.Errorf("closed=%d queried=%t, want %d %t", closed, queried, tt.wantClosed, tt.wantQueried)
			}
		})
	}
}

func TestEvaluateSeparatorPredicateSuccess(t *testing.T) {
	closed := 0
	got, err := evaluateSeparatorPredicate(NoVertices(), "test separator", separatorPredicateAdapters{
		materialize: func(VertexSelector) ([]int, error) { return []int{1}, nil },
		newSelector: func(VertexSelector) (*cVertexSelector, error) { return &cVertexSelector{}, nil },
		close:       func(*cVertexSelector) { closed++ },
		query:       func(*cVertexSelector) (bool, int) { return true, 0 },
	})
	if err != nil || !got || closed != 1 {
		t.Fatalf("evaluateSeparatorPredicate() = %t, %v, closed=%d", got, err, closed)
	}
}

package igraph

import (
	"errors"
	"reflect"
	"testing"
)

func TestCollectPercolationFailuresAndCleanup(t *testing.T) {
	forced := errors.New("forced failure")
	tests := []struct {
		name          string
		failInitCall  int
		failQuery     bool
		failSliceCall int
		wantClosed    int
	}{
		{name: "input initialization", failInitCall: 1},
		{name: "giant initialization", failInitCall: 2, wantClosed: 1},
		{name: "active initialization", failInitCall: 3, wantClosed: 2},
		{name: "upstream", failQuery: true, wantClosed: 3},
		{name: "giant conversion", failSliceCall: 1, wantClosed: 3},
		{name: "active conversion", failSliceCall: 2, wantClosed: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initCalls := 0
			sliceCalls := 0
			closed := 0
			_, err := collectPercolation([]int{0}, 1, "test", percolationOperations{
				newVector: func([]int) (*intVector, error) {
					initCalls++
					if initCalls == tt.failInitCall {
						return nil, forced
					}
					return &intVector{}, nil
				},
				close: func(*intVector) { closed++ },
				query: func(_, _, _ *intVector) error {
					if tt.failQuery {
						return forced
					}
					return nil
				},
				slice: func(*intVector) ([]int, error) {
					sliceCalls++
					if sliceCalls == tt.failSliceCall {
						return nil, forced
					}
					return []int{0}, nil
				},
			})
			if !errors.Is(err, forced) {
				t.Errorf("error = %v, want %v", err, forced)
			}
			if closed != tt.wantClosed {
				t.Errorf("close count = %d, want %d", closed, tt.wantClosed)
			}
		})
	}
}

func TestValidatePercolationSeries(t *testing.T) {
	valid := percolationSeries{giant: []int{1, 2}, active: []int{0, 1}}
	if err := validatePercolationSeries(valid, 2, "test"); err != nil {
		t.Fatal(err)
	}
	invalid := []percolationSeries{
		{},
		{giant: []int{}, active: nil},
		{giant: []int{1}, active: []int{0, 1}},
		{giant: []int{-1}, active: []int{0}},
		{giant: []int{1}, active: []int{-1}},
		{giant: []int{2, 1}, active: []int{0, 1}},
		{giant: []int{1, 2}, active: []int{1, 0}},
	}
	for _, series := range invalid {
		if err := validatePercolationSeries(series, len(series.giant), "test"); err == nil {
			t.Errorf("validatePercolationSeries(%#v) error = nil", series)
		}
	}
}

func TestPercolationInputHelpers(t *testing.T) {
	if err := validatePercolationOrder([]int{1, 0}, 2, "edge"); err != nil {
		t.Fatal(err)
	}
	for _, order := range [][]int{{0}, {0, 0}, {-1, 1}, {0, 2}} {
		if err := validatePercolationOrder(order, 2, "edge"); err == nil {
			t.Errorf("validatePercolationOrder(%v) error = nil", order)
		}
	}
	endpoints, err := percolationEndpoints([]Edge{{From: 2, To: 1}, {From: 1, To: 1}})
	if err != nil || !reflect.DeepEqual(endpoints, []int{2, 1, 1, 1}) {
		t.Errorf("percolationEndpoints() = %v, %v", endpoints, err)
	}
}

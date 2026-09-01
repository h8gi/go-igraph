package igraph

import (
	"errors"
	"testing"
)

func TestRunningMeanFailureAdapters(t *testing.T) {
	forced := errors.New("forced running-mean failure")
	for name, mutate := range map[string]func(*statisticalUtilityAdapters){
		"input initialization": func(a *statisticalUtilityAdapters) {
			a.newReal = func([]float64) (*realVector, error) { return nil, forced }
		},
		"upstream": func(a *statisticalUtilityAdapters) {
			a.runningMean = func(*realVector, int, *realVector) int { return 4 }
		},
		"conversion": func(a *statisticalUtilityAdapters) {
			a.readReal = func(*realVector) ([]float64, error) { return nil, forced }
		},
		"malformed output": func(a *statisticalUtilityAdapters) {
			a.readReal = func(*realVector) ([]float64, error) { return []float64{}, nil }
		},
	} {
		t.Run(name, func(t *testing.T) {
			adapters := defaultStatisticalUtilityAdapters()
			mutate(&adapters)
			if _, err := runningMean([]float64{1, 2, 3}, 2, &adapters); err == nil {
				t.Fatal("failure not propagated")
			}
		})
	}

	secondInitialization := defaultStatisticalUtilityAdapters()
	calls := 0
	secondInitialization.newReal = func(values []float64) (*realVector, error) {
		calls++
		if calls == 2 {
			return nil, forced
		}
		return newRealVector(values)
	}
	if _, err := runningMean([]float64{1, 2}, 1, &secondInitialization); !errors.Is(err, forced) {
		t.Fatalf("result initialization error = %v", err)
	}
}

func TestSampleIntegersFailureAdapters(t *testing.T) {
	forced := errors.New("forced integer-sample failure")
	for name, mutate := range map[string]func(*statisticalUtilityAdapters){
		"initialization": func(a *statisticalUtilityAdapters) { a.newInt = func([]int) (*intVector, error) { return nil, forced } },
		"upstream": func(a *statisticalUtilityAdapters) {
			a.sample = func(int, int, int, *uint64, *intVector) int { return 4 }
		},
		"conversion": func(a *statisticalUtilityAdapters) {
			a.readInt = func(*intVector) ([]int, error) { return nil, forced }
		},
		"malformed output": func(a *statisticalUtilityAdapters) {
			a.readInt = func(*intVector) ([]int, error) { return []int{2, 1}, nil }
		},
	} {
		t.Run(name, func(t *testing.T) {
			adapters := defaultStatisticalUtilityAdapters()
			mutate(&adapters)
			if _, err := sampleIntegers(0, 10, 2, IntegerSampleOptions{}, &adapters); err == nil {
				t.Fatal("failure not propagated")
			}
		})
	}
}

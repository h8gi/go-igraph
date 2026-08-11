package igraph

import (
	"errors"
	"strings"
	"testing"
)

func TestCycleBasisFailureAdapters(t *testing.T) {
	g := newCycleTestGraph(t, 3, []Edge{{0, 1}, {1, 2}, {2, 0}}, false)
	defer g.Close()
	failure := errors.New("simulated cycle-basis failure")

	for _, minimum := range []bool{false, true} {
		initialization := defaultCycleBasisAdapters()
		initialization.initialize = func() (*intVectorList, error) { return nil, failure }
		if minimum {
			if _, err := g.minimumCycleBasis(MinimumCycleBasisOptions{}, &initialization); !errors.Is(err, failure) {
				t.Errorf("minimum initialization error = %v", err)
			}
		} else if _, err := g.fundamentalCycleBasis(FundamentalCycleBasisOptions{}, &initialization); !errors.Is(err, failure) {
			t.Errorf("fundamental initialization error = %v", err)
		}

		upstream := defaultCycleBasisAdapters()
		closes := 0
		upstream.close = func(list *intVectorList) {
			closes++
			list.close()
		}
		if minimum {
			upstream.minimumCall = func(*Graph, *intVectorList, float64, bool, bool) int { return 1 }
			if _, err := g.minimumCycleBasis(MinimumCycleBasisOptions{}, &upstream); err == nil || !strings.Contains(err.Error(), "calculate minimum cycle basis") {
				t.Errorf("minimum upstream error = %v", err)
			}
		} else {
			upstream.fundamentalCall = func(*Graph, *intVectorList, int, float64) int { return 1 }
			if _, err := g.fundamentalCycleBasis(FundamentalCycleBasisOptions{}, &upstream); err == nil || !strings.Contains(err.Error(), "calculate fundamental cycle basis") {
				t.Errorf("fundamental upstream error = %v", err)
			}
		}
		if closes != 1 {
			t.Errorf("upstream failure closed %d lists, want 1", closes)
		}

		conversion := defaultCycleBasisAdapters()
		conversion.convert = func(*intVectorList) ([][]int, error) { return nil, failure }
		if minimum {
			if _, err := g.minimumCycleBasis(MinimumCycleBasisOptions{}, &conversion); !errors.Is(err, failure) {
				t.Errorf("minimum conversion error = %v", err)
			}
		} else if _, err := g.fundamentalCycleBasis(FundamentalCycleBasisOptions{}, &conversion); !errors.Is(err, failure) {
			t.Errorf("fundamental conversion error = %v", err)
		}
	}
}

func TestCycleBasisPartialNestedConversion(t *testing.T) {
	g := newCycleTestGraph(t, 4, []Edge{{0, 1}, {1, 2}, {2, 0}, {1, 3}, {3, 2}}, false)
	defer g.Close()
	failure := errors.New("simulated second nested conversion failure")
	adapters := defaultCycleBasisAdapters()
	adapters.convert = func(list *intVectorList) ([][]int, error) {
		return list.slicesWithHooks(intVectorListSliceHooks{beforeConvert: func(index int) error {
			if index == 1 {
				return failure
			}
			return nil
		}})
	}
	result, err := g.minimumCycleBasis(MinimumCycleBasisOptions{}, &adapters)
	if result != nil || !errors.Is(err, failure) {
		t.Errorf("partial nested conversion = %#v, %v", result, err)
	}
}

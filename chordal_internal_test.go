package igraph

import (
	"errors"
	"testing"
)

func TestOrderingVectorsInitializationFailures(t *testing.T) {
	injected := errors.New("injected vector initialization failure")
	if alpha, inverse, err := orderingVectorsWithInitializer([]int{0}, 1, func([]int) (*intVector, error) {
		return nil, injected
	}); !errors.Is(err, injected) || alpha != nil || inverse != nil {
		t.Fatalf("first initialization = %v, %v, %v", alpha, inverse, err)
	}

	calls := 0
	initializer := func(values []int) (*intVector, error) {
		calls++
		if calls == 2 {
			return nil, injected
		}
		return newIntVector(values)
	}
	if alpha, inverse, err := orderingVectorsWithInitializer([]int{0}, 1, initializer); !errors.Is(err, injected) || alpha != nil || inverse != nil {
		t.Fatalf("second initialization = %v, %v, %v", alpha, inverse, err)
	}
}

func TestMaximumCardinalityOrderConversionValidation(t *testing.T) {
	valid, err := maximumCardinalityOrderFromSlices([]int{1, 0}, []int{1, 0})
	if err != nil || len(valid.Vertices) != 2 {
		t.Fatalf("valid conversion = %#v, %v", valid, err)
	}
	for _, test := range []struct {
		ranks  []int
		byRank []int
	}{
		{[]int{0}, nil},
		{[]int{1, 0}, []int{1, 2}},
		{[]int{1, 0}, []int{1, 1}},
		{[]int{0, 1}, []int{1, 0}},
	} {
		if _, err := maximumCardinalityOrderFromSlices(test.ranks, test.byRank); err == nil {
			t.Fatalf("malformed order accepted: ranks=%v inverse=%v", test.ranks, test.byRank)
		}
	}
}

func TestPerfectGraphUpstreamErrors(t *testing.T) {
	g, err := NewGraphFromEdges(1, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	if _, err := g.isPerfect(&perfectGraphAdapters{
		isSimple:  func(*Graph) (bool, int) { return false, 4 },
		isPerfect: func(*Graph) (bool, int) { t.Fatal("perfect callback called"); return false, 0 },
	}); err == nil {
		t.Fatal("simple-check error not propagated")
	}
	if _, err := g.isPerfect(&perfectGraphAdapters{
		isSimple:  func(*Graph) (bool, int) { return true, 0 },
		isPerfect: func(*Graph) (bool, int) { return false, 4 },
	}); err == nil {
		t.Fatal("perfect-check error not propagated")
	}
}

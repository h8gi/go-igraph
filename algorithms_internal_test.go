package igraph

import (
	"errors"
	"testing"
)

func TestAlgorithmsInvalidInputs(t *testing.T) {
	g, err := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	if _, err := g.Degree(VertexSelector{kind: vertexSelectorKind(99)}, DegreeOptions{}); err == nil {
		t.Error("expected error for invalid vertex selector in Degree")
	}

	if _, err := g.Neighborhoods(AllVertices(), NeighborhoodOptions{Direction: DirectionMode(99)}); err == nil {
		t.Error("expected error for invalid DirectionMode in Neighborhoods")
	}

	if _, err := g.Neighborhoods(AllVertices(), NeighborhoodOptions{Order: -1}); err == nil {
		t.Error("expected error for negative order in Neighborhoods")
	}

	if _, err := g.Neighborhoods(AllVertices(), NeighborhoodOptions{MinDistance: -1}); err == nil {
		t.Error("expected error for negative MinDistance in Neighborhoods")
	}

	if _, err := g.Neighborhoods(AllVertices(), NeighborhoodOptions{Order: 1, MinDistance: 2}); err == nil {
		t.Error("expected error for MinDistance > Order in Neighborhoods")
	}

	if _, err := g.NeighborhoodSizes(AllVertices(), NeighborhoodOptions{Order: -1}); err == nil {
		t.Error("expected error for negative order in NeighborhoodSizes")
	}
}

func TestIntVectorListInitializationAndPartialConversionFailures(t *testing.T) {
	injected := errors.New("injected vector-list failure")
	list, err := initializeIntVectorList(&intVectorList{}, func() int { return 1 })
	if list != nil || err == nil {
		t.Fatalf("initializeIntVectorList(failure) = %v, %v", list, err)
	}
	if list, err = initializeIntVectorList(nil, func() int { return 0 }); list != nil || err == nil {
		t.Fatalf("initializeIntVectorList(nil) = %v, %v", list, err)
	}

	list, err = newLADDomainList([][]int{{0}, {1}})
	if err != nil {
		t.Fatal(err)
	}
	defer list.close()
	slices, err := list.slicesWithHooks(intVectorListSliceHooks{
		beforeConvert: func(index int) error {
			if index == 1 {
				return injected
			}
			return nil
		},
	})
	if slices != nil || !errors.Is(err, injected) {
		t.Fatalf("partial slices = %v, %v; want nil, injected error", slices, err)
	}
}

func TestIntVectorListEmptySlices(t *testing.T) {
	list, err := newIntVectorList()
	if err != nil {
		t.Fatal(err)
	}
	defer list.close()

	slices, err := list.slices()
	if err != nil {
		t.Fatalf("unexpected error converting empty intVectorList to slices: %v", err)
	}
	if len(slices) != 0 {
		t.Errorf("expected 0 slices, got %d", len(slices))
	}
}

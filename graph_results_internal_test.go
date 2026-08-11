package igraph

import (
	"testing"
)

func TestValidateCapacitiesOrUnitAndFlowsOrZero(t *testing.T) {
	g, err := NewGraphFromEdges(3, []Edge{{From: 0, To: 1}, {From: 1, To: 2}}, true)
	if err != nil {
		t.Fatalf("NewGraphFromEdges failed: %v", err)
	}
	defer g.Close()

	// Test validateCapacitiesOrUnit with nil
	cVec, err := validateCapacitiesOrUnit(g, nil)
	if err != nil {
		t.Fatalf("validateCapacitiesOrUnit(nil) error: %v", err)
	}
	defer cVec.close()

	cSlice, err := cVec.slice()
	if err != nil {
		t.Fatalf("cVec.slice() error: %v", err)
	}
	if len(cSlice) != 2 || cSlice[0] != 1.0 || cSlice[1] != 1.0 {
		t.Errorf("unexpected unit capacities: %v", cSlice)
	}

	// Test validateFlowsOrZero with nil
	fVec, err := validateFlowsOrZero(g, nil)
	if err != nil {
		t.Fatalf("validateFlowsOrZero(nil) error: %v", err)
	}
	defer fVec.close()

	fSlice, err := fVec.slice()
	if err != nil {
		t.Fatalf("fVec.slice() error: %v", err)
	}
	if len(fSlice) != 2 || fSlice[0] != 0.0 || fSlice[1] != 0.0 {
		t.Errorf("unexpected zero flows: %v", fSlice)
	}
}

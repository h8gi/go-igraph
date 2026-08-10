package igraph

import (
	"math"
	"testing"
)

func TestValidateCapacitiesInternal(t *testing.T) {
	g, err := NewGraphFromEdges(3, []Edge{{From: 0, To: 1}, {From: 1, To: 2}}, true)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}
	defer g.Close()

	v, err := validateCapacities(g, nil)
	if err != nil || v != nil {
		t.Errorf("expected nil vector and no error for nil capacities")
	}

	v, err = validateCapacities(g, []float64{1.0, 2.0})
	if err != nil {
		t.Fatalf("unexpected error for valid capacities: %v", err)
	}
	if v == nil {
		t.Fatalf("expected non-nil vector for valid capacities")
	}
	v.close()

	// Invalid length
	if _, err := validateCapacities(g, []float64{1.0}); err == nil {
		t.Errorf("expected error for invalid capacity length")
	}

	// Negative capacity
	if _, err := validateCapacities(g, []float64{-1.0, 2.0}); err == nil {
		t.Errorf("expected error for negative capacity")
	}

	// NaN capacity
	if _, err := validateCapacities(g, []float64{1.0, math.NaN()}); err == nil {
		t.Errorf("expected error for NaN capacity")
	}
}

func TestValidateSourceTargetInternal(t *testing.T) {
	g, err := NewGraphFromEdges(3, []Edge{{From: 0, To: 1}, {From: 1, To: 2}}, true)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}
	defer g.Close()

	src, tgt, err := validateSourceTarget(g, 0, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if int(src) != 0 || int(tgt) != 2 {
		t.Errorf("expected (0, 2), got (%d, %d)", src, tgt)
	}

	if _, _, err := validateSourceTarget(g, -1, 2); err == nil {
		t.Errorf("expected error for negative source")
	}

	if _, _, err := validateSourceTarget(g, 0, 10); err == nil {
		t.Errorf("expected error for out of bounds target")
	}

	if _, _, err := validateSourceTarget(g, 1, 1); err == nil {
		t.Errorf("expected error for source == target")
	}
}

package igraph

import (
	"math"
	"testing"
)

func TestCalculateCentralizationValidationErrors(t *testing.T) {
	if _, err := CalculateCentralization([]float64{math.NaN()}, 1.0, false); err == nil {
		t.Error("expected error for NaN score in CalculateCentralization")
	}

	if _, err := CalculateCentralization([]float64{1.0}, -1.0, false); err == nil {
		t.Error("expected error for negative theoreticalMaximum")
	}

	if _, err := CalculateCentralization([]float64{1.0}, 0.0, true); err == nil {
		t.Error("expected error for zero theoreticalMaximum when normalized")
	}
}

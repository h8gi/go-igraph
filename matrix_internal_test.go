package igraph

import (
	"testing"
)

func TestNewCMatrixCorruptedLengthError(t *testing.T) {
	invalidSize := Matrix{rows: -1, columns: 2, values: []float64{}}
	if _, err := newCMatrix(invalidSize); err == nil {
		t.Error("expected error for invalid matrix size")
	}

	invalidValues := Matrix{rows: 2, columns: 2, values: []float64{1.0}}
	if _, err := newCMatrix(invalidValues); err == nil {
		t.Error("expected error for corrupted matrix values length")
	}
}

func TestNewMatrixValidationErrors(t *testing.T) {
	if _, err := NewMatrix(-1, 2); err == nil {
		t.Error("expected error for negative row size in NewMatrix")
	}

	ragged := [][]float64{{1.0, 2.0}, {3.0}}
	if _, err := NewMatrixFromRows(ragged); err == nil {
		t.Error("expected error for ragged rows in NewMatrixFromRows")
	}
}

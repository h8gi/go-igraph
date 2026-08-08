package igraph

import (
	"reflect"
	"strings"
	"testing"
)

func TestMatrixFromRows(t *testing.T) {
	matrix, err := NewMatrixFromRows([][]float64{{1, 2, 3}, {4, 5, 6}})
	if err != nil {
		t.Fatal(err)
	}
	if rows, columns := matrix.Dims(); rows != 2 || columns != 3 {
		t.Errorf("Dims() = (%d, %d), want (2, 3)", rows, columns)
	}
	if got, err := matrix.At(1, 2); err != nil || got != 6 {
		t.Errorf("At(1, 2) = %v, %v, want 6, nil", got, err)
	}
	if got := matrix.Rows(); !reflect.DeepEqual(got, [][]float64{{1, 2, 3}, {4, 5, 6}}) {
		t.Errorf("Rows() = %#v", got)
	}
}

func TestMatrixZeroDimensions(t *testing.T) {
	tests := []struct {
		rows    int
		columns int
	}{
		{rows: 0, columns: 0},
		{rows: 0, columns: 3},
		{rows: 3, columns: 0},
	}
	for _, tt := range tests {
		matrix, err := NewMatrix(tt.rows, tt.columns)
		if err != nil {
			t.Fatalf("NewMatrix(%d, %d) error = %v", tt.rows, tt.columns, err)
		}
		if rows, columns := matrix.Dims(); rows != tt.rows || columns != tt.columns {
			t.Errorf("Dims() = (%d, %d), want (%d, %d)", rows, columns, tt.rows, tt.columns)
		}
		gotRows := matrix.Rows()
		if gotRows == nil || len(gotRows) != tt.rows {
			t.Errorf("Rows() = %#v, want non-nil with %d rows", gotRows, tt.rows)
		}
		for row, values := range gotRows {
			if values == nil || len(values) != tt.columns {
				t.Errorf("Rows()[%d] = %#v, want non-nil with %d columns", row, values, tt.columns)
			}
		}
	}

	fromNil, err := NewMatrixFromRows(nil)
	if err != nil {
		t.Fatal(err)
	}
	if rows, columns := fromNil.Dims(); rows != 0 || columns != 0 {
		t.Errorf("NewMatrixFromRows(nil) dimensions = (%d, %d)", rows, columns)
	}
}

func TestMatrixRejectsInvalidDimensionsAndRows(t *testing.T) {
	for _, dimensions := range [][2]int{{-1, 0}, {0, -1}, {int(^uint(0) >> 1), 2}} {
		if _, err := NewMatrix(dimensions[0], dimensions[1]); err == nil {
			t.Errorf("NewMatrix(%d, %d) error = nil", dimensions[0], dimensions[1])
		}
	}
	if _, err := NewMatrixFromRows([][]float64{{1, 2}, {3}}); err == nil || !strings.Contains(err.Error(), "row 1") {
		t.Errorf("ragged NewMatrixFromRows() error = %v", err)
	}
}

func TestMatrixAtRejectsInvalidIndexes(t *testing.T) {
	matrix, err := NewMatrix(2, 3)
	if err != nil {
		t.Fatal(err)
	}
	for _, index := range [][2]int{{-1, 0}, {2, 0}, {0, -1}, {0, 3}} {
		if _, err := matrix.At(index[0], index[1]); err == nil {
			t.Errorf("At(%d, %d) error = nil", index[0], index[1])
		}
	}
}

func TestMatrixRowsOwnTheirStorage(t *testing.T) {
	matrix, err := NewMatrixFromRows([][]float64{{1, 2}})
	if err != nil {
		t.Fatal(err)
	}
	rows := matrix.Rows()
	rows[0][0] = 9
	if got, _ := matrix.At(0, 0); got != 1 {
		t.Errorf("At(0, 0) after Rows mutation = %v, want 1", got)
	}
}

func TestCMatrixRoundTripAndOwnership(t *testing.T) {
	want, err := NewMatrixFromRows([][]float64{{1, 2, 3}, {4, 5, 6}})
	if err != nil {
		t.Fatal(err)
	}
	cValue, err := newCMatrix(want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := cValue.matrix()
	if err != nil {
		t.Fatal(err)
	}
	cValue.close()

	if !reflect.DeepEqual(got.Rows(), want.Rows()) {
		t.Errorf("matrix round trip = %#v, want %#v", got.Rows(), want.Rows())
	}
	if value, err := got.At(1, 2); err != nil || value != 6 {
		t.Errorf("At() after C close = %v, %v", value, err)
	}
}

func TestCMatrixRoundTripZeroDimensions(t *testing.T) {
	for _, dimensions := range [][2]int{{0, 0}, {0, 3}, {3, 0}} {
		matrix, err := NewMatrix(dimensions[0], dimensions[1])
		if err != nil {
			t.Fatal(err)
		}
		cValue, err := newCMatrix(matrix)
		if err != nil {
			t.Fatal(err)
		}
		got, err := cValue.matrix()
		cValue.close()
		if err != nil {
			t.Fatal(err)
		}
		if rows, columns := got.Dims(); rows != dimensions[0] || columns != dimensions[1] {
			t.Errorf("round-trip Dims() = (%d, %d), want (%d, %d)", rows, columns, dimensions[0], dimensions[1])
		}
	}
}

func TestCMatrixRejectsInconsistentGoValue(t *testing.T) {
	_, err := newCMatrix(Matrix{rows: 2, columns: 2, values: []float64{1}})
	if err == nil || !strings.Contains(err.Error(), "has 1 values") {
		t.Errorf("newCMatrix() error = %v", err)
	}
}

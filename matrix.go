package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
//
// static void go_igraph_matrix_set_value(
//     igraph_matrix_t *matrix, igraph_int_t row, igraph_int_t column,
//     igraph_real_t value) {
//   MATRIX(*matrix, row, column) = value;
// }
//
// static igraph_real_t go_igraph_matrix_get_value(
//     const igraph_matrix_t *matrix, igraph_int_t row, igraph_int_t column) {
//   return MATRIX(*matrix, row, column);
// }
import "C"

import "fmt"

// Matrix is an immutable dense matrix with explicit row and column
// dimensions. Its zero value is a valid 0-by-0 matrix. Matrix does not expose
// the storage order used by either Go or igraph.
type Matrix struct {
	rows    int
	columns int
	values  []float64
}

// NewMatrix returns a rows-by-columns zero matrix as a Go-owned Matrix value.
// Zero-sized dimensions are valid, including matrices such as 0-by-3 and 3-by-0.
func NewMatrix(rows, columns int) (Matrix, error) {
	size, err := matrixSize(rows, columns)
	if err != nil {
		return Matrix{}, err
	}
	return Matrix{rows: rows, columns: columns, values: make([]float64, size)}, nil
}

// NewMatrixFromRows copies a rectangular slice of rows into a Go-owned Matrix.
// The values slice is borrowed and copied. A nil or empty input produces a 0-by-0
// matrix. Ragged rows are rejected.
func NewMatrixFromRows(values [][]float64) (Matrix, error) {
	if len(values) == 0 {
		return NewMatrix(0, 0)
	}
	columns := len(values[0])
	for row, valuesRow := range values {
		if len(valuesRow) != columns {
			return Matrix{}, fmt.Errorf(
				"igraph: matrix row %d has %d columns, want %d",
				row, len(valuesRow), columns,
			)
		}
	}
	matrix, err := NewMatrix(len(values), columns)
	if err != nil {
		return Matrix{}, err
	}
	for row, valuesRow := range values {
		copy(matrix.values[row*columns:(row+1)*columns], valuesRow)
	}
	return matrix, nil
}

// Dims returns the number of rows and columns.
func (m Matrix) Dims() (rows, columns int) {
	return m.rows, m.columns
}

// At returns the value at the zero-based row and column.
func (m Matrix) At(row, column int) (float64, error) {
	if row < 0 || row >= m.rows {
		return 0, fmt.Errorf("igraph: matrix row %d out of range [0, %d)", row, m.rows)
	}
	if column < 0 || column >= m.columns {
		return 0, fmt.Errorf("igraph: matrix column %d out of range [0, %d)", column, m.columns)
	}
	return m.values[row*m.columns+column], nil
}

// Rows returns a deep, Go-owned copy in row-major logical order. The returned
// value may be changed without affecting the Matrix.
func (m Matrix) Rows() [][]float64 {
	result := make([][]float64, m.rows)
	for row := range result {
		result[row] = append([]float64{}, m.values[row*m.columns:(row+1)*m.columns]...)
	}
	return result
}

func matrixSize(rows, columns int) (int, error) {
	if rows < 0 {
		return 0, fmt.Errorf("igraph: matrix row count must be non-negative: %d", rows)
	}
	if columns < 0 {
		return 0, fmt.Errorf("igraph: matrix column count must be non-negative: %d", columns)
	}
	if rows != 0 && columns > int(^uint(0)>>1)/rows {
		return 0, fmt.Errorf("igraph: matrix dimensions overflow: %d by %d", rows, columns)
	}
	return rows * columns, nil
}

// cMatrix owns an initialized igraph_matrix_t. Go values are copied during
// construction; no Go pointer is retained by C. Call close after every
// successful construction.
type cMatrix struct {
	value C.igraph_matrix_t
}

//igraph:internal igraph_matrix_init
func newCMatrix(matrix Matrix) (*cMatrix, error) {
	size, err := matrixSize(matrix.rows, matrix.columns)
	if err != nil {
		return nil, err
	}
	if len(matrix.values) != size {
		return nil, fmt.Errorf(
			"igraph: matrix has %d values, want %d for %d by %d dimensions",
			len(matrix.values), size, matrix.rows, matrix.columns,
		)
	}
	cRows, err := intToIgraphInt(matrix.rows, "matrix row count")
	if err != nil {
		return nil, err
	}
	cColumns, err := intToIgraphInt(matrix.columns, "matrix column count")
	if err != nil {
		return nil, err
	}
	result := &cMatrix{}
	if code := C.go_igraph_matrix_init(&result.value, cRows, cColumns); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("initialize matrix", int(code))
	}
	for row := 0; row < matrix.rows; row++ {
		for column := 0; column < matrix.columns; column++ {
			C.go_igraph_matrix_set_value(
				&result.value,
				C.igraph_int_t(row),
				C.igraph_int_t(column),
				C.igraph_real_t(matrix.values[row*matrix.columns+column]),
			)
		}
	}
	return result, nil
}

// matrix returns an independently owned Go value that remains valid after
// close.
//
//igraph:internal igraph_matrix_nrow
//igraph:internal igraph_matrix_ncol
func (m *cMatrix) matrix() (Matrix, error) {
	rows, err := igraphIntToInt(C.igraph_matrix_nrow(&m.value), "matrix row count")
	if err != nil {
		return Matrix{}, err
	}
	columns, err := igraphIntToInt(C.igraph_matrix_ncol(&m.value), "matrix column count")
	if err != nil {
		return Matrix{}, err
	}
	result, err := NewMatrix(rows, columns)
	if err != nil {
		return Matrix{}, err
	}
	for row := 0; row < rows; row++ {
		for column := 0; column < columns; column++ {
			result.values[row*columns+column] = float64(C.go_igraph_matrix_get_value(
				&m.value,
				C.igraph_int_t(row),
				C.igraph_int_t(column),
			))
		}
	}
	return result, nil
}

//igraph:internal igraph_matrix_destroy
func (m *cMatrix) close() {
	C.igraph_matrix_destroy(&m.value)
}

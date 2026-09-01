package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "assignment_cgo.h"
import "C"

import (
	"fmt"
	"math"
)

// SolveLinearAssignment solves a balanced linear assignment problem by
// minimizing total cost. Rows of costs are agents and columns are tasks; the
// returned Go-owned slice contains, for each row, the zero-based column
// assigned to it. The immutable input Matrix is borrowed only for the call and
// is never retained or modified. Negative finite costs are supported, so
// maximizing a score can be expressed by negating it before construction.
//
// A zero-by-zero matrix returns a non-nil empty slice. Non-square matrices and
// matrices containing NaN or infinity are rejected.
//
//igraph:bind igraph_solve_lsap
func SolveLinearAssignment(costs Matrix) ([]int, error) {
	return solveLinearAssignment(costs, nil)
}

type assignmentAdapters struct {
	newMatrix func(Matrix) (*cMatrix, error)
	newInt    func([]int) (*intVector, error)
	call      func(*cMatrix, int, *intVector) int
	convert   func(*intVector) ([]int, error)
}

func defaultAssignmentAdapters() assignmentAdapters {
	return assignmentAdapters{
		newMatrix: newCMatrix,
		newInt:    newIntVector,
		call: func(costs *cMatrix, size int, result *intVector) int {
			return int(C.go_igraph_solve_lsap(&costs.value, C.igraph_int_t(size), &result.value))
		},
		convert: (*intVector).slice,
	}
}

func solveLinearAssignment(costs Matrix, adapters *assignmentAdapters) ([]int, error) {
	rows, columns := costs.Dims()
	if rows != columns {
		return nil, fmt.Errorf("igraph: assignment cost matrix must be square: %d by %d", rows, columns)
	}
	size, err := matrixSize(rows, columns)
	if err != nil {
		return nil, err
	}
	if len(costs.values) != size {
		return nil, fmt.Errorf(
			"igraph: matrix has %d values, want %d for %d by %d dimensions",
			len(costs.values), size, rows, columns,
		)
	}
	if _, err := intToIgraphInt(rows, "assignment size"); err != nil {
		return nil, err
	}
	for index, value := range costs.values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("igraph: assignment cost at row %d column %d must be finite: %v", index/columns, index%columns, value)
		}
	}
	if rows == 0 {
		return []int{}, nil
	}

	resolved := defaultAssignmentAdapters()
	if adapters != nil {
		resolved = *adapters
	}
	cCosts, err := resolved.newMatrix(costs)
	if err != nil {
		return nil, err
	}
	defer cCosts.close()
	result, err := resolved.newInt(nil)
	if err != nil {
		return nil, err
	}
	defer result.close()
	if code := resolved.call(cCosts, rows, result); code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("solve linear assignment", code)
	}
	assignment, err := resolved.convert(result)
	if err != nil {
		return nil, err
	}
	if err := validateAssignment(assignment, rows); err != nil {
		return nil, err
	}
	return assignment, nil
}

func validateAssignment(assignment []int, size int) error {
	if len(assignment) != size {
		return fmt.Errorf("igraph: assignment result has length %d, want %d", len(assignment), size)
	}
	seen := make([]bool, size)
	for row, column := range assignment {
		if column < 0 || column >= size {
			return fmt.Errorf("igraph: assignment result for row %d has column %d outside [0, %d)", row, column, size)
		}
		if seen[column] {
			return fmt.Errorf("igraph: assignment result repeats column %d", column)
		}
		seen[column] = true
	}
	return nil
}

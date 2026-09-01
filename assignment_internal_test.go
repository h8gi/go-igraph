package igraph

import (
	"errors"
	"testing"
)

func TestSolveLinearAssignmentFailureAdapters(t *testing.T) {
	costs, _ := NewMatrixFromRows([][]float64{{1, 2}, {2, 1}})
	forced := errors.New("forced assignment failure")

	for name, mutate := range map[string]func(*assignmentAdapters){
		"matrix initialization": func(a *assignmentAdapters) {
			a.newMatrix = func(Matrix) (*cMatrix, error) { return nil, forced }
		},
		"result initialization": func(a *assignmentAdapters) {
			a.newInt = func([]int) (*intVector, error) { return nil, forced }
		},
		"upstream": func(a *assignmentAdapters) {
			a.call = func(*cMatrix, int, *intVector) int { return 4 }
		},
		"conversion": func(a *assignmentAdapters) {
			a.convert = func(*intVector) ([]int, error) { return nil, forced }
		},
		"malformed output": func(a *assignmentAdapters) {
			a.convert = func(*intVector) ([]int, error) { return []int{0, 0}, nil }
		},
	} {
		t.Run(name, func(t *testing.T) {
			adapters := defaultAssignmentAdapters()
			mutate(&adapters)
			if _, err := solveLinearAssignment(costs, &adapters); err == nil {
				t.Fatal("failure not propagated")
			}
		})
	}
}

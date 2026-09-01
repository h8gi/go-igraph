package igraph

import (
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestSolveLinearAssignmentKnownAnswerAndInputOwnership(t *testing.T) {
	costs, err := NewMatrixFromRows([][]float64{
		{4, 1, 3},
		{2, 0, 5},
		{3, 2, 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	before := costs.Rows()
	got, err := SolveLinearAssignment(costs)
	if err != nil {
		t.Fatal(err)
	}
	if want := []int{1, 0, 2}; !reflect.DeepEqual(got, want) {
		t.Fatalf("assignment = %v, want %v", got, want)
	}
	if !reflect.DeepEqual(costs.Rows(), before) {
		t.Fatalf("input changed: got %v, want %v", costs.Rows(), before)
	}
	got[0] = 2
	again, err := SolveLinearAssignment(costs)
	if err != nil || !reflect.DeepEqual(again, []int{1, 0, 2}) {
		t.Fatalf("second assignment = %v, %v", again, err)
	}
}

func TestSolveLinearAssignmentEmptySingletonNegativeAndTie(t *testing.T) {
	empty, err := SolveLinearAssignment(Matrix{})
	if err != nil || empty == nil || len(empty) != 0 {
		t.Fatalf("empty = %#v, %v", empty, err)
	}
	singleton, _ := NewMatrixFromRows([][]float64{{-7}})
	if got, err := SolveLinearAssignment(singleton); err != nil || !reflect.DeepEqual(got, []int{0}) {
		t.Fatalf("singleton = %v, %v", got, err)
	}
	large, _ := NewMatrixFromRows([][]float64{{math.MaxFloat64}})
	if got, err := SolveLinearAssignment(large); err != nil || !reflect.DeepEqual(got, []int{0}) {
		t.Fatalf("large finite cost = %v, %v", got, err)
	}
	tie, _ := NewMatrixFromRows([][]float64{{1, 1}, {1, 1}})
	got, err := SolveLinearAssignment(tie)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateAssignment(got, 2); err != nil {
		t.Fatalf("tie result = %v: %v", got, err)
	}
}

func TestSolveLinearAssignmentValidation(t *testing.T) {
	nonSquare, _ := NewMatrixFromRows([][]float64{{1, 2}})
	for name, matrix := range map[string]Matrix{
		"non-square":        nonSquare,
		"nan":               {rows: 1, columns: 1, values: []float64{math.NaN()}},
		"positive infinity": {rows: 1, columns: 1, values: []float64{math.Inf(1)}},
		"negative infinity": {rows: 1, columns: 1, values: []float64{math.Inf(-1)}},
		"malformed storage": {rows: 2, columns: 2, values: []float64{1}},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := SolveLinearAssignment(matrix); err == nil {
				t.Fatal("invalid matrix accepted")
			}
		})
	}
}

func TestValidateAssignmentRejectsMalformedResults(t *testing.T) {
	for name, assignment := range map[string][]int{
		"length":    {0},
		"negative":  {-1, 1},
		"too-large": {0, 2},
		"duplicate": {1, 1},
	} {
		t.Run(name, func(t *testing.T) {
			if err := validateAssignment(assignment, 2); err == nil || !strings.Contains(err.Error(), "assignment result") {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestSolveLinearAssignmentConcurrent(t *testing.T) {
	costs, _ := NewMatrixFromRows([][]float64{{4, 1, 3}, {2, 0, 5}, {3, 2, 2}})
	errors := make(chan error, 16)
	for range 16 {
		go func() {
			got, err := SolveLinearAssignment(costs)
			if err == nil && !reflect.DeepEqual(got, []int{1, 0, 2}) {
				err = &assignmentTestError{got: got}
			}
			errors <- err
		}()
	}
	for range 16 {
		if err := <-errors; err != nil {
			t.Fatal(err)
		}
	}
}

type assignmentTestError struct{ got []int }

func (err *assignmentTestError) Error() string { return "unexpected assignment" }

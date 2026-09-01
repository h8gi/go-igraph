package igraph_test

import (
	"fmt"
	"reflect"
	"sync"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestMilestone29StatisticalAssignmentWorkflow(t *testing.T) {
	data := []float64{1, 1, 2, 2, 3, 4, 6, 10, 20, 40}
	means, err := igraph.RunningMean(data, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(means) != len(data)-1 || means[0] != 1 || means[len(means)-1] != 30 {
		t.Fatalf("running means = %v", means)
	}

	seed := uint64(29)
	indices, err := igraph.SampleIntegers(0, len(means)-1, 4, igraph.IntegerSampleOptions{Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	if len(indices) != 4 {
		t.Fatalf("sampled indices = %v", indices)
	}

	fit, err := igraph.FitPowerLaw(data, igraph.PowerLawFitOptions{})
	if err != nil {
		t.Fatal(err)
	}
	pValue, err := fit.PValue(igraph.PowerLawPValueOptions{Precision: 0.5, Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	if fit.Alpha <= 1 || pValue < 0 || pValue > 1 {
		t.Fatalf("fit = %+v, p-value = %v", fit, pValue)
	}

	costs, err := igraph.NewMatrixFromRows([][]float64{{4, 1, 3}, {2, 0, 5}, {3, 2, 2}})
	if err != nil {
		t.Fatal(err)
	}
	assignment, err := igraph.SolveLinearAssignment(costs)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(assignment, []int{1, 0, 2}) {
		t.Fatalf("assignment = %v", assignment)
	}
	total := 0.0
	for row, column := range assignment {
		cost, err := costs.At(row, column)
		if err != nil {
			t.Fatal(err)
		}
		total += cost
	}
	if total != 5 {
		t.Fatalf("assignment cost = %v", total)
	}
}

func TestMilestone29ConcurrentWorkflows(t *testing.T) {
	data := []float64{1, 1, 2, 2, 3, 4, 6, 10, 20, 40}
	fit, err := igraph.FitPowerLaw(data, igraph.PowerLawFitOptions{})
	if err != nil {
		t.Fatal(err)
	}
	costs, err := igraph.NewMatrixFromRows([][]float64{{4, 1, 3}, {2, 0, 5}, {3, 2, 2}})
	if err != nil {
		t.Fatal(err)
	}

	var wait sync.WaitGroup
	errors := make(chan error, 4)
	for worker := range 4 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			if _, err := igraph.RunningMean(data, 3); err != nil {
				errors <- err
				return
			}
			seed := uint64(worker + 1)
			if _, err := igraph.SampleIntegers(0, 100, 8, igraph.IntegerSampleOptions{Seed: &seed}); err != nil {
				errors <- err
				return
			}
			if _, err := fit.PValue(igraph.PowerLawPValueOptions{Precision: 0.5, Seed: &seed}); err != nil {
				errors <- err
				return
			}
			if assignment, err := igraph.SolveLinearAssignment(costs); err != nil {
				errors <- err
			} else if !reflect.DeepEqual(assignment, []int{1, 0, 2}) {
				errors <- fmt.Errorf("assignment = %v", assignment)
			}
		}()
	}
	wait.Wait()
	close(errors)
	for err := range errors {
		t.Error(err)
	}
}

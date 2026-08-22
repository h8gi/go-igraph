package igraph

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestCategoricalAssortativity(t *testing.T) {
	graph, err := NewGraphFromEdges(4, []Edge{{0, 1}, {1, 2}, {2, 3}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()

	for name, categories := range map[string]CategoryValues{
		"integers": IntegerCategories([]int{10, 10, -3, -3}),
		"strings":  StringCategories([]string{"a", "a", "b", "b"}),
	} {
		normalized, err := graph.CategoricalAssortativity(categories, CategoricalAssortativityOptions{Normalized: true})
		if err != nil {
			t.Fatalf("%s normalized error = %v", name, err)
		}
		if math.Abs(normalized-1.0/3.0) > 1e-12 {
			t.Errorf("%s normalized = %v, want 1/3", name, normalized)
		}
		raw, err := graph.CategoricalAssortativity(categories, CategoricalAssortativityOptions{})
		if err != nil {
			t.Fatalf("%s raw error = %v", name, err)
		}
		if math.Abs(raw-1.0/6.0) > 1e-12 {
			t.Errorf("%s raw = %v, want 1/6", name, raw)
		}
	}

	original := []int{0, 0, 1, 1}
	categories := IntegerCategories(original)
	original[0] = 99
	if compact, err := categories.compact(4); err != nil || !reflect.DeepEqual(compact, []int{0, 0, 1, 1}) {
		t.Errorf("copied categories compact = %v, %v", compact, err)
	}
}

func TestNumericAndDegreeAssortativity(t *testing.T) {
	graph, err := NewGraphFromEdges(4, []Edge{{0, 1}, {1, 2}, {2, 3}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()

	values := []float64{0, 1, 2, 3}
	normalized, err := graph.NumericAssortativity(values, NumericAssortativityOptions{Normalized: true})
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(normalized-5.0/11.0) > 1e-12 {
		t.Errorf("normalized numeric = %v, want 5/11", normalized)
	}
	raw, err := graph.NumericAssortativity(values, NumericAssortativityOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(raw-5.0/12.0) > 1e-12 {
		t.Errorf("raw numeric = %v, want 5/12", raw)
	}
	weighted, err := graph.NumericAssortativity(values, NumericAssortativityOptions{
		Weights: []float64{1, 2, 3}, Normalized: true,
	})
	if err != nil || math.IsNaN(weighted) {
		t.Errorf("weighted numeric = %v, %v", weighted, err)
	}

	degrees, err := graph.Degree(AllVertices(), DegreeOptions{Direction: DirectionAll, CountLoops: true})
	if err != nil {
		t.Fatal(err)
	}
	degreeValues := make([]float64, len(degrees))
	for index, degree := range degrees {
		degreeValues[index] = float64(degree)
	}
	fromValues, err := graph.NumericAssortativity(degreeValues, NumericAssortativityOptions{Normalized: true})
	if err != nil {
		t.Fatal(err)
	}
	fromDegree, err := graph.DegreeAssortativity(false)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(fromValues-fromDegree) > 1e-12 {
		t.Errorf("degree assortativity = %v, numeric degree reference = %v", fromDegree, fromValues)
	}
}

func TestDirectedNumericAssortativity(t *testing.T) {
	graph, err := NewGraphFromEdges(3, []Edge{{0, 1}, {0, 2}, {1, 2}}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	result, err := graph.NumericAssortativity(
		[]float64{1, 2, 3},
		NumericAssortativityOptions{
			TargetValues: []float64{4, 5, 6},
			Weights:      []float64{1, 2, 1},
			Directed:     true,
			Normalized:   true,
		},
	)
	if err != nil || math.IsNaN(result) {
		t.Errorf("directed numeric = %v, %v", result, err)
	}
	if _, err := graph.DegreeAssortativity(true); err != nil {
		t.Errorf("directed degree error = %v", err)
	}
}

func TestAssortativityValidationFailuresAndUndefined(t *testing.T) {
	graph, err := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}

	badCategories := []CategoryValues{
		IntegerCategories([]int{0}),
		{kind: categoryKind(99)},
	}
	for _, categories := range badCategories {
		if _, err := graph.CategoricalAssortativity(categories, CategoricalAssortativityOptions{}); err == nil {
			t.Errorf("CategoricalAssortativity(%v) error = nil", categories)
		}
	}
	for _, values := range [][]float64{{1}, {1, math.NaN()}, {1, math.Inf(1)}} {
		if _, err := graph.NumericAssortativity(values, NumericAssortativityOptions{}); err == nil {
			t.Errorf("NumericAssortativity(%v) error = nil", values)
		}
	}
	if _, err := graph.NumericAssortativity([]float64{1, 2}, NumericAssortativityOptions{TargetValues: []float64{1, 2}}); err == nil {
		t.Error("undirected target values error = nil")
	}
	if _, err := graph.NumericAssortativity([]float64{1, 2}, NumericAssortativityOptions{Directed: true, TargetValues: []float64{1}}); err == nil {
		t.Error("target length error = nil")
	}
	if _, err := graph.NumericAssortativity([]float64{1, 2}, NumericAssortativityOptions{Weights: []float64{1, 2}}); err == nil {
		t.Error("weight length error = nil")
	}

	operationError := errors.New("assortativity operation failed")
	hooks := assortativityHooks{run: func() error { return operationError }}
	if _, err := graph.categoricalAssortativity(IntegerCategories([]int{0, 1}), CategoricalAssortativityOptions{}, hooks); !errors.Is(err, operationError) {
		t.Errorf("categorical operation error = %v", err)
	}
	if _, err := graph.numericAssortativity([]float64{1, 2}, NumericAssortativityOptions{}, hooks); !errors.Is(err, operationError) {
		t.Errorf("numeric operation error = %v", err)
	}
	if _, err := graph.degreeAssortativity(false, hooks); !errors.Is(err, operationError) {
		t.Errorf("degree operation error = %v", err)
	}

	undefined, err := graph.NumericAssortativity([]float64{1, 1}, NumericAssortativityOptions{Normalized: true})
	if err != nil || !math.IsNaN(undefined) {
		t.Errorf("constant numeric = %v, %v, want NaN", undefined, err)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	for _, call := range []func() error{
		func() error {
			_, err := graph.CategoricalAssortativity(IntegerCategories([]int{0, 1}), CategoricalAssortativityOptions{})
			return err
		},
		func() error {
			_, err := graph.NumericAssortativity([]float64{1, 2}, NumericAssortativityOptions{})
			return err
		},
		func() error { _, err := graph.DegreeAssortativity(false); return err },
	} {
		if err := call(); !errors.Is(err, ErrClosed) {
			t.Errorf("closed graph error = %v", err)
		}
	}
	var nilGraph *Graph
	if _, err := nilGraph.DegreeAssortativity(false); !errors.Is(err, ErrClosed) {
		t.Errorf("nil graph error = %v", err)
	}
}

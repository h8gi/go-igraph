package igraph

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestCategoricalJointDistributionAxesAndValues(t *testing.T) {
	g, err := NewGraphFromEdges(3, []Edge{{0, 1}, {0, 2}, {2, 1}}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	categories := StringCategories([]string{"a", "b", "a"})
	target := IntegerCategories([]int{10, 10, 20})
	result, err := g.CategoricalJointDistribution(categories, CategoryJointDistributionOptions{TargetCategories: &target, Directed: true})
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := result.RowCategories.StringValues(); !ok || !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("row categories = %v, %v", got, ok)
	}
	if got, ok := result.ColumnCategories.IntegerValues(); !ok || !reflect.DeepEqual(got, []int{10, 20}) {
		t.Fatalf("column categories = %v, %v", got, ok)
	}
	if got := result.Matrix.Rows(); !reflect.DeepEqual(got, [][]float64{{2, 1}, {0, 0}}) {
		t.Fatalf("matrix = %v", got)
	}
	rowLabels, _ := result.RowCategories.StringValues()
	rowLabels[0] = "changed"
	if got, _ := result.RowCategories.StringValues(); got[0] != "a" {
		t.Fatalf("axis aliases caller: %v", got)
	}
}

func TestCategoricalJointDistributionNormalizedWeights(t *testing.T) {
	g, err := NewGraphFromEdges(2, []Edge{{0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	result, err := g.CategoricalJointDistribution(IntegerCategories([]int{7, 9}), CategoryJointDistributionOptions{Weights: []float64{2}, Normalized: true})
	if err != nil {
		t.Fatal(err)
	}
	if got := result.Matrix.Rows(); !reflect.DeepEqual(got, [][]float64{{0, .5}, {.5, 0}}) {
		t.Fatalf("matrix = %v", got)
	}
	if _, err := g.CategoricalJointDistribution(IntegerCategories([]int{7, 9}), CategoryJointDistributionOptions{Weights: []float64{-1}, Normalized: true}); err == nil {
		t.Fatal("expected negative normalized weight error")
	}
	if _, err := g.CategoricalJointDistribution(IntegerCategories([]int{7, 9}), CategoryJointDistributionOptions{Weights: []float64{0}, Normalized: true}); err == nil {
		t.Fatal("expected zero-total error")
	}
	target := IntegerCategories([]int{1, 2})
	if _, err := g.CategoricalJointDistribution(IntegerCategories([]int{1, 2}), CategoryJointDistributionOptions{TargetCategories: &target}); err == nil {
		t.Fatal("expected undirected target error")
	}
	directed, err := NewGraphFromEdges(2, []Edge{{0, 1}}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer directed.Close()
	if _, err := directed.CategoricalJointDistribution(IntegerCategories([]int{1, 2}), CategoryJointDistributionOptions{TargetCategories: &target}); err == nil {
		t.Fatal("expected non-directed mixing target error")
	}
}

func TestDegreeJointDistributions(t *testing.T) {
	g, err := NewGraphFromEdges(3, []Edge{{0, 1}, {0, 2}, {2, 1}}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	distribution, err := g.DegreeJointDistribution(DegreeJointDistributionOptions{FromMode: DirectionOut, ToMode: DirectionIn, DirectedNeighbors: true})
	if err != nil {
		t.Fatal(err)
	}
	if got := distribution.Rows(); !reflect.DeepEqual(got, [][]float64{{0, 0, 0}, {0, 0, 1}, {0, 1, 1}}) {
		t.Fatalf("distribution = %v", got)
	}
	matrix, err := g.JointDegreeMatrix(JointDegreeMatrixOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if got := matrix.Rows(); !reflect.DeepEqual(got, [][]float64{{0, 1}, {1, 1}}) {
		t.Fatalf("joint degree matrix = %v", got)
	}
	maximum := 1
	truncated, err := g.DegreeJointDistribution(DegreeJointDistributionOptions{FromMode: DirectionOut, ToMode: DirectionIn, DirectedNeighbors: true, MaximumFromDegree: &maximum, MaximumToDegree: &maximum})
	if err != nil {
		t.Fatal(err)
	}
	if got := truncated.Rows(); !reflect.DeepEqual(got, [][]float64{{0, 0}, {0, 0}}) {
		t.Fatalf("truncated = %v", got)
	}
}

func TestJointDistributionValidationAndFailures(t *testing.T) {
	g, err := NewGraphFromEdges(2, []Edge{{0, 1}}, true)
	if err != nil {
		t.Fatal(err)
	}
	bad := -1
	if _, err := g.DegreeJointDistribution(DegreeJointDistributionOptions{MaximumFromDegree: &bad}); err == nil {
		t.Fatal("expected degree limit error")
	}
	if _, err := g.DegreeJointDistribution(DegreeJointDistributionOptions{FromMode: DirectionMode(99)}); err == nil {
		t.Fatal("expected mode error")
	}
	if _, err := g.DegreeJointDistribution(DegreeJointDistributionOptions{ToMode: DirectionMode(99)}); err == nil {
		t.Fatal("expected target mode error")
	}
	if _, err := g.DegreeJointDistribution(DegreeJointDistributionOptions{FromMode: DirectionOut, ToMode: DirectionIn}); err == nil {
		t.Fatal("expected unsafe reciprocal mode error")
	}
	if _, err := g.JointDegreeMatrix(JointDegreeMatrixOptions{MaximumOutDegree: &bad}); err == nil {
		t.Fatal("expected matrix degree limit error")
	}
	if _, err := g.JointDegreeMatrix(JointDegreeMatrixOptions{Weights: []float64{math.NaN()}}); err == nil {
		t.Fatal("expected weight error")
	}
	failure := errors.New("failure")
	if _, err := g.categoricalJointDistribution(IntegerCategories([]int{0, 1}), CategoryJointDistributionOptions{}, jointDistributionHooks{newMatrix: func(Matrix) (*cMatrix, error) { return nil, failure }}); !errors.Is(err, failure) {
		t.Fatalf("initialization error = %v", err)
	}
	if _, err := g.degreeJointDistribution(DegreeJointDistributionOptions{}, jointDistributionHooks{run: func() error { return failure }}); !errors.Is(err, failure) {
		t.Fatalf("operation error = %v", err)
	}
	if _, err := g.jointDegreeMatrix(JointDegreeMatrixOptions{}, jointDistributionHooks{newMatrix: func(Matrix) (*cMatrix, error) { return nil, failure }}); !errors.Is(err, failure) {
		t.Fatalf("matrix initialization error = %v", err)
	}
	if _, err := g.jointDegreeMatrix(JointDegreeMatrixOptions{}, jointDistributionHooks{run: func() error { return failure }}); !errors.Is(err, failure) {
		t.Fatalf("matrix operation error = %v", err)
	}
	var nilGraph *Graph
	if _, err := nilGraph.CategoricalJointDistribution(IntegerCategories(nil), CategoryJointDistributionOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil categorical error = %v", err)
	}
	if _, err := nilGraph.DegreeJointDistribution(DegreeJointDistributionOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil degree error = %v", err)
	}
	if _, err := nilGraph.JointDegreeMatrix(JointDegreeMatrixOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil matrix error = %v", err)
	}
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := g.CategoricalJointDistribution(IntegerCategories([]int{0, 1}), CategoryJointDistributionOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed categorical error = %v", err)
	}
	if _, err := g.DegreeJointDistribution(DegreeJointDistributionOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed degree error = %v", err)
	}
	if _, err := g.JointDegreeMatrix(JointDegreeMatrixOptions{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed matrix error = %v", err)
	}
}

func TestCategoryValueAccessorKinds(t *testing.T) {
	integers := IntegerCategories([]int{1})
	if _, ok := integers.StringValues(); ok {
		t.Fatal("integer categories reported as strings")
	}
	strings := StringCategories([]string{"x"})
	if _, ok := strings.IntegerValues(); ok {
		t.Fatal("string categories reported as integers")
	}
}

func TestJointDistributionsEmptyGraph(t *testing.T) {
	g, err := NewGraphFromEdges(0, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	category, err := g.CategoricalJointDistribution(IntegerCategories(nil), CategoryJointDistributionOptions{Normalized: true})
	if err != nil {
		t.Fatal(err)
	}
	if rows, columns := category.Matrix.Dims(); rows != 0 || columns != 0 {
		t.Fatalf("category dimensions = %d by %d", rows, columns)
	}
	degree, err := g.DegreeJointDistribution(DegreeJointDistributionOptions{Normalized: true})
	if err != nil {
		t.Fatal(err)
	}
	if rows, columns := degree.Dims(); rows != 0 || columns != 0 {
		t.Fatalf("degree dimensions = %d by %d", rows, columns)
	}
}

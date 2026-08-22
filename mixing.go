package igraph

/*
#cgo pkg-config: igraph
#include <igraph.h>
#include "mixing_cgo.h"
*/
import "C"

import (
	"fmt"
	"math"
)

type categoryKind uint8

const (
	categoryIntegers categoryKind = iota
	categoryStrings
)

// CategoryValues is an immutable Go-owned categorical vertex-value vector.
// Its zero value is an empty integer category vector.
type CategoryValues struct {
	kind     categoryKind
	integers []int
	strings  []string
}

// IntegerCategories copies integer category labels. Labels need not be
// contiguous or non-negative; calls compact them in stable first-occurrence
// order before entering C.
func IntegerCategories(values []int) CategoryValues {
	return CategoryValues{integers: append([]int{}, values...)}
}

// StringCategories copies string category labels. Empty strings are valid
// labels and calls compact labels in stable first-occurrence order.
func StringCategories(values []string) CategoryValues {
	return CategoryValues{kind: categoryStrings, strings: append([]string{}, values...)}
}

// IntegerValues returns a Go-owned copy of integer labels and true when this
// value uses integer labels.
func (categories CategoryValues) IntegerValues() ([]int, bool) {
	if categories.kind != categoryIntegers {
		return nil, false
	}
	return append([]int{}, categories.integers...), true
}

// StringValues returns a Go-owned copy of string labels and true when this
// value uses string labels.
func (categories CategoryValues) StringValues() ([]string, bool) {
	if categories.kind != categoryStrings {
		return nil, false
	}
	return append([]string{}, categories.strings...), true
}

// CategoryJointDistribution contains a Go-owned mixing matrix and its row and
// column labels. Matrix rows are source categories and columns are target
// categories, each compacted independently in stable first-occurrence order.
type CategoryJointDistribution struct {
	Matrix           Matrix
	RowCategories    CategoryValues
	ColumnCategories CategoryValues
}

// CategoryJointDistributionOptions controls categorical mixing. A nil target
// reuses Categories for both endpoints. Weights and TargetCategories are
// borrowed for the synchronous call. Normalization requires non-negative
// weights with a positive total whenever the graph has edges.
type CategoryJointDistributionOptions struct {
	TargetCategories *CategoryValues
	Weights          []float64
	Directed         bool
	Normalized       bool
}

// DegreeJointDistributionOptions controls ordered endpoint-degree mixing.
// Nil maximums select the observed maxima; non-nil values include degrees from
// zero through the specified maximum. Direction modes and DirectedNeighbors
// are ignored for undirected graphs. Weights and maximum pointers are borrowed
// only for the synchronous call. Normalization uses the same non-negative,
// positive-total weight contract as CategoryJointDistributionOptions.
type DegreeJointDistributionOptions struct {
	Weights           []float64
	FromMode          DirectionMode
	ToMode            DirectionMode
	DirectedNeighbors bool
	Normalized        bool
	MaximumFromDegree *int
	MaximumToDegree   *int
}

// JointDegreeMatrixOptions controls edge-count degree mixing. Rows represent
// source/out-degrees and columns target/in-degrees in directed graphs. In an
// undirected graph both axes represent total degree. Degree zero is excluded:
// matrix index zero represents degree one. Weights and maximum pointers are
// borrowed only for the synchronous call.
type JointDegreeMatrixOptions struct {
	Weights          []float64
	MaximumOutDegree *int
	MaximumInDegree  *int
}

type jointDistributionHooks struct {
	newMatrix func(Matrix) (*cMatrix, error)
	run       func() error
}

// CategoricalAssortativityOptions controls nominal assortativity. Directed is
// ignored for an undirected graph. Normalized returns the standard coefficient;
// false returns the corresponding modularity-like unnormalized value. Pinned
// igraph 1.0.1 does not implement weighted nominal assortativity.
type CategoricalAssortativityOptions struct {
	Directed   bool
	Normalized bool
}

// NumericAssortativityOptions controls continuous-value assortativity. A nil
// TargetValues slice reuses the source values for edge targets. A non-nil
// target slice is valid only when Directed is true. Weights and TargetValues
// are borrowed for the synchronous call; weights contain one finite value per
// edge and target values one finite value per vertex.
type NumericAssortativityOptions struct {
	TargetValues []float64
	Weights      []float64
	Directed     bool
	Normalized   bool
}

type assortativityHooks struct {
	run func() error
}

// CategoricalAssortativity computes nominal mixing from one category per
// vertex. Categories and options are borrowed for the synchronous call; the
// returned scalar is Go-owned. Undefined results remain NaN.
//
//igraph:bind igraph_assortativity_nominal
func (g *Graph) CategoricalAssortativity(categories CategoryValues, options CategoricalAssortativityOptions) (float64, error) {
	return g.categoricalAssortativity(categories, options, assortativityHooks{})
}

func (g *Graph) categoricalAssortativity(categories CategoryValues, options CategoricalAssortativityOptions, hooks assortativityHooks) (float64, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return 0, ErrClosed
	}
	values, err := categories.compact(int(C.igraph_vcount(&g.graph)))
	if err != nil {
		return 0, err
	}
	categoriesVector, err := newIntVector(values)
	if err != nil {
		return 0, err
	}
	defer categoriesVector.close()
	var result C.igraph_real_t
	run := func() error {
		code := C.go_igraph_assortativity_nominal(
			&g.graph, &categoriesVector.value, &result,
			booltoint(options.Directed), booltoint(options.Normalized),
		)
		if code != C.IGRAPH_SUCCESS {
			return igraphError("calculate categorical assortativity", int(code))
		}
		return nil
	}
	if hooks.run != nil {
		run = hooks.run
	}
	if err := run(); err != nil {
		return 0, err
	}
	return float64(result), nil
}

// NumericAssortativity computes continuous-value mixing. Values contain one
// finite source value per vertex and are borrowed only for the synchronous
// call. Undefined covariance or correlation remains NaN.
//
//igraph:bind igraph_assortativity
func (g *Graph) NumericAssortativity(values []float64, options NumericAssortativityOptions) (float64, error) {
	return g.numericAssortativity(values, options, assortativityHooks{})
}

func (g *Graph) numericAssortativity(values []float64, options NumericAssortativityOptions, hooks assortativityHooks) (float64, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return 0, ErrClosed
	}
	vertexCount := int(C.igraph_vcount(&g.graph))
	if err := validateFiniteVertexValues(values, vertexCount, "source"); err != nil {
		return 0, err
	}
	if options.TargetValues != nil {
		if !options.Directed {
			return 0, fmt.Errorf("igraph: target assortativity values require directed mixing")
		}
		if err := validateFiniteVertexValues(options.TargetValues, vertexCount, "target"); err != nil {
			return 0, err
		}
	}
	source, err := newRealVector(values)
	if err != nil {
		return 0, err
	}
	defer source.close()
	var target *realVector
	if options.TargetValues != nil {
		target, err = newRealVector(options.TargetValues)
		if err != nil {
			return 0, err
		}
		defer target.close()
	}
	weights, err := newOptionalEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return 0, err
	}
	if weights != nil {
		defer weights.close()
	}
	var result C.igraph_real_t
	run := func() error {
		code := C.go_igraph_assortativity(
			&g.graph, edgeWeightPointer(weights), &source.value,
			realVectorPointer(target), &result,
			booltoint(options.Directed), booltoint(options.Normalized),
		)
		if code != C.IGRAPH_SUCCESS {
			return igraphError("calculate numeric assortativity", int(code))
		}
		return nil
	}
	if hooks.run != nil {
		run = hooks.run
	}
	if err := run(); err != nil {
		return 0, err
	}
	return float64(result), nil
}

// DegreeAssortativity computes the standard normalized degree assortativity.
// Directed uses out-degree at edge sources and in-degree at edge targets; it
// is ignored on undirected graphs. Undefined results remain NaN.
//
//igraph:bind igraph_assortativity_degree
func (g *Graph) DegreeAssortativity(directed bool) (float64, error) {
	return g.degreeAssortativity(directed, assortativityHooks{})
}

func (g *Graph) degreeAssortativity(directed bool, hooks assortativityHooks) (float64, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return 0, ErrClosed
	}
	var result C.igraph_real_t
	run := func() error {
		if code := C.go_igraph_assortativity_degree(&g.graph, &result, booltoint(directed)); code != C.IGRAPH_SUCCESS {
			return igraphError("calculate degree assortativity", int(code))
		}
		return nil
	}
	if hooks.run != nil {
		run = hooks.run
	}
	if err := run(); err != nil {
		return 0, err
	}
	return float64(result), nil
}

// CategoricalJointDistribution computes endpoint-category mixing. Categories,
// options, target categories, and weights are borrowed only for the call; all
// returned values are independent Go-owned copies.
//
//igraph:bind igraph_joint_type_distribution
func (g *Graph) CategoricalJointDistribution(categories CategoryValues, options CategoryJointDistributionOptions) (CategoryJointDistribution, error) {
	return g.categoricalJointDistribution(categories, options, jointDistributionHooks{})
}

func (g *Graph) categoricalJointDistribution(categories CategoryValues, options CategoryJointDistributionOptions, hooks jointDistributionHooks) (CategoryJointDistribution, error) {
	if g == nil {
		return CategoryJointDistribution{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return CategoryJointDistribution{}, ErrClosed
	}
	vertexCount := int(C.igraph_vcount(&g.graph))
	fromValues, rowAxis, err := categories.compactWithAxis(vertexCount)
	if err != nil {
		return CategoryJointDistribution{}, err
	}
	toValues, columnAxis := fromValues, rowAxis
	if options.TargetCategories != nil {
		if !options.Directed || C.go_igraph_is_directed(&g.graph) == 0 {
			return CategoryJointDistribution{}, fmt.Errorf("igraph: target categories require directed mixing on a directed graph")
		}
		toValues, columnAxis, err = options.TargetCategories.compactWithAxis(vertexCount)
		if err != nil {
			return CategoryJointDistribution{}, err
		}
	}
	if _, err := matrixSize(categoryCount(rowAxis), categoryCount(columnAxis)); err != nil {
		return CategoryJointDistribution{}, err
	}
	weights, err := newJointWeights(options.Weights, int(C.igraph_ecount(&g.graph)), options.Normalized)
	if err != nil {
		return CategoryJointDistribution{}, err
	}
	if weights != nil {
		defer weights.close()
	}
	from, err := newIntVector(fromValues)
	if err != nil {
		return CategoryJointDistribution{}, err
	}
	defer from.close()
	var to *intVector
	if options.TargetCategories != nil {
		to, err = newIntVector(toValues)
		if err != nil {
			return CategoryJointDistribution{}, err
		}
		defer to.close()
	}
	newMatrix := hooks.newMatrix
	if newMatrix == nil {
		newMatrix = newCMatrix
	}
	result, err := newMatrix(Matrix{})
	if err != nil {
		return CategoryJointDistribution{}, err
	}
	defer result.close()
	run := func() error {
		var toPointer *C.igraph_vector_int_t
		if to != nil {
			toPointer = &to.value
		}
		code := C.go_igraph_joint_type_distribution(&g.graph, edgeWeightPointer(weights), &result.value, &from.value, toPointer, booltoint(options.Directed), booltoint(options.Normalized))
		if code != C.IGRAPH_SUCCESS {
			return igraphError("calculate categorical joint distribution", int(code))
		}
		return nil
	}
	if hooks.run != nil {
		run = hooks.run
	}
	if err := run(); err != nil {
		return CategoryJointDistribution{}, err
	}
	matrix, err := result.matrix()
	if err != nil {
		return CategoryJointDistribution{}, err
	}
	return CategoryJointDistribution{Matrix: matrix, RowCategories: rowAxis, ColumnCategories: columnAxis}, nil
}

// DegreeJointDistribution computes the ordered endpoint-degree distribution.
// Index i on either axis represents degree i. Options and their referenced
// values are borrowed only for the call; the returned matrix is Go-owned.
//
//igraph:bind igraph_joint_degree_distribution
func (g *Graph) DegreeJointDistribution(options DegreeJointDistributionOptions) (Matrix, error) {
	return g.degreeJointDistribution(options, jointDistributionHooks{})
}

func (g *Graph) degreeJointDistribution(options DegreeJointDistributionOptions, hooks jointDistributionHooks) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}
	fromMode, err := options.FromMode.cValue()
	if err != nil {
		return Matrix{}, err
	}
	toMode, err := options.ToMode.cValue()
	if err != nil {
		return Matrix{}, err
	}
	edges := int(C.igraph_ecount(&g.graph))
	if C.go_igraph_is_directed(&g.graph) != 0 && !options.DirectedNeighbors &&
		(options.FromMode != options.ToMode || !equalDegreeLimits(options.MaximumFromDegree, options.MaximumToDegree)) {
		return Matrix{}, fmt.Errorf("igraph: reciprocal degree mixing requires identical modes and maximums")
	}
	maxFrom, err := degreeLimit(options.MaximumFromDegree)
	if err != nil {
		return Matrix{}, err
	}
	maxTo, err := degreeLimit(options.MaximumToDegree)
	if err != nil {
		return Matrix{}, err
	}
	if err := validateDegreeAllocation(maxFrom, maxTo, edges, true); err != nil {
		return Matrix{}, err
	}
	weights, err := newJointWeights(options.Weights, edges, options.Normalized)
	if err != nil {
		return Matrix{}, err
	}
	if weights != nil {
		defer weights.close()
	}
	return runJointMatrix(hooks, func(matrix *cMatrix) error {
		code := C.go_igraph_joint_degree_distribution(&g.graph, edgeWeightPointer(weights), &matrix.value, fromMode, toMode, booltoint(options.DirectedNeighbors), booltoint(options.Normalized), maxFrom, maxTo)
		if code != C.IGRAPH_SUCCESS {
			return igraphError("calculate joint degree distribution", int(code))
		}
		return nil
	})
}

// JointDegreeMatrix counts each edge once by its endpoint degrees. Matrix
// index zero represents degree one. Options and their referenced values are
// borrowed only for the call; the returned matrix is Go-owned.
//
//igraph:bind igraph_joint_degree_matrix
func (g *Graph) JointDegreeMatrix(options JointDegreeMatrixOptions) (Matrix, error) {
	return g.jointDegreeMatrix(options, jointDistributionHooks{})
}

func (g *Graph) jointDegreeMatrix(options JointDegreeMatrixOptions, hooks jointDistributionHooks) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}
	edges := int(C.igraph_ecount(&g.graph))
	maxOut, err := degreeLimit(options.MaximumOutDegree)
	if err != nil {
		return Matrix{}, err
	}
	maxIn, err := degreeLimit(options.MaximumInDegree)
	if err != nil {
		return Matrix{}, err
	}
	if err := validateDegreeAllocation(maxOut, maxIn, edges, false); err != nil {
		return Matrix{}, err
	}
	weights, err := newJointWeights(options.Weights, edges, false)
	if err != nil {
		return Matrix{}, err
	}
	if weights != nil {
		defer weights.close()
	}
	return runJointMatrix(hooks, func(matrix *cMatrix) error {
		code := C.go_igraph_joint_degree_matrix(&g.graph, edgeWeightPointer(weights), &matrix.value, maxOut, maxIn)
		if code != C.IGRAPH_SUCCESS {
			return igraphError("calculate joint degree matrix", int(code))
		}
		return nil
	})
}

func runJointMatrix(hooks jointDistributionHooks, operation func(*cMatrix) error) (Matrix, error) {
	newMatrix := hooks.newMatrix
	if newMatrix == nil {
		newMatrix = newCMatrix
	}
	result, err := newMatrix(Matrix{})
	if err != nil {
		return Matrix{}, err
	}
	defer result.close()
	run := func() error { return operation(result) }
	if hooks.run != nil {
		run = hooks.run
	}
	if err := run(); err != nil {
		return Matrix{}, err
	}
	return result.matrix()
}

func newJointWeights(values []float64, edgeCount int, normalized bool) (*realVector, error) {
	weights, err := newOptionalEdgeWeights(values, edgeCount)
	if err != nil {
		return nil, err
	}
	if normalized && edgeCount > 0 && values != nil {
		total := 0.0
		for index, value := range values {
			if value < 0 {
				if weights != nil {
					weights.close()
				}
				return nil, fmt.Errorf("igraph: normalized weight at index %d must be non-negative: %v", index, value)
			}
			total += value
		}
		if total <= 0 {
			if weights != nil {
				weights.close()
			}
			return nil, fmt.Errorf("igraph: normalized weights must have a positive total")
		}
	}
	return weights, nil
}

func degreeLimit(limit *int) (C.igraph_int_t, error) {
	if limit == nil {
		return C.igraph_int_t(-1), nil
	}
	if *limit < 0 {
		return 0, fmt.Errorf("igraph: maximum degree must be non-negative: %d", *limit)
	}
	return intToIgraphInt(*limit, "maximum degree")
}

func equalDegreeLimits(first, second *int) bool {
	if first == nil || second == nil {
		return first == nil && second == nil
	}
	return *first == *second
}

func validateDegreeAllocation(from, to C.igraph_int_t, edgeCount int, includeZero bool) error {
	upper := edgeCount
	if upper > int(^uint(0)>>1)/2 {
		return fmt.Errorf("igraph: degree bound is too large")
	}
	upper *= 2
	rows, columns := int(from), int(to)
	if rows < 0 {
		rows = upper
	}
	if columns < 0 {
		columns = upper
	}
	if includeZero {
		rows++
		columns++
	}
	_, err := matrixSize(rows, columns)
	return err
}

func (categories CategoryValues) compact(vertexCount int) ([]int, error) {
	result, _, err := categories.compactWithAxis(vertexCount)
	return result, err
}

func (categories CategoryValues) compactWithAxis(vertexCount int) ([]int, CategoryValues, error) {
	var length int
	switch categories.kind {
	case categoryIntegers:
		length = len(categories.integers)
	case categoryStrings:
		length = len(categories.strings)
	default:
		return nil, CategoryValues{}, fmt.Errorf("igraph: invalid category value kind: %d", categories.kind)
	}
	if length != vertexCount {
		return nil, CategoryValues{}, fmt.Errorf("igraph: category count %d does not match vertex count %d", length, vertexCount)
	}
	result := make([]int, length)
	axis := CategoryValues{kind: categories.kind}
	if categories.kind == categoryIntegers {
		indices := make(map[int]int, length)
		for index, value := range categories.integers {
			compact, exists := indices[value]
			if !exists {
				compact = len(indices)
				indices[value] = compact
				axis.integers = append(axis.integers, value)
			}
			result[index] = compact
		}
	} else {
		indices := make(map[string]int, length)
		for index, value := range categories.strings {
			compact, exists := indices[value]
			if !exists {
				compact = len(indices)
				indices[value] = compact
				axis.strings = append(axis.strings, value)
			}
			result[index] = compact
		}
	}
	return result, axis, nil
}

func categoryCount(categories CategoryValues) int {
	if categories.kind == categoryStrings {
		return len(categories.strings)
	}
	return len(categories.integers)
}

func validateFiniteVertexValues(values []float64, vertexCount int, kind string) error {
	if len(values) != vertexCount {
		return fmt.Errorf("igraph: %s value count %d does not match vertex count %d", kind, len(values), vertexCount)
	}
	for index, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("igraph: %s value at index %d must be finite: %v", kind, index, value)
		}
	}
	return nil
}

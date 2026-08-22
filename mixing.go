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

func (categories CategoryValues) compact(vertexCount int) ([]int, error) {
	var length int
	switch categories.kind {
	case categoryIntegers:
		length = len(categories.integers)
	case categoryStrings:
		length = len(categories.strings)
	default:
		return nil, fmt.Errorf("igraph: invalid category value kind: %d", categories.kind)
	}
	if length != vertexCount {
		return nil, fmt.Errorf("igraph: category count %d does not match vertex count %d", length, vertexCount)
	}
	result := make([]int, length)
	if categories.kind == categoryIntegers {
		indices := make(map[int]int, length)
		for index, value := range categories.integers {
			compact, exists := indices[value]
			if !exists {
				compact = len(indices)
				indices[value] = compact
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
			}
			result[index] = compact
		}
	}
	return result, nil
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

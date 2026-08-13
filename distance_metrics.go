package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
import "C"

import (
	"fmt"
	"math"
)

// PathLengthHistogram is a Go-owned unweighted shortest-path histogram.
// Counts[i] is the number of ordered (directed) or unordered (undirected)
// vertex pairs at distance i+1. Unreachable is the number of unreachable pairs.
type PathLengthHistogram struct {
	Counts      []float64
	Unreachable float64
}

// PseudoDiameterOptions controls approximate diameter calculation.
type PseudoDiameterOptions struct {
	// Start optionally selects the initial vertex. Nil lets C/igraph choose one
	// using its RNG; Seed makes that choice reproducible.
	Start        *int
	Seed         *uint64
	Directed     bool
	Disconnected bool
	Weights      []float64
}

// PseudoDiameterResult contains a lower-bound estimate and its endpoint IDs.
// For an empty graph, Diameter is NaN and From and To are -1.
type PseudoDiameterResult struct {
	Diameter float64
	From     int
	To       int
}

// CutoffDistances is like Distances but ignores paths longer than cutoff.
// cutoff must be finite and non-negative; ignored or unreachable entries are
// positive infinity. Selector order and duplicates are preserved.
//
//igraph:bind igraph_distances_cutoff
func (g *Graph) CutoffDistances(sources, targets VertexSelector, cutoff float64, options PathOptions) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}
	if math.IsNaN(cutoff) || math.IsInf(cutoff, 0) || cutoff < 0 {
		return Matrix{}, fmt.Errorf("igraph: distance cutoff must be finite and non-negative: %v", cutoff)
	}
	mode, err := options.Direction.cValue()
	if err != nil {
		return Matrix{}, err
	}
	vc := int(C.igraph_vcount(&g.graph))
	if err := validateVertexSelector(sources, vc); err != nil {
		return Matrix{}, fmt.Errorf("igraph: invalid source selector: %w", err)
	}
	if err := validateVertexSelector(targets, vc); err != nil {
		return Matrix{}, fmt.Errorf("igraph: invalid target selector: %w", err)
	}
	targetIDs, err := materializeVertexIDs(&g.graph, targets)
	if err != nil {
		return Matrix{}, err
	}
	uniqueTargets, columns := deduplicateIDs(targetIDs)
	uniqueSelector, err := VertexIDs(uniqueTargets...)
	if err != nil {
		return Matrix{}, err
	}
	cs, err := newCVertexSelector(sources)
	if err != nil {
		return Matrix{}, err
	}
	defer cs.close()
	ct, err := newCVertexSelector(uniqueSelector)
	if err != nil {
		return Matrix{}, err
	}
	defer ct.close()
	weights, err := nonNegativeEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)), "cutoff distance")
	if err != nil {
		return Matrix{}, err
	}
	if weights != nil {
		defer weights.close()
	}
	result, err := newCMatrix(Matrix{})
	if err != nil {
		return Matrix{}, err
	}
	defer result.close()
	code := C.go_igraph_distances_cutoff(&g.graph, edgeWeightPointer(weights), &result.value, cs.value, ct.value, mode, C.igraph_real_t(cutoff))
	if code != C.IGRAPH_SUCCESS {
		return Matrix{}, igraphError("calculate cutoff distances", int(code))
	}
	matrix, err := result.matrix()
	if err != nil {
		return Matrix{}, err
	}
	if len(uniqueTargets) == len(targetIDs) {
		return matrix, nil
	}
	return expandDistanceColumns(matrix, columns)
}

// Eccentricities returns selector-aligned maximum finite distances. In a
// disconnected graph, each value is calculated within the vertex component.
//
//igraph:bind igraph_eccentricity
func (g *Graph) Eccentricities(vertices VertexSelector, options PathOptions) ([]float64, error) {
	return g.selectorDistanceValues(vertices, options, "calculate eccentricities", func(graph *C.igraph_t, weights *C.igraph_vector_t, result *C.igraph_vector_t, selector C.igraph_vs_t, mode C.igraph_neimode_t) C.igraph_error_t {
		return C.go_igraph_eccentricity(graph, weights, result, selector, mode)
	})
}

// Radius returns the minimum eccentricity. The empty graph returns NaN.
//
//igraph:bind igraph_radius
func (g *Graph) Radius(options PathOptions) (float64, error) {
	return g.scalarDistanceMetric(options, "calculate radius", func(graph *C.igraph_t, weights *C.igraph_vector_t, result *C.igraph_real_t, mode C.igraph_neimode_t) C.igraph_error_t {
		return C.go_igraph_radius(graph, weights, result, mode)
	})
}

// GraphCenter returns all vertices attaining Radius. This binds an experimental
// upstream API in pinned igraph 1.0.1. The result is non-nil and Go-owned.
//
//igraph:bind igraph_graph_center
func (g *Graph) GraphCenter(options PathOptions) ([]int, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	mode, err := options.Direction.cValue()
	if err != nil {
		return nil, err
	}
	weights, err := nonNegativeEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)), "graph center")
	if err != nil {
		return nil, err
	}
	if weights != nil {
		defer weights.close()
	}
	result, err := newIntVector(nil)
	if err != nil {
		return nil, err
	}
	defer result.close()
	code := C.go_igraph_graph_center(&g.graph, edgeWeightPointer(weights), &result.value, mode)
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("calculate graph center", int(code))
	}
	return result.slice()
}

// PseudoDiameter estimates the diameter and returns a pair of vertices that
// realizes the estimate. The result is a lower bound, not an exact diameter.
// Options and weights are borrowed for the duration of the call.
//
//igraph:bind igraph_pseudo_diameter
func (g *Graph) PseudoDiameter(options PseudoDiameterOptions) (PseudoDiameterResult, error) {
	if g == nil {
		return PseudoDiameterResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return PseudoDiameterResult{}, ErrClosed
	}

	vertexCount := int(C.igraph_vcount(&g.graph))
	start := -1
	if options.Start != nil {
		start = *options.Start
		if start < 0 || start >= vertexCount {
			return PseudoDiameterResult{}, fmt.Errorf("igraph: pseudo-diameter start vertex out of range: %d", start)
		}
	}
	weights, err := nonNegativeEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)), "pseudo-diameter")
	if err != nil {
		return PseudoDiameterResult{}, err
	}
	if weights != nil {
		defer weights.close()
	}

	var diameter C.igraph_real_t
	var from, to C.igraph_int_t
	err = withRNG(options.Seed, func() error {
		code := C.go_igraph_pseudo_diameter(
			&g.graph, edgeWeightPointer(weights), &diameter, C.igraph_int_t(start),
			&from, &to, booltoint(options.Directed), booltoint(options.Disconnected),
		)
		if code != C.IGRAPH_SUCCESS {
			return igraphError("calculate pseudo-diameter", int(code))
		}
		return nil
	})
	if err != nil {
		return PseudoDiameterResult{}, err
	}
	return PseudoDiameterResult{Diameter: float64(diameter), From: int(from), To: int(to)}, nil
}

// GlobalEfficiency returns the mean reciprocal distance between vertex pairs.
//
//igraph:bind igraph_global_efficiency
func (g *Graph) GlobalEfficiency(directed bool, weights []float64) (float64, error) {
	return g.efficiencyScalar(directed, DirectionOut, weights, false)
}

// LocalEfficiencies returns selector-aligned local efficiencies.
//
//igraph:bind igraph_local_efficiency
func (g *Graph) LocalEfficiencies(vertices VertexSelector, directed bool, direction DirectionMode, weights []float64) ([]float64, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	mode, err := direction.cValue()
	if err != nil {
		return nil, err
	}
	vc := int(C.igraph_vcount(&g.graph))
	if err := validateVertexSelector(vertices, vc); err != nil {
		return nil, err
	}
	selector, err := newCVertexSelector(vertices)
	if err != nil {
		return nil, err
	}
	defer selector.close()
	cweights, err := nonNegativeEdgeWeights(weights, int(C.igraph_ecount(&g.graph)), "local efficiency")
	if err != nil {
		return nil, err
	}
	if cweights != nil {
		defer cweights.close()
	}
	result, err := newRealVector(nil)
	if err != nil {
		return nil, err
	}
	defer result.close()
	code := C.go_igraph_local_efficiency(&g.graph, edgeWeightPointer(cweights), &result.value, selector.value, booltoint(directed), mode)
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("calculate local efficiencies", int(code))
	}
	return result.slice()
}

// AverageLocalEfficiency returns the mean local efficiency.
//
//igraph:bind igraph_average_local_efficiency
func (g *Graph) AverageLocalEfficiency(directed bool, direction DirectionMode, weights []float64) (float64, error) {
	return g.efficiencyScalar(directed, direction, weights, true)
}

// PathLengthHistogram returns the unweighted shortest-path length distribution.
//
//igraph:bind igraph_path_length_hist
func (g *Graph) PathLengthHistogram(directed bool) (PathLengthHistogram, error) {
	if g == nil {
		return PathLengthHistogram{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return PathLengthHistogram{}, ErrClosed
	}
	result, err := newRealVector(nil)
	if err != nil {
		return PathLengthHistogram{}, err
	}
	defer result.close()
	var unreachable C.igraph_real_t
	code := C.go_igraph_path_length_hist(&g.graph, &result.value, &unreachable, booltoint(directed))
	if code != C.IGRAPH_SUCCESS {
		return PathLengthHistogram{}, igraphError("calculate path length histogram", int(code))
	}
	counts, err := result.slice()
	if err != nil {
		return PathLengthHistogram{}, err
	}
	return PathLengthHistogram{Counts: counts, Unreachable: float64(unreachable)}, nil
}

type selectorMetricCall func(*C.igraph_t, *C.igraph_vector_t, *C.igraph_vector_t, C.igraph_vs_t, C.igraph_neimode_t) C.igraph_error_t

func (g *Graph) selectorDistanceValues(vertices VertexSelector, options PathOptions, operation string, call selectorMetricCall) ([]float64, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	mode, err := options.Direction.cValue()
	if err != nil {
		return nil, err
	}
	if err := validateVertexSelector(vertices, int(C.igraph_vcount(&g.graph))); err != nil {
		return nil, err
	}
	selector, err := newCVertexSelector(vertices)
	if err != nil {
		return nil, err
	}
	defer selector.close()
	weights, err := nonNegativeEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)), operation)
	if err != nil {
		return nil, err
	}
	if weights != nil {
		defer weights.close()
	}
	result, err := newRealVector(nil)
	if err != nil {
		return nil, err
	}
	defer result.close()
	if code := call(&g.graph, edgeWeightPointer(weights), &result.value, selector.value, mode); code != C.IGRAPH_SUCCESS {
		return nil, igraphError(operation, int(code))
	}
	return result.slice()
}

type scalarMetricCall func(*C.igraph_t, *C.igraph_vector_t, *C.igraph_real_t, C.igraph_neimode_t) C.igraph_error_t

func (g *Graph) scalarDistanceMetric(options PathOptions, operation string, call scalarMetricCall) (float64, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return 0, ErrClosed
	}
	mode, err := options.Direction.cValue()
	if err != nil {
		return 0, err
	}
	weights, err := nonNegativeEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)), operation)
	if err != nil {
		return 0, err
	}
	if weights != nil {
		defer weights.close()
	}
	var result C.igraph_real_t
	if code := call(&g.graph, edgeWeightPointer(weights), &result, mode); code != C.IGRAPH_SUCCESS {
		return 0, igraphError(operation, int(code))
	}
	return float64(result), nil
}

func (g *Graph) efficiencyScalar(directed bool, direction DirectionMode, weights []float64, local bool) (float64, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return 0, ErrClosed
	}
	mode, err := direction.cValue()
	if err != nil {
		return 0, err
	}
	cweights, err := nonNegativeEdgeWeights(weights, int(C.igraph_ecount(&g.graph)), "efficiency")
	if err != nil {
		return 0, err
	}
	if cweights != nil {
		defer cweights.close()
	}
	var result C.igraph_real_t
	var code C.igraph_error_t
	if local {
		code = C.go_igraph_average_local_efficiency(&g.graph, edgeWeightPointer(cweights), &result, booltoint(directed), mode)
	} else {
		code = C.go_igraph_global_efficiency(&g.graph, edgeWeightPointer(cweights), &result, booltoint(directed))
	}
	if code != C.IGRAPH_SUCCESS {
		return 0, igraphError("calculate efficiency", int(code))
	}
	return float64(result), nil
}

func nonNegativeEdgeWeights(values []float64, edgeCount int, operation string) (*realVector, error) {
	for i, value := range values {
		if value < 0 {
			return nil, fmt.Errorf("igraph: %s weight at index %d must be non-negative: %v", operation, i, value)
		}
	}
	return newOptionalEdgeWeights(values, edgeCount)
}

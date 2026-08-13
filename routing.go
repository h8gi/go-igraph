package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
// #include "subgraph_cgo.h"
import "C"

import (
	"fmt"
	"math"
)

type VoronoiTieBreaker uint8

const (
	VoronoiFirst VoronoiTieBreaker = iota
	VoronoiLast
	VoronoiRandom
)

func (tie VoronoiTieBreaker) cValue() (C.igraph_voronoi_tiebreaker_t, error) {
	switch tie {
	case VoronoiFirst:
		return C.IGRAPH_VORONOI_FIRST, nil
	case VoronoiLast:
		return C.IGRAPH_VORONOI_LAST, nil
	case VoronoiRandom:
		return C.IGRAPH_VORONOI_RANDOM, nil
	default:
		return 0, fmt.Errorf("igraph: invalid Voronoi tie breaker: %d", tie)
	}
}

type VoronoiOptions struct {
	Direction  DirectionMode
	Weights    []float64
	TieBreaker VoronoiTieBreaker
	Seed       *uint64
}

type VoronoiResult struct {
	Membership []int
	Distances  []float64
}

type SpannerOptions struct {
	Stretch float64
	Weights []float64
	Seed    *uint64
}

type SpannerResult struct {
	Graph       *Graph
	SourceEdges []int
}

// WidestPath returns one maximum-bottleneck path. Weights are required,
// borrowed for the call, and may be any finite values.
func (g *Graph) WidestPath(source, target int, weights []float64, direction DirectionMode) (Path, error) {
	targets, err := VertexIDs(target)
	if err != nil {
		return Path{}, err
	}
	paths, err := g.WidestPaths(source, targets, weights, direction)
	if err != nil {
		return Path{}, err
	}
	return paths[0], nil
}

// WidestPaths returns selector-ordered maximum-bottleneck paths, preserving
// duplicate targets. Every returned Path is Go-owned.
//
//igraph:bind igraph_get_widest_paths
func (g *Graph) WidestPaths(source int, targets VertexSelector, weights []float64, direction DirectionMode) ([]Path, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	vc := int(C.igraph_vcount(&g.graph))
	if err := validateVertexID(source, vc); err != nil {
		return nil, err
	}
	if err := validateVertexSelector(targets, vc); err != nil {
		return nil, err
	}
	mode, err := direction.cValue()
	if err != nil {
		return nil, err
	}
	targetIDs, err := materializeVertexIDs(&g.graph, targets)
	if err != nil {
		return nil, err
	}
	selector, err := newCVertexSelector(targets)
	if err != nil {
		return nil, err
	}
	defer selector.close()
	cweights, err := requiredFiniteEdgeWeights(weights, int(C.igraph_ecount(&g.graph)), "widest path")
	if err != nil {
		return nil, err
	}
	defer cweights.close()
	vertices, edges, err := newPathVectorLists()
	if err != nil {
		return nil, err
	}
	defer vertices.close()
	defer edges.close()
	code := C.go_igraph_get_widest_paths(&g.graph, &vertices.value, &edges.value, C.igraph_int_t(source), selector.value, &cweights.value, mode)
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("calculate widest paths", int(code))
	}
	result, err := convertPathLists(vertices, edges, true)
	if err != nil {
		return nil, err
	}
	if len(result) != len(targetIDs) {
		return nil, fmt.Errorf("igraph: widest path result length mismatch")
	}
	return result, nil
}

// WidestPathWidths returns selector-aligned maximum bottleneck widths.
//
//igraph:bind igraph_widest_path_widths_dijkstra
func (g *Graph) WidestPathWidths(sources, targets VertexSelector, weights []float64, direction DirectionMode) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}
	vc := int(C.igraph_vcount(&g.graph))
	if err := validateVertexSelector(sources, vc); err != nil {
		return Matrix{}, err
	}
	if err := validateVertexSelector(targets, vc); err != nil {
		return Matrix{}, err
	}
	targetIDs, err := materializeVertexIDs(&g.graph, targets)
	if err != nil {
		return Matrix{}, err
	}
	uniqueTargets, columns := deduplicateIDs(targetIDs)
	targetSelector, err := VertexIDs(uniqueTargets...)
	if err != nil {
		return Matrix{}, err
	}
	mode, err := direction.cValue()
	if err != nil {
		return Matrix{}, err
	}
	cs, err := newCVertexSelector(sources)
	if err != nil {
		return Matrix{}, err
	}
	defer cs.close()
	ct, err := newCVertexSelector(targetSelector)
	if err != nil {
		return Matrix{}, err
	}
	defer ct.close()
	cweights, err := requiredFiniteEdgeWeights(weights, int(C.igraph_ecount(&g.graph)), "widest path width")
	if err != nil {
		return Matrix{}, err
	}
	defer cweights.close()
	result, err := newCMatrix(Matrix{})
	if err != nil {
		return Matrix{}, err
	}
	defer result.close()
	if code := C.go_igraph_widest_path_widths(&g.graph, &result.value, cs.value, ct.value, &cweights.value, mode); code != C.IGRAPH_SUCCESS {
		return Matrix{}, igraphError("calculate widest path widths", int(code))
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

// Voronoi partitions vertices by nearest generator. Membership entries index
// the generator input; unreachable vertices have membership -1 and +Inf distance.
//
//igraph:bind igraph_voronoi
func (g *Graph) Voronoi(generators []int, options VoronoiOptions) (VoronoiResult, error) {
	if g == nil {
		return VoronoiResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return VoronoiResult{}, ErrClosed
	}
	mode, err := options.Direction.cValue()
	if err != nil {
		return VoronoiResult{}, err
	}
	tie, err := options.TieBreaker.cValue()
	if err != nil {
		return VoronoiResult{}, err
	}
	vc := int(C.igraph_vcount(&g.graph))
	for _, generator := range generators {
		if err := validateVertexID(generator, vc); err != nil {
			return VoronoiResult{}, err
		}
	}
	cgenerators, err := newIntVector(generators)
	if err != nil {
		return VoronoiResult{}, err
	}
	defer cgenerators.close()
	weights, err := nonNegativeEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)), "Voronoi")
	if err != nil {
		return VoronoiResult{}, err
	}
	if weights != nil {
		defer weights.close()
	}
	membership, err := newIntVector(nil)
	if err != nil {
		return VoronoiResult{}, err
	}
	defer membership.close()
	distances, err := newRealVector(nil)
	if err != nil {
		return VoronoiResult{}, err
	}
	defer distances.close()
	err = withRNG(options.Seed, func() error {
		if code := C.go_igraph_voronoi(&g.graph, &membership.value, &distances.value, &cgenerators.value, edgeWeightPointer(weights), mode, tie); code != C.IGRAPH_SUCCESS {
			return igraphError("calculate Voronoi partition", int(code))
		}
		return nil
	})
	if err != nil {
		return VoronoiResult{}, err
	}
	members, err := membership.slice()
	if err != nil {
		return VoronoiResult{}, err
	}
	dists, err := distances.slice()
	if err != nil {
		return VoronoiResult{}, err
	}
	return VoronoiResult{Membership: members, Distances: dists}, nil
}

// Spanner returns an independently owned edge-induced spanner and its source
// edge IDs. The caller must close Graph.
//
//igraph:bind igraph_spanner
func (g *Graph) Spanner(options SpannerOptions) (SpannerResult, error) {
	if g == nil {
		return SpannerResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return SpannerResult{}, ErrClosed
	}
	if math.IsNaN(options.Stretch) || math.IsInf(options.Stretch, 0) || options.Stretch < 1 {
		return SpannerResult{}, fmt.Errorf("igraph: spanner stretch must be finite and at least 1: %v", options.Stretch)
	}
	weights, err := nonNegativeEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)), "spanner")
	if err != nil {
		return SpannerResult{}, err
	}
	if weights != nil {
		defer weights.close()
	}
	edges, err := newIntVector(nil)
	if err != nil {
		return SpannerResult{}, err
	}
	defer edges.close()
	err = withRNG(options.Seed, func() error {
		if code := C.go_igraph_spanner(&g.graph, &edges.value, C.igraph_real_t(options.Stretch), edgeWeightPointer(weights)); code != C.IGRAPH_SUCCESS {
			return igraphError("calculate spanner", int(code))
		}
		return nil
	})
	if err != nil {
		return SpannerResult{}, err
	}
	edgeIDs, err := edges.slice()
	if err != nil {
		return SpannerResult{}, err
	}
	edgeIDs = sortedUniqueIDs(edgeIDs)
	selectorValue, err := EdgeIDs(edgeIDs...)
	if err != nil {
		return SpannerResult{}, err
	}
	selector, err := newCEdgeSelector(selectorValue)
	if err != nil {
		return SpannerResult{}, err
	}
	defer selector.close()
	var graph C.igraph_t
	if code := C.go_igraph_subgraph_from_edges(&g.graph, &graph, selector.value, booltoint(false)); code != C.IGRAPH_SUCCESS {
		return SpannerResult{}, igraphError("create spanner graph", int(code))
	}
	return SpannerResult{Graph: adoptInitializedGraph(&graph), SourceEdges: edgeIDs}, nil
}

func requiredFiniteEdgeWeights(values []float64, edgeCount int, operation string) (*realVector, error) {
	if values == nil {
		return nil, fmt.Errorf("igraph: %s weights are required", operation)
	}
	return newOptionalEdgeWeights(values, edgeCount)
}

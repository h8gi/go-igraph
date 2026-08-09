package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
import "C"

import (
	"fmt"
	"math"
)

// DistanceCentralityOptions controls closeness and harmonic centrality.
// Direction is ignored for undirected graphs. Normalized divides closeness by
// the number of reachable other vertices and harmonic centrality by the total
// number of other graph vertices. A nil Cutoff includes paths of every length;
// a non-nil Cutoff includes only paths whose length is at most that finite,
// non-negative value.
//
// A nil Weights slice requests an unweighted calculation. A non-nil slice is
// borrowed only for the call, copied into temporary C storage, and must contain
// one finite, strictly positive value per edge. Cutoff is likewise borrowed
// only for the call.
type DistanceCentralityOptions struct {
	Direction  DirectionMode
	Weights    []float64
	Normalized bool
	Cutoff     *float64
}

// ClosenessResult contains values in materialized selector order, including
// duplicates. Scores and ReachableCounts are non-nil, Go-owned slices that
// remain valid after the source graph is closed. ReachableCounts counts
// reachable other vertices and excludes each selected vertex itself.
// AllReachable reports whether every selected vertex can reach every other
// graph vertex under the requested direction and cutoff.
type ClosenessResult struct {
	Scores          []float64
	ReachableCounts []int
	AllReachable    bool
}

// Closeness calculates the reciprocal sum of distances from each selected
// vertex. With Normalized set, it multiplies that value by the number of
// reachable other vertices, yielding the reciprocal mean distance. An isolated
// vertex has a NaN score because it reaches no other vertex. With a finite
// cutoff, only paths no longer than the cutoff contribute.
//
//igraph:bind igraph_closeness
//igraph:bind igraph_closeness_cutoff
func (g *Graph) Closeness(vertices VertexSelector, options DistanceCentralityOptions) (ClosenessResult, error) {
	if g == nil {
		return ClosenessResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return ClosenessResult{}, ErrClosed
	}

	mode, weights, cutoff, hasCutoff, selectedIDs, uniqueVertices, positions, err :=
		g.prepareDistanceCentrality(vertices, options)
	if err != nil {
		return ClosenessResult{}, err
	}
	if weights != nil {
		defer weights.close()
	}
	selector, err := newCVertexSelector(uniqueVertices)
	if err != nil {
		return ClosenessResult{}, err
	}
	defer selector.close()
	scores, err := newRealVectorSize(0)
	if err != nil {
		return ClosenessResult{}, err
	}
	defer scores.close()
	reachable, err := newIntVector(nil)
	if err != nil {
		return ClosenessResult{}, err
	}
	defer reachable.close()

	var allReachable C.igraph_bool_t
	code := C.go_igraph_closeness(
		&g.graph, &scores.value, &reachable.value, &allReachable,
		selector.value, mode, edgeWeightPointer(weights),
		C.igraph_bool_t(booltoint(options.Normalized)),
		C.igraph_bool_t(booltoint(hasCutoff)), C.igraph_real_t(cutoff),
	)
	if code != C.IGRAPH_SUCCESS {
		return ClosenessResult{}, igraphError("calculate closeness centrality", int(code))
	}
	goScores, err := scores.slice()
	if err != nil {
		return ClosenessResult{}, err
	}
	goReachable, err := reachable.slice()
	if err != nil {
		return ClosenessResult{}, err
	}
	if len(selectedIDs) != len(goScores) {
		goScores = expandByPositions(goScores, positions)
		goReachable = expandByPositions(goReachable, positions)
	}
	return ClosenessResult{
		Scores:          goScores,
		ReachableCounts: goReachable,
		AllReachable:    allReachable != booltoint(false),
	}, nil
}

// HarmonicCentrality calculates the sum of reciprocal distances from each
// selected vertex. Unreachable vertices contribute zero. The result follows
// materialized selector order, including duplicates, and is a non-nil,
// Go-owned slice that remains valid after the graph is closed.
//
//igraph:bind igraph_harmonic_centrality
//igraph:bind igraph_harmonic_centrality_cutoff
func (g *Graph) HarmonicCentrality(vertices VertexSelector, options DistanceCentralityOptions) ([]float64, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}

	mode, weights, cutoff, hasCutoff, selectedIDs, uniqueVertices, positions, err :=
		g.prepareDistanceCentrality(vertices, options)
	if err != nil {
		return nil, err
	}
	if weights != nil {
		defer weights.close()
	}
	selector, err := newCVertexSelector(uniqueVertices)
	if err != nil {
		return nil, err
	}
	defer selector.close()
	result, err := newRealVectorSize(0)
	if err != nil {
		return nil, err
	}
	defer result.close()

	code := C.go_igraph_harmonic_centrality(
		&g.graph, &result.value, selector.value, mode,
		edgeWeightPointer(weights), C.igraph_bool_t(booltoint(options.Normalized)),
		C.igraph_bool_t(booltoint(hasCutoff)), C.igraph_real_t(cutoff),
	)
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("calculate harmonic centrality", int(code))
	}
	values, err := result.slice()
	if err != nil {
		return nil, err
	}
	if len(selectedIDs) != len(values) {
		values = expandByPositions(values, positions)
	}
	return values, nil
}

func (g *Graph) prepareDistanceCentrality(
	vertices VertexSelector,
	options DistanceCentralityOptions,
) (C.igraph_neimode_t, *realVector, float64, bool, []int, VertexSelector, []int, error) {
	mode, err := options.Direction.cValue()
	if err != nil {
		return 0, nil, 0, false, nil, VertexSelector{}, nil, err
	}
	vertexCount := int(C.igraph_vcount(&g.graph))
	if err := validateVertexSelector(vertices, vertexCount); err != nil {
		return 0, nil, 0, false, nil, VertexSelector{}, nil,
			fmt.Errorf("igraph: invalid centrality selector: %w", err)
	}
	selectedIDs, err := materializeVertexIDs(&g.graph, vertices)
	if err != nil {
		return 0, nil, 0, false, nil, VertexSelector{}, nil,
			fmt.Errorf("igraph: materialize centrality selector: %w", err)
	}
	uniqueIDs, positions := deduplicateVertexIDs(selectedIDs)
	uniqueVertices := vertices
	if len(uniqueIDs) != len(selectedIDs) {
		uniqueVertices, err = VertexIDs(uniqueIDs...)
		if err != nil {
			return 0, nil, 0, false, nil, VertexSelector{}, nil, err
		}
	}
	cutoff, hasCutoff, err := validateCentralityCutoff(options.Cutoff)
	if err != nil {
		return 0, nil, 0, false, nil, VertexSelector{}, nil, err
	}
	weights, err := newOptionalPositiveEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return 0, nil, 0, false, nil, VertexSelector{}, nil, err
	}
	return mode, weights, cutoff, hasCutoff, selectedIDs, uniqueVertices, positions, nil
}

func validateCentralityCutoff(cutoff *float64) (float64, bool, error) {
	if cutoff == nil {
		return 0, false, nil
	}
	if math.IsNaN(*cutoff) || math.IsInf(*cutoff, 0) || *cutoff < 0 {
		return 0, false, fmt.Errorf("igraph: centrality cutoff must be finite and non-negative: %v", *cutoff)
	}
	return *cutoff, true, nil
}

func newOptionalPositiveEdgeWeights(values []float64, edgeCount int) (*realVector, error) {
	if values != nil {
		for index, value := range values {
			if value <= 0 {
				return nil, fmt.Errorf("igraph: weight at index %d must be positive: %v", index, value)
			}
		}
	}
	return newOptionalEdgeWeights(values, edgeCount)
}

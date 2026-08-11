package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "cycle_cgo.h"
import "C"

import (
	"fmt"
	"math"
)

// FeedbackEdgeStrategy selects exact optimization or the directed Eades
// heuristic. Its zero value is FeedbackEdgeExact. The current exact IP backend
// remains an upstream implementation detail.
type FeedbackEdgeStrategy uint8

const (
	// FeedbackEdgeExact minimizes total non-negative edge weight. With nil
	// weights this is minimum cardinality.
	FeedbackEdgeExact FeedbackEdgeStrategy = iota
	// FeedbackEdgeApproximateEades returns a valid directed feedback edge set
	// using a linear-time heuristic; it is not promised to be minimum.
	FeedbackEdgeApproximateEades
)

func (strategy FeedbackEdgeStrategy) cValue() (C.igraph_fas_algorithm_t, error) {
	switch strategy {
	case FeedbackEdgeExact:
		return C.IGRAPH_FAS_EXACT_IP, nil
	case FeedbackEdgeApproximateEades:
		return C.IGRAPH_FAS_APPROX_EADES, nil
	default:
		return 0, fmt.Errorf("igraph: invalid feedback edge strategy: %d", strategy)
	}
}

// FeedbackEdgeOptions controls feedback edge-set computation. Weights is
// borrowed only for the synchronous call and copied into temporary C storage.
// Nil means unit weights; a non-nil slice must contain one finite non-negative
// value per edge. Zero is valid.
type FeedbackEdgeOptions struct {
	Strategy FeedbackEdgeStrategy
	Weights  []float64
}

// FeedbackEdgeSet returns edge IDs whose removal makes the graph acyclic.
// Exact mode minimizes total non-negative weight; zero-weight ties may return
// different equally optimal valid sets. ApproximateEades is valid but is not
// described as minimum.
//
// For undirected graphs, pinned igraph computes the complement of a maximum
// weight spanning forest regardless of strategy. Self-loops are therefore
// always returned, while at most one edge from each parallel connection can
// remain in the forest. The result is non-nil, Go-owned, and survives Close.
//
//igraph:bind igraph_feedback_arc_set
func (g *Graph) FeedbackEdgeSet(options FeedbackEdgeOptions) ([]int, error) {
	return g.feedbackEdgeSet(options, nil)
}

// FeedbackVertexSet returns a minimum-total-weight set of vertex IDs whose
// removal makes the graph acyclic. Weights follows the same borrowed, copied,
// finite non-negative contract as FeedbackEdgeOptions, except that it must
// contain one value per vertex. Nil means unit weights and zero is valid. The
// sole pinned exact algorithm is used without exposing a meaningless selector.
// The result is non-nil, Go-owned, and survives Close.
//
//igraph:bind igraph_feedback_vertex_set
func (g *Graph) FeedbackVertexSet(weights []float64) ([]int, error) {
	return g.feedbackVertexSet(weights, nil)
}

type feedbackAdapters struct {
	initializeResult  func() (*intVector, error)
	initializeWeights func([]float64, int, string) (*realVector, error)
	closeResult       func(*intVector)
	closeWeights      func(*realVector)
	convert           func(*intVector) ([]int, error)
	edgeCall          func(
		*Graph, *intVector, *realVector, FeedbackEdgeStrategy,
	) int
	vertexCall func(*Graph, *intVector, *realVector) int
}

func defaultFeedbackAdapters() feedbackAdapters {
	return feedbackAdapters{
		initializeResult:  func() (*intVector, error) { return newIntVector(nil) },
		initializeWeights: newFeedbackWeights,
		closeResult:       func(vector *intVector) { vector.close() },
		closeWeights:      func(vector *realVector) { vector.close() },
		convert:           func(vector *intVector) ([]int, error) { return vector.slice() },
		edgeCall: func(g *Graph, result *intVector, weights *realVector, strategy FeedbackEdgeStrategy) int {
			algorithm, _ := strategy.cValue()
			return int(C.go_igraph_feedback_arc_set(
				&g.graph, &result.value, feedbackWeightPointer(weights), algorithm,
			))
		},
		vertexCall: func(g *Graph, result *intVector, weights *realVector) int {
			return int(C.go_igraph_feedback_vertex_set(
				&g.graph, &result.value, feedbackWeightPointer(weights),
			))
		},
	}
}

func newFeedbackWeights(values []float64, count int, kind string) (*realVector, error) {
	if values == nil {
		return nil, nil
	}
	if len(values) != count {
		return nil, fmt.Errorf("igraph: feedback %s weight count %d does not match %s count %d", kind, len(values), kind, count)
	}
	for index, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
			return nil, fmt.Errorf("igraph: feedback %s weight at index %d must be finite and non-negative: %v", kind, index, value)
		}
	}
	return newRealVector(values)
}

func feedbackWeightPointer(weights *realVector) *C.igraph_vector_t {
	if weights == nil {
		return nil
	}
	return &weights.value
}

func (g *Graph) feedbackEdgeSet(options FeedbackEdgeOptions, adapters *feedbackAdapters) ([]int, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	if _, err := options.Strategy.cValue(); err != nil {
		return nil, err
	}
	resolved := defaultFeedbackAdapters()
	if adapters != nil {
		resolved = *adapters
	}
	result, err := resolved.initializeResult()
	if err != nil {
		return nil, err
	}
	defer resolved.closeResult(result)
	weights, err := resolved.initializeWeights(options.Weights, int(C.igraph_ecount(&g.graph)), "edge")
	if err != nil {
		return nil, err
	}
	if weights != nil {
		defer resolved.closeWeights(weights)
	}
	if code := resolved.edgeCall(g, result, weights, options.Strategy); code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("calculate feedback edge set", code)
	}
	return resolved.convert(result)
}

func (g *Graph) feedbackVertexSet(values []float64, adapters *feedbackAdapters) ([]int, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	resolved := defaultFeedbackAdapters()
	if adapters != nil {
		resolved = *adapters
	}
	result, err := resolved.initializeResult()
	if err != nil {
		return nil, err
	}
	defer resolved.closeResult(result)
	weights, err := resolved.initializeWeights(values, int(C.igraph_vcount(&g.graph)), "vertex")
	if err != nil {
		return nil, err
	}
	if weights != nil {
		defer resolved.closeWeights(weights)
	}
	if code := resolved.vertexCall(g, result, weights); code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("calculate feedback vertex set", code)
	}
	return resolved.convert(result)
}

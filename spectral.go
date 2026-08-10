package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
import "C"

import (
	"fmt"
	"math"
)

// SpectralSolverOptions exposes caller-relevant convergence settings without
// exposing or retaining an ARPACK object. Zero values request igraph defaults.
type SpectralSolverOptions struct {
	MaxIterations int
	Tolerance     float64
}

// EigenvectorCentralityOptions controls adjacency-eigenvector centrality.
// Direction selects incoming, outgoing, or direction-ignoring adjacency on a
// directed graph and is ignored on an undirected graph. Non-nil Weights are
// borrowed only for the call, copied into temporary C storage, and must contain
// one finite, non-negative value per edge.
type EigenvectorCentralityOptions struct {
	Direction DirectionMode
	Weights   []float64
	Solver    SpectralSolverOptions
}

// EigenvectorCentralityResult contains max-absolute-value-scaled scores in
// vertex ID order and the leading adjacency eigenvalue. Scores is a non-nil,
// Go-owned slice and remains valid after the source graph is closed. An empty
// graph returns an empty slice and eigenvalue zero. If the weighted adjacency
// matrix is all zero, igraph's warning path yields all-one scores and
// eigenvalue zero; that eigenvector is non-unique.
type EigenvectorCentralityResult struct {
	Scores     []float64
	Eigenvalue float64
}

// EigenvectorCentrality calculates adjacency-eigenvector centrality for every
// vertex. Solver and weight inputs are borrowed only for the call.
//
//igraph:bind igraph_eigenvector_centrality
func (g *Graph) EigenvectorCentrality(options EigenvectorCentralityOptions) (EigenvectorCentralityResult, error) {
	if g == nil {
		return EigenvectorCentralityResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return EigenvectorCentralityResult{}, ErrClosed
	}
	mode, err := options.Direction.cValue()
	if err != nil {
		return EigenvectorCentralityResult{}, err
	}
	maxIterations, tolerance, err := validateSpectralSolver(options.Solver)
	if err != nil {
		return EigenvectorCentralityResult{}, err
	}
	weights, err := newOptionalNonNegativeEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return EigenvectorCentralityResult{}, err
	}
	if weights != nil {
		defer weights.close()
	}
	result, err := newRealVectorSize(0)
	if err != nil {
		return EigenvectorCentralityResult{}, err
	}
	defer result.close()
	var eigenvalue C.igraph_real_t
	code := C.go_igraph_eigenvector_centrality(
		&g.graph, &result.value, &eigenvalue, mode, edgeWeightPointer(weights),
		C.int(maxIterations), C.igraph_real_t(tolerance),
	)
	if code != C.IGRAPH_SUCCESS {
		return EigenvectorCentralityResult{}, igraphError("calculate eigenvector centrality", int(code))
	}
	scores, err := result.slice()
	if err != nil {
		return EigenvectorCentralityResult{}, err
	}
	return EigenvectorCentralityResult{Scores: scores, Eigenvalue: float64(eigenvalue)}, nil
}

// HITSOptions controls hub and authority score calculation. Non-nil weights
// follow the same borrowing, copying, and non-negative-value rules as
// EigenvectorCentralityOptions.
type HITSOptions struct {
	Weights []float64
	Solver  SpectralSolverOptions
}

// HITSResult contains max-absolute-value-scaled hub and authority scores in
// vertex ID order plus their common leading eigenvalue. Both slices are
// non-nil, Go-owned, and remain valid after graph closure. Empty graphs return
// empty slices and eigenvalue zero. For an all-zero weighted adjacency matrix,
// both score slices contain ones and the eigenvalue is zero; the vectors are
// non-unique.
type HITSResult struct {
	HubScores       []float64
	AuthorityScores []float64
	Eigenvalue      float64
}

// HITS calculates hub and authority scores together for every vertex.
//
//igraph:bind igraph_hub_and_authority_scores
func (g *Graph) HITS(options HITSOptions) (HITSResult, error) {
	if g == nil {
		return HITSResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return HITSResult{}, ErrClosed
	}
	maxIterations, tolerance, err := validateSpectralSolver(options.Solver)
	if err != nil {
		return HITSResult{}, err
	}
	weights, err := newOptionalNonNegativeEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return HITSResult{}, err
	}
	if weights != nil {
		defer weights.close()
	}
	hubs, err := newRealVectorSize(0)
	if err != nil {
		return HITSResult{}, err
	}
	defer hubs.close()
	authorities, err := newRealVectorSize(0)
	if err != nil {
		return HITSResult{}, err
	}
	defer authorities.close()
	var eigenvalue C.igraph_real_t
	code := C.go_igraph_hub_and_authority_scores(
		&g.graph, &hubs.value, &authorities.value, &eigenvalue,
		edgeWeightPointer(weights), C.int(maxIterations), C.igraph_real_t(tolerance),
	)
	if code != C.IGRAPH_SUCCESS {
		return HITSResult{}, igraphError("calculate HITS scores", int(code))
	}
	hubScores, err := hubs.slice()
	if err != nil {
		return HITSResult{}, err
	}
	authorityScores, err := authorities.slice()
	if err != nil {
		return HITSResult{}, err
	}
	return HITSResult{
		HubScores: hubScores, AuthorityScores: authorityScores,
		Eigenvalue: float64(eigenvalue),
	}, nil
}

// PageRankDirection controls whether PageRank respects edge directions.
type PageRankDirection uint8

const (
	// PageRankRespectDirections follows source-to-target edges on directed graphs.
	PageRankRespectDirections PageRankDirection = iota
	// PageRankIgnoreDirections treats directed edges as undirected connections.
	PageRankIgnoreDirections
)

func (direction PageRankDirection) directed(graphDirected bool) (bool, error) {
	switch direction {
	case PageRankRespectDirections:
		return graphDirected, nil
	case PageRankIgnoreDirections:
		return false, nil
	default:
		return false, fmt.Errorf("igraph: invalid PageRank direction: %d", direction)
	}
}

// PageRankAlgorithm selects the PageRank backend. The zero value is the
// upstream-recommended PRPACK backend.
type PageRankAlgorithm uint8

const (
	PageRankPRPACK PageRankAlgorithm = iota
	PageRankARPACK
)

func (algorithm PageRankAlgorithm) cValue() (C.igraph_pagerank_algo_t, error) {
	switch algorithm {
	case PageRankPRPACK:
		return C.IGRAPH_PAGERANK_ALGO_PRPACK, nil
	case PageRankARPACK:
		return C.IGRAPH_PAGERANK_ALGO_ARPACK, nil
	default:
		return 0, fmt.Errorf("igraph: invalid PageRank algorithm: %d", algorithm)
	}
}

// PageRankOptions controls PageRank and personalized PageRank. Nil Damping
// uses 0.85; an explicit value must be finite and in [0, 1). Direction is
// ignored for undirected graphs. Solver settings affect only the ARPACK
// backend. Dangling vertices redistribute their probability according to the
// active reset distribution (uniform for standard PageRank).
//
// ResetDistribution and ResetVertices are mutually exclusive. A non-nil reset
// distribution must contain one finite, non-negative value per vertex and have
// a positive sum. A non-nil reset selector must select at least one vertex;
// duplicate IDs have set semantics. All option inputs are borrowed only for the
// call. Non-nil edge weights are copied into temporary C storage and must
// contain one finite, non-negative value per edge.
type PageRankOptions struct {
	Damping           *float64
	Direction         PageRankDirection
	Algorithm         PageRankAlgorithm
	Weights           []float64
	ResetDistribution []float64
	ResetVertices     *VertexSelector
	Solver            SpectralSolverOptions
}

// PageRankResult contains probability scores in materialized selector order,
// including duplicates, and the dominant eigenvalue (normally one). Scores is
// a non-nil, Go-owned slice that remains valid after graph closure.
type PageRankResult struct {
	Scores     []float64
	Eigenvalue float64
}

// PageRank calculates standard or personalized PageRank. Selecting fewer
// return vertices does not reduce the upstream whole-graph calculation cost.
//
//igraph:bind igraph_pagerank
//igraph:bind igraph_personalized_pagerank
//igraph:bind igraph_personalized_pagerank_vs
func (g *Graph) PageRank(vertices VertexSelector, options PageRankOptions) (PageRankResult, error) {
	if g == nil {
		return PageRankResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return PageRankResult{}, ErrClosed
	}
	vertexCount := int(C.igraph_vcount(&g.graph))
	if err := validateVertexSelector(vertices, vertexCount); err != nil {
		return PageRankResult{}, fmt.Errorf("igraph: invalid PageRank result selector: %w", err)
	}
	selectedIDs, err := materializeVertexIDs(&g.graph, vertices)
	if err != nil {
		return PageRankResult{}, fmt.Errorf("igraph: materialize PageRank result selector: %w", err)
	}
	uniqueIDs, positions := deduplicateIDs(selectedIDs)
	uniqueVertices := vertices
	if len(uniqueIDs) != len(selectedIDs) {
		uniqueVertices, err = VertexIDs(uniqueIDs...)
		if err != nil {
			return PageRankResult{}, err
		}
	}
	resultSelector, err := newCVertexSelector(uniqueVertices)
	if err != nil {
		return PageRankResult{}, err
	}
	defer resultSelector.close()

	damping, directed, algorithm, maxIterations, tolerance, weights, err :=
		g.preparePageRankOptions(options)
	if err != nil {
		return PageRankResult{}, err
	}
	if weights != nil {
		defer weights.close()
	}
	reset, resetSelector, resetKind, err := g.preparePageRankReset(options, vertexCount)
	if err != nil {
		return PageRankResult{}, err
	}
	if reset != nil {
		defer reset.close()
	}
	if resetSelector != nil {
		defer resetSelector.close()
	}
	var cResetSelector C.igraph_vs_t
	if resetSelector != nil {
		cResetSelector = resetSelector.value
	}
	result, err := newRealVectorSize(0)
	if err != nil {
		return PageRankResult{}, err
	}
	defer result.close()
	var eigenvalue C.igraph_real_t
	code := C.go_igraph_pagerank(
		&g.graph, edgeWeightPointer(weights), &result.value, &eigenvalue,
		edgeWeightPointer(reset), cResetSelector, C.int(resetKind),
		C.igraph_real_t(damping), booltoint(directed), resultSelector.value,
		algorithm, C.int(maxIterations), C.igraph_real_t(tolerance),
	)
	if code != C.IGRAPH_SUCCESS {
		return PageRankResult{}, igraphError("calculate PageRank", int(code))
	}
	scores, err := result.slice()
	if err != nil {
		return PageRankResult{}, err
	}
	if len(scores) != len(selectedIDs) {
		scores = expandByPositions(scores, positions)
	}
	return PageRankResult{Scores: scores, Eigenvalue: float64(eigenvalue)}, nil
}

func validateSpectralSolver(options SpectralSolverOptions) (int, float64, error) {
	if options.MaxIterations < 0 {
		return 0, 0, fmt.Errorf("igraph: solver maximum iterations must be non-negative: %d", options.MaxIterations)
	}
	if options.MaxIterations > math.MaxInt32 {
		return 0, 0, fmt.Errorf("igraph: solver maximum iterations must not exceed %d: %d", math.MaxInt32, options.MaxIterations)
	}
	if math.IsNaN(options.Tolerance) || math.IsInf(options.Tolerance, 0) || options.Tolerance < 0 {
		return 0, 0, fmt.Errorf("igraph: solver tolerance must be finite and non-negative: %v", options.Tolerance)
	}
	return options.MaxIterations, options.Tolerance, nil
}

func newOptionalNonNegativeEdgeWeights(values []float64, edgeCount int) (*realVector, error) {
	if values != nil {
		for index, value := range values {
			if value < 0 {
				return nil, fmt.Errorf("igraph: weight at index %d must be non-negative: %v", index, value)
			}
		}
	}
	return newOptionalEdgeWeights(values, edgeCount)
}

func (g *Graph) preparePageRankOptions(
	options PageRankOptions,
) (float64, bool, C.igraph_pagerank_algo_t, int, float64, *realVector, error) {
	damping := 0.85
	if options.Damping != nil {
		damping = *options.Damping
	}
	if math.IsNaN(damping) || math.IsInf(damping, 0) || damping < 0 || damping >= 1 {
		return 0, false, 0, 0, 0, nil, fmt.Errorf("igraph: PageRank damping must be finite and in [0, 1): %v", damping)
	}
	directed, err := options.Direction.directed(C.igraph_is_directed(&g.graph) != booltoint(false))
	if err != nil {
		return 0, false, 0, 0, 0, nil, err
	}
	algorithm, err := options.Algorithm.cValue()
	if err != nil {
		return 0, false, 0, 0, 0, nil, err
	}
	maxIterations, tolerance, err := validateSpectralSolver(options.Solver)
	if err != nil {
		return 0, false, 0, 0, 0, nil, err
	}
	weights, err := newOptionalNonNegativeEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return 0, false, 0, 0, 0, nil, err
	}
	return damping, directed, algorithm, maxIterations, tolerance, weights, nil
}

func (g *Graph) preparePageRankReset(
	options PageRankOptions,
	vertexCount int,
) (*realVector, *cVertexSelector, int, error) {
	if options.ResetDistribution != nil && options.ResetVertices != nil {
		return nil, nil, 0, fmt.Errorf("igraph: PageRank reset distribution and reset vertices are mutually exclusive")
	}
	if options.ResetDistribution != nil {
		if len(options.ResetDistribution) != vertexCount {
			return nil, nil, 0, fmt.Errorf(
				"igraph: PageRank reset distribution length %d does not match vertex count %d",
				len(options.ResetDistribution), vertexCount,
			)
		}
		sum := 0.0
		for index, value := range options.ResetDistribution {
			if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
				return nil, nil, 0, fmt.Errorf("igraph: PageRank reset value at index %d must be finite and non-negative: %v", index, value)
			}
			sum += value
		}
		if sum <= 0 || math.IsInf(sum, 0) {
			return nil, nil, 0, fmt.Errorf("igraph: PageRank reset distribution must have a finite positive sum")
		}
		reset, err := newRealVector(options.ResetDistribution)
		return reset, nil, 1, err
	}
	if options.ResetVertices == nil {
		return nil, nil, 0, nil
	}
	if err := validateVertexSelector(*options.ResetVertices, vertexCount); err != nil {
		return nil, nil, 0, fmt.Errorf("igraph: invalid PageRank reset selector: %w", err)
	}
	resetIDs, err := materializeVertexIDs(&g.graph, *options.ResetVertices)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("igraph: materialize PageRank reset selector: %w", err)
	}
	uniqueIDs, _ := deduplicateIDs(resetIDs)
	if len(uniqueIDs) == 0 {
		return nil, nil, 0, fmt.Errorf("igraph: PageRank reset selector must not be empty")
	}
	resetVertices, err := VertexIDs(uniqueIDs...)
	if err != nil {
		return nil, nil, 0, err
	}
	selector, err := newCVertexSelector(resetVertices)
	if err != nil {
		return nil, nil, 0, err
	}
	return nil, selector, 2, nil
}

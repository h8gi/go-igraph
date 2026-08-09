package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
import "C"

import (
	"fmt"
	"math"
)

// CentralizationResult contains node scores where the calculation produces
// them, the graph-level centralization, and its theoretical maximum. Scores is
// always a non-nil, Go-owned slice. Normalized records whether Value was divided
// by TheoreticalMaximum. Specialized empty-graph calculations return empty
// Scores, NaN Value, and a zero theoretical maximum in raw mode; normalized
// empty or single-vertex calculations return an error instead of dividing by
// zero.
type CentralizationResult struct {
	Scores             []float64
	Value              float64
	TheoreticalMaximum float64
	Normalized         bool
}

// CalculateCentralization computes Freeman centralization from caller-provided
// node scores and an explicit theoretical maximum. Scores is borrowed only for
// the call and copied into the result. Scores and the maximum must be finite;
// the maximum must be non-negative and must be positive when normalized is
// requested. Empty scores produce upstream NaN centralization.
//
//igraph:bind igraph_centralization
func CalculateCentralization(
	scores []float64,
	theoreticalMaximum float64,
	normalized bool,
) (CentralizationResult, error) {
	for index, score := range scores {
		if math.IsNaN(score) || math.IsInf(score, 0) {
			return CentralizationResult{}, fmt.Errorf("igraph: centralization score at index %d must be finite: %v", index, score)
		}
	}
	if err := validateCentralizationMaximum(theoreticalMaximum, normalized); err != nil {
		return CentralizationResult{}, err
	}
	cScores, err := newRealVector(scores)
	if err != nil {
		return CentralizationResult{}, err
	}
	defer cScores.close()
	value := float64(C.igraph_centralization(
		&cScores.value, C.igraph_real_t(theoreticalMaximum), booltoint(normalized),
	))
	return CentralizationResult{
		Scores:             append([]float64{}, scores...),
		Value:              value,
		TheoreticalMaximum: theoreticalMaximum,
		Normalized:         normalized,
	}, nil
}

// DegreeCentralizationOptions controls degree centralization. CountLoops uses
// the same graph-theoretic loop convention as DegreeOptions.
type DegreeCentralizationOptions struct {
	Direction  DirectionMode
	CountLoops bool
	Normalized bool
}

// DegreeCentralization calculates unweighted degree scores and their graph
// centralization. Node scores match Degree converted to float64.
//
//igraph:bind igraph_centralization_degree
//igraph:bind igraph_centralization_degree_tmax
func (g *Graph) DegreeCentralization(options DegreeCentralizationOptions) (CentralizationResult, error) {
	if g == nil {
		return CentralizationResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return CentralizationResult{}, ErrClosed
	}
	mode, err := options.Direction.cValue()
	if err != nil {
		return CentralizationResult{}, err
	}
	loops := C.igraph_loops_t(C.IGRAPH_NO_LOOPS)
	if options.CountLoops {
		loops = C.igraph_loops_t(C.IGRAPH_LOOPS)
	}
	if C.igraph_vcount(&g.graph) == 0 {
		return emptyGraphCentralization(options.Normalized)
	}
	var theoreticalMaximum C.igraph_real_t
	if code := C.go_igraph_centralization_degree_tmax(
		&g.graph, &theoreticalMaximum, mode, loops,
	); code != C.IGRAPH_SUCCESS {
		return CentralizationResult{}, igraphError("calculate degree centralization maximum", int(code))
	}
	return collectCentralization(float64(theoreticalMaximum), options.Normalized, func(
		scores *C.igraph_vector_t,
		value *C.igraph_real_t,
	) C.igraph_error_t {
		return C.go_igraph_centralization_degree(
			&g.graph, scores, mode, loops, value, booltoint(options.Normalized),
		)
	})
}

// BetweennessCentralizationOptions controls unweighted whole-graph vertex
// betweenness centralization. DirectedPaths is ignored for undirected graphs.
type BetweennessCentralizationOptions struct {
	DirectedPaths bool
	Normalized    bool
}

// BetweennessCentralization returns raw vertex betweenness scores and their
// graph centralization. Node scores match VertexBetweenness with no cutoff or
// weights.
//
//igraph:bind igraph_centralization_betweenness
//igraph:bind igraph_centralization_betweenness_tmax
func (g *Graph) BetweennessCentralization(options BetweennessCentralizationOptions) (CentralizationResult, error) {
	if g == nil {
		return CentralizationResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return CentralizationResult{}, ErrClosed
	}
	directed := options.DirectedPaths && C.igraph_is_directed(&g.graph) != booltoint(false)
	if C.igraph_vcount(&g.graph) == 0 {
		return emptyGraphCentralization(options.Normalized)
	}
	var theoreticalMaximum C.igraph_real_t
	if code := C.go_igraph_centralization_betweenness_tmax(
		&g.graph, &theoreticalMaximum, booltoint(directed),
	); code != C.IGRAPH_SUCCESS {
		return CentralizationResult{}, igraphError("calculate betweenness centralization maximum", int(code))
	}
	return collectCentralization(float64(theoreticalMaximum), options.Normalized, func(
		scores *C.igraph_vector_t,
		value *C.igraph_real_t,
	) C.igraph_error_t {
		return C.go_igraph_centralization_betweenness(
			&g.graph, scores, booltoint(directed), value, booltoint(options.Normalized),
		)
	})
}

// ClosenessCentralizationOptions controls unweighted closeness centralization.
type ClosenessCentralizationOptions struct {
	Direction  DirectionMode
	Normalized bool
}

// ClosenessCentralization returns raw closeness scores and their graph
// centralization. Node scores match normalized Closeness with no cutoff or
// weights; this upstream specialized routine always uses normalized node
// closeness even when graph centralization itself is raw. If upstream closeness
// is undefined for a vertex, NaN is retained in Scores and may propagate to
// Value.
//
//igraph:bind igraph_centralization_closeness
//igraph:bind igraph_centralization_closeness_tmax
func (g *Graph) ClosenessCentralization(options ClosenessCentralizationOptions) (CentralizationResult, error) {
	if g == nil {
		return CentralizationResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return CentralizationResult{}, ErrClosed
	}
	mode, err := options.Direction.cValue()
	if err != nil {
		return CentralizationResult{}, err
	}
	if C.igraph_vcount(&g.graph) == 0 {
		return emptyGraphCentralization(options.Normalized)
	}
	var theoreticalMaximum C.igraph_real_t
	if code := C.go_igraph_centralization_closeness_tmax(
		&g.graph, &theoreticalMaximum, mode,
	); code != C.IGRAPH_SUCCESS {
		return CentralizationResult{}, igraphError("calculate closeness centralization maximum", int(code))
	}
	return collectCentralization(float64(theoreticalMaximum), options.Normalized, func(
		scores *C.igraph_vector_t,
		value *C.igraph_real_t,
	) C.igraph_error_t {
		return C.go_igraph_centralization_closeness(
			&g.graph, scores, mode, value, booltoint(options.Normalized),
		)
	})
}

// EigenvectorCentralizationOptions controls unweighted eigenvector
// centralization without exposing an ARPACK object.
type EigenvectorCentralizationOptions struct {
	Direction  DirectionMode
	Normalized bool
	Solver     SpectralSolverOptions
}

// EigenvectorCentralization returns max-absolute-value-scaled eigenvector
// scores and their graph centralization. Node scores match unweighted
// EigenvectorCentrality with the same direction and solver settings. Degenerate
// all-zero adjacency matrices retain igraph's documented all-one score vector.
//
//igraph:bind igraph_centralization_eigenvector_centrality
//igraph:bind igraph_centralization_eigenvector_centrality_tmax
func (g *Graph) EigenvectorCentralization(options EigenvectorCentralizationOptions) (CentralizationResult, error) {
	if g == nil {
		return CentralizationResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return CentralizationResult{}, ErrClosed
	}
	mode, err := options.Direction.cValue()
	if err != nil {
		return CentralizationResult{}, err
	}
	maxIterations, tolerance, err := validateSpectralSolver(options.Solver)
	if err != nil {
		return CentralizationResult{}, err
	}
	if C.igraph_vcount(&g.graph) == 0 {
		return emptyGraphCentralization(options.Normalized)
	}
	var theoreticalMaximum C.igraph_real_t
	if code := C.go_igraph_centralization_eigenvector_tmax(
		&g.graph, &theoreticalMaximum, mode,
	); code != C.IGRAPH_SUCCESS {
		return CentralizationResult{}, igraphError("calculate eigenvector centralization maximum", int(code))
	}
	return collectCentralization(float64(theoreticalMaximum), options.Normalized, func(
		scores *C.igraph_vector_t,
		value *C.igraph_real_t,
	) C.igraph_error_t {
		return C.go_igraph_centralization_eigenvector(
			&g.graph, scores, mode, C.int(maxIterations), C.igraph_real_t(tolerance),
			value, booltoint(options.Normalized),
		)
	})
}

func collectCentralization(
	theoreticalMaximum float64,
	normalized bool,
	calculate func(*C.igraph_vector_t, *C.igraph_real_t) C.igraph_error_t,
) (CentralizationResult, error) {
	if err := validateCentralizationMaximum(theoreticalMaximum, normalized); err != nil {
		return CentralizationResult{}, err
	}
	scores, err := newRealVectorSize(0)
	if err != nil {
		return CentralizationResult{}, err
	}
	defer scores.close()
	var value C.igraph_real_t
	if code := calculate(&scores.value, &value); code != C.IGRAPH_SUCCESS {
		return CentralizationResult{}, igraphError("calculate graph centralization", int(code))
	}
	goScores, err := scores.slice()
	if err != nil {
		return CentralizationResult{}, err
	}
	return CentralizationResult{
		Scores:             goScores,
		Value:              float64(value),
		TheoreticalMaximum: theoreticalMaximum,
		Normalized:         normalized,
	}, nil
}

func validateCentralizationMaximum(theoreticalMaximum float64, normalized bool) error {
	if math.IsNaN(theoreticalMaximum) || math.IsInf(theoreticalMaximum, 0) || theoreticalMaximum < 0 {
		return fmt.Errorf("igraph: centralization theoretical maximum must be finite and non-negative: %v", theoreticalMaximum)
	}
	if normalized && theoreticalMaximum == 0 {
		return fmt.Errorf("igraph: cannot normalize centralization with a zero theoretical maximum")
	}
	return nil
}

func emptyGraphCentralization(normalized bool) (CentralizationResult, error) {
	if normalized {
		return CentralizationResult{}, fmt.Errorf("igraph: cannot normalize centralization of an empty graph")
	}
	return CentralizationResult{
		Scores:             []float64{},
		Value:              math.NaN(),
		TheoreticalMaximum: 0,
	}, nil
}

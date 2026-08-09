package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
import "C"

import "fmt"

// BetweennessOptions controls vertex and edge betweenness calculations.
// DirectedPaths makes path direction significant on directed graphs and is
// ignored on undirected graphs. Normalized applies igraph's graph-size- and
// directedness-dependent normalization; normalized vertex scores are NaN when
// fewer than three vertices make the denominator zero. A nil Cutoff includes
// paths of every length; a non-nil Cutoff includes only paths whose length is at
// most that finite, non-negative value.
//
// A nil Weights slice requests an unweighted calculation. A non-nil slice is
// borrowed only for the call, copied into temporary C storage, and must contain
// one finite, strictly positive value per edge. Cutoff is borrowed only for the
// call as well.
type BetweennessOptions struct {
	Weights       []float64
	DirectedPaths bool
	Normalized    bool
	Cutoff        *float64
}

// VertexBetweenness returns one score per selected vertex in materialized
// selector order, including duplicates. Selection controls only which scores
// are returned; the upstream algorithm still performs a whole-graph
// calculation. The result is a non-nil, Go-owned slice that remains valid after
// the source graph is closed.
//
//igraph:bind igraph_betweenness
//igraph:bind igraph_betweenness_cutoff
func (g *Graph) VertexBetweenness(vertices VertexSelector, options BetweennessOptions) ([]float64, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}

	vertexCount := int(C.igraph_vcount(&g.graph))
	if err := validateVertexSelector(vertices, vertexCount); err != nil {
		return nil, fmt.Errorf("igraph: invalid betweenness vertex selector: %w", err)
	}
	selectedIDs, err := materializeVertexIDs(&g.graph, vertices)
	if err != nil {
		return nil, fmt.Errorf("igraph: materialize betweenness vertex selector: %w", err)
	}
	uniqueIDs, positions := deduplicateIDs(selectedIDs)
	uniqueVertices := vertices
	if len(uniqueIDs) != len(selectedIDs) {
		uniqueVertices, err = VertexIDs(uniqueIDs...)
		if err != nil {
			return nil, err
		}
	}
	selector, err := newCVertexSelector(uniqueVertices)
	if err != nil {
		return nil, err
	}
	defer selector.close()
	weights, cutoff, hasCutoff, directed, err := g.prepareBetweennessOptions(options)
	if err != nil {
		return nil, err
	}
	if weights != nil {
		defer weights.close()
	}
	values, err := collectBetweenness("calculate vertex betweenness", func(result *realVector) int {
		return int(C.go_igraph_betweenness(
			&g.graph, edgeWeightPointer(weights), &result.value, selector.value,
			directed, booltoint(options.Normalized), booltoint(hasCutoff),
			C.igraph_real_t(cutoff),
		))
	})
	if err != nil {
		return nil, err
	}
	if len(values) != len(selectedIDs) {
		values = expandByPositions(values, positions)
	}
	return values, nil
}

// EdgeBetweenness returns one score per selected edge in materialized selector
// order, including duplicates. Selection controls only which scores are
// returned; the upstream algorithm still performs a whole-graph calculation.
// The result and input ownership rules match VertexBetweenness.
//
//igraph:bind igraph_edge_betweenness
//igraph:bind igraph_edge_betweenness_cutoff
func (g *Graph) EdgeBetweenness(edges EdgeSelector, options BetweennessOptions) ([]float64, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}

	selectedIDs, err := materializeSelectedEdgeIDs(&g.graph, edges)
	if err != nil {
		return nil, fmt.Errorf("igraph: invalid betweenness edge selector: %w", err)
	}
	uniqueIDs, positions := deduplicateIDs(selectedIDs)
	uniqueEdges := edges
	if edges.kind != edgeSelectorAll || len(uniqueIDs) != len(selectedIDs) {
		uniqueEdges, err = EdgeIDs(uniqueIDs...)
		if err != nil {
			return nil, err
		}
	}
	selector, err := newCEdgeSelector(uniqueEdges)
	if err != nil {
		return nil, err
	}
	defer selector.close()
	weights, cutoff, hasCutoff, directed, err := g.prepareBetweennessOptions(options)
	if err != nil {
		return nil, err
	}
	if weights != nil {
		defer weights.close()
	}
	values, err := collectBetweenness("calculate edge betweenness", func(result *realVector) int {
		return int(C.go_igraph_edge_betweenness(
			&g.graph, edgeWeightPointer(weights), &result.value, selector.value,
			directed, booltoint(options.Normalized), booltoint(hasCutoff),
			C.igraph_real_t(cutoff),
		))
	})
	if err != nil {
		return nil, err
	}
	if len(values) != len(selectedIDs) {
		values = expandByPositions(values, positions)
	}
	return values, nil
}

func collectBetweenness(
	operation string,
	calculate func(*realVector) int,
) ([]float64, error) {
	return collectBetweennessWithInitializer(operation, calculate, newRealVectorSize)
}

func collectBetweennessWithInitializer(
	operation string,
	calculate func(*realVector) int,
	initialize func(int) (*realVector, error),
) ([]float64, error) {
	result, err := initialize(0)
	if err != nil {
		return nil, err
	}
	defer result.close()
	if code := calculate(result); code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError(operation, code)
	}
	return result.slice()
}

func (g *Graph) prepareBetweennessOptions(
	options BetweennessOptions,
) (*realVector, float64, bool, C.igraph_bool_t, error) {
	cutoff, hasCutoff, err := validateCentralityCutoff(options.Cutoff)
	if err != nil {
		return nil, 0, false, booltoint(false), err
	}
	weights, err := newOptionalPositiveEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return nil, 0, false, booltoint(false), err
	}
	directed := options.DirectedPaths && C.igraph_is_directed(&g.graph) != booltoint(false)
	return weights, cutoff, hasCutoff, booltoint(directed), nil
}

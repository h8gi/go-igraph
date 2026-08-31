package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
import "C"

import "fmt"

// BarratTransitivity returns weighted local clustering coefficients in
// materialized vertex-selector order, including duplicates. Edge directions
// are ignored. The graph must be simple after directions are ignored: loops,
// parallel edges, and mutual directed pairs are rejected.
//
// Weights is required, borrowed only for the synchronous call, copied into
// temporary C storage, and must contain one finite value per edge. Mode
// controls whether coefficients for vertices with zero strength are NaN or
// zero. The returned non-nil slice is Go-owned and survives graph closure.
//
//igraph:bind igraph_transitivity_barrat
func (g *Graph) BarratTransitivity(
	vertices VertexSelector,
	weights []float64,
	mode TransitivityMode,
) ([]float64, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	if weights == nil {
		return nil, fmt.Errorf("igraph: Barrat transitivity requires edge weights")
	}
	cMode, err := mode.cValue()
	if err != nil {
		return nil, err
	}
	if err := validateVertexSelector(vertices, int(C.igraph_vcount(&g.graph))); err != nil {
		return nil, fmt.Errorf("igraph: invalid Barrat transitivity vertex selector: %w", err)
	}
	var simple C.igraph_bool_t
	if code := C.go_igraph_is_simple_undirected(&g.graph, &simple); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("validate Barrat transitivity graph", int(code))
	}
	if simple == booltoint(false) {
		return nil, fmt.Errorf("igraph: Barrat transitivity requires a simple graph when directions are ignored")
	}
	selectedIDs, err := materializeVertexIDs(&g.graph, vertices)
	if err != nil {
		return nil, fmt.Errorf("igraph: materialize Barrat transitivity vertex selector: %w", err)
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
	cWeights, err := newOptionalEdgeWeights(weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return nil, err
	}
	defer cWeights.close()
	values, err := collectBetweenness("calculate Barrat transitivity", func(result *realVector) int {
		return int(C.go_igraph_transitivity_barrat(
			&g.graph, &result.value, selector.value, &cWeights.value, cMode,
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

// EdgeClusteringOptions controls k-cycle edge clustering. CycleSize must be 3
// or 4. Offset adds one to the cycle count z. Normalize divides by the maximum
// possible count s, independently of Offset.
type EdgeClusteringOptions struct {
	CycleSize int
	Offset    bool
	Normalize bool
}

// EdgeClustering returns one k-cycle clustering coefficient per selected edge
// in materialized selector order, including duplicates. Directions and edge
// multiplicities are ignored when counting cycles z, while loops and
// multiplicities remain in endpoint degrees used by normalization denominator
// s. Thus normalized loops can produce NaN from a zero denominator.
//
// The selector is borrowed for the synchronous call. The returned non-nil
// slice is Go-owned and remains valid after graph closure.
//
//igraph:bind igraph_ecc
func (g *Graph) EdgeClustering(
	edges EdgeSelector,
	options EdgeClusteringOptions,
) ([]float64, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	if options.CycleSize != 3 && options.CycleSize != 4 {
		return nil, fmt.Errorf("igraph: edge clustering cycle size must be 3 or 4: %d", options.CycleSize)
	}
	cycleSize, err := intToIgraphInt(options.CycleSize, "edge clustering cycle size")
	if err != nil {
		return nil, err
	}
	selectedIDs, err := materializeSelectedEdgeIDs(&g.graph, edges)
	if err != nil {
		return nil, fmt.Errorf("igraph: invalid edge clustering selector: %w", err)
	}
	uniqueIDs, positions := deduplicateIDs(selectedIDs)
	uniqueEdges, err := EdgeIDs(uniqueIDs...)
	if err != nil {
		return nil, err
	}
	selector, err := newCEdgeSelector(uniqueEdges)
	if err != nil {
		return nil, err
	}
	defer selector.close()
	values, err := collectBetweenness("calculate edge clustering", func(result *realVector) int {
		return int(C.go_igraph_ecc(
			&g.graph, &result.value, selector.value, cycleSize,
			booltoint(options.Offset), booltoint(options.Normalize),
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

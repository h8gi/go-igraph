package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
import "C"

import (
	"fmt"
	"math"
)

// BurtConstraint returns one structural-hole constraint score per selected
// vertex in materialized selector order, including duplicates. For directed
// graphs, each tie strength combines both directions. Parallel-edge strengths
// are added; loops are excluded from the proportional-strength denominator.
// Isolates yield NaN. A vertex incident only to zero-strength ties follows the
// pinned igraph arithmetic and may also yield NaN.
//
// A nil weights slice assigns equal strength to every edge. Otherwise weights
// is borrowed only for the call, copied into temporary C storage, and must
// contain one finite, non-negative strength per edge. The returned non-nil
// slice is Go-owned and remains valid after the graph is closed.
//
//igraph:bind igraph_constraint
func (g *Graph) BurtConstraint(vertices VertexSelector, weights []float64) ([]float64, error) {
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
		return nil, fmt.Errorf("igraph: invalid Burt constraint vertex selector: %w", err)
	}
	selectedIDs, err := materializeVertexIDs(&g.graph, vertices)
	if err != nil {
		return nil, fmt.Errorf("igraph: materialize Burt constraint vertex selector: %w", err)
	}
	uniqueIDs, positions := deduplicateIDs(selectedIDs)
	uniqueVertices, err := VertexIDs(uniqueIDs...)
	if err != nil {
		return nil, err
	}
	selector, err := newCVertexSelector(uniqueVertices)
	if err != nil {
		return nil, err
	}
	defer selector.close()
	cWeights, err := newOptionalConstraintWeights(weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return nil, err
	}
	if cWeights != nil {
		defer cWeights.close()
	}

	values, err := collectBetweenness("calculate Burt constraint", func(result *realVector) int {
		return int(C.go_igraph_constraint(
			&g.graph, &result.value, selector.value, edgeWeightPointer(cWeights),
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

func newOptionalConstraintWeights(values []float64, edgeCount int) (*realVector, error) {
	if values != nil {
		for index, value := range values {
			if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
				return nil, fmt.Errorf("igraph: constraint weight at index %d must be finite and non-negative: %v", index, value)
			}
		}
	}
	return newOptionalEdgeWeights(values, edgeCount)
}

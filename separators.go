package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "separators_cgo.h"
import "C"

import "fmt"

// IsSeparator reports whether removing the selected vertex set separates at
// least one pair of remaining vertices. Edge directions are ignored. The
// selector is borrowed only for this call, materialized while the graph is
// locked, and must not contain duplicate vertex IDs. The empty set and the set
// of all vertices are not separators. On an already disconnected graph, the
// empty set is not considered a separator.
//
//igraph:bind igraph_is_separator
func (g *Graph) IsSeparator(candidate VertexSelector) (bool, error) {
	return g.separatorPredicate(candidate, "check vertex separator", func(graph *C.igraph_t, selector C.igraph_vs_t) (bool, int) {
		var result C.igraph_bool_t
		code := C.go_igraph_is_separator(graph, selector, &result)
		return result != booltoint(false), int(code)
	})
}

// IsMinimalSeparator reports whether the selected vertices form a separator
// and no proper subset is also a separator. Its direction, borrowing,
// duplicate, empty, and degenerate-graph semantics match IsSeparator.
//
//igraph:bind igraph_is_minimal_separator
func (g *Graph) IsMinimalSeparator(candidate VertexSelector) (bool, error) {
	return g.separatorPredicate(candidate, "check minimal vertex separator", func(graph *C.igraph_t, selector C.igraph_vs_t) (bool, int) {
		var result C.igraph_bool_t
		code := C.go_igraph_is_minimal_separator(graph, selector, &result)
		return result != booltoint(false), int(code)
	})
}

type separatorPredicateAdapters struct {
	materialize func(VertexSelector) ([]int, error)
	newSelector func(VertexSelector) (*cVertexSelector, error)
	close       func(*cVertexSelector)
	query       func(*cVertexSelector) (bool, int)
}

func (g *Graph) separatorPredicate(
	candidate VertexSelector,
	action string,
	query func(*C.igraph_t, C.igraph_vs_t) (bool, int),
) (bool, error) {
	if g == nil {
		return false, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return false, ErrClosed
	}
	if err := validateVertexSelector(candidate, int(C.igraph_vcount(&g.graph))); err != nil {
		return false, err
	}
	return evaluateSeparatorPredicate(candidate, action, separatorPredicateAdapters{
		materialize: func(selector VertexSelector) ([]int, error) {
			return materializeVertexIDs(&g.graph, selector)
		},
		newSelector: newCVertexSelector,
		close:       (*cVertexSelector).close,
		query: func(selector *cVertexSelector) (bool, int) {
			return query(&g.graph, selector.value)
		},
	})
}

func evaluateSeparatorPredicate(
	candidate VertexSelector,
	action string,
	adapters separatorPredicateAdapters,
) (bool, error) {
	ids, err := adapters.materialize(candidate)
	if err != nil {
		return false, err
	}
	seen := make(map[int]struct{}, len(ids))
	for index, id := range ids {
		if _, duplicate := seen[id]; duplicate {
			return false, fmt.Errorf("igraph: separator candidate contains duplicate vertex ID %d at materialized index %d", id, index)
		}
		seen[id] = struct{}{}
	}
	explicit, err := VertexIDs(ids...)
	if err != nil {
		return false, err
	}
	selector, err := adapters.newSelector(explicit)
	if err != nil {
		return false, err
	}
	defer adapters.close(selector)
	result, code := adapters.query(selector)
	if code != int(C.IGRAPH_SUCCESS) {
		return false, igraphError(action, code)
	}
	return result, nil
}

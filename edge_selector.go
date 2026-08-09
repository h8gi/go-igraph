package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
import "C"

import "fmt"

type edgeSelectorKind uint8

const (
	edgeSelectorAll edgeSelectorKind = iota
	edgeSelectorNone
	edgeSelectorIDs
	edgeSelectorPairs
)

// EdgeSelector describes a reusable edge selection without retaining a graph
// or C resource. Its zero value selects all edges in edge ID order.
//
// Explicit IDs preserve order and duplicates. Endpoint pairs also preserve
// order and duplicates. On directed graphs, the pair's directed flag controls
// whether endpoint order matters; undirected graphs always ignore it. Loops
// are valid pairs. When parallel edges exist, a pair resolves to one stable
// edge chosen by igraph; use EdgeIDs to select a specific parallel edge.
// Missing edges and graph-specific bounds are checked when the selector is
// materialized.
type EdgeSelector struct {
	kind     edgeSelectorKind
	ids      []int
	pairs    []Edge
	directed bool
}

// AllEdges returns a selector for every edge in edge ID order.
func AllEdges() EdgeSelector {
	return EdgeSelector{}
}

// NoEdges returns an empty selector.
func NoEdges() EdgeSelector {
	return EdgeSelector{kind: edgeSelectorNone}
}

// EdgeIDs returns a selector that preserves supplied order and duplicates. The
// input is copied and may be reused or changed immediately.
func EdgeIDs(ids ...int) (EdgeSelector, error) {
	result := EdgeSelector{kind: edgeSelectorIDs, ids: append([]int{}, ids...)}
	for index, id := range result.ids {
		if id < 0 {
			return EdgeSelector{}, fmt.Errorf("igraph: edge ID at index %d must be non-negative: %d", index, id)
		}
	}
	return result, nil
}

// EdgePairs returns a selector for edges matching endpoint pairs. When
// directed is true, endpoint order matters on directed graphs. The input is
// copied and may be reused or changed immediately.
func EdgePairs(pairs []Edge, directed bool) (EdgeSelector, error) {
	result := EdgeSelector{
		kind:     edgeSelectorPairs,
		pairs:    append([]Edge{}, pairs...),
		directed: directed,
	}
	for index, pair := range result.pairs {
		if pair.From < 0 {
			return EdgeSelector{}, fmt.Errorf("igraph: edge pair %d source must be non-negative: %d", index, pair.From)
		}
		if pair.To < 0 {
			return EdgeSelector{}, fmt.Errorf("igraph: edge pair %d target must be non-negative: %d", index, pair.To)
		}
	}
	return result, nil
}

func (selector EdgeSelector) empty() bool {
	switch selector.kind {
	case edgeSelectorNone:
		return true
	case edgeSelectorIDs:
		return len(selector.ids) == 0
	case edgeSelectorPairs:
		return len(selector.pairs) == 0
	default:
		return false
	}
}

// cEdgeSelector holds an immediate all selector or an owned regular selector
// containing a C-owned copy of explicit edge IDs. No C iterator retains a
// pointer into Go memory.
type cEdgeSelector struct {
	value C.igraph_es_t
	owned bool
}

//igraph:internal igraph_ess_all
//igraph:internal igraph_es_vector_copy
func newCEdgeSelector(selector EdgeSelector) (*cEdgeSelector, error) {
	result := &cEdgeSelector{}
	switch selector.kind {
	case edgeSelectorAll:
		result.value = C.igraph_ess_all(C.IGRAPH_EDGEORDER_ID)
	case edgeSelectorIDs:
		backing, err := newIntVector(selector.ids)
		if err != nil {
			return nil, err
		}
		code := C.go_igraph_es_vector_copy(&result.value, &backing.value)
		backing.close()
		if code != C.IGRAPH_SUCCESS {
			return nil, igraphError("copy explicit edge selector", int(code))
		}
		result.owned = true
	default:
		return nil, fmt.Errorf("igraph: invalid edge selector kind: %d", selector.kind)
	}
	return result, nil
}

//igraph:internal igraph_es_destroy
func (selector *cEdgeSelector) close() {
	if selector.owned {
		C.igraph_es_destroy(&selector.value)
	}
}

// edgeIDs materializes a selector as independently owned Go storage. It is the
// common graph-specific validation boundary for later algorithm consumers.
func (g *Graph) edgeIDs(selector EdgeSelector) ([]int, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	return materializeSelectedEdgeIDs(&g.graph, selector)
}

// materializeSelectedEdgeIDs validates and materializes an edge selector while
// its caller holds the graph lock.
func materializeSelectedEdgeIDs(graph *C.igraph_t, selector EdgeSelector) ([]int, error) {
	edgeCount := int(C.igraph_ecount(graph))
	vertexCount := int(C.igraph_vcount(graph))
	switch selector.kind {
	case edgeSelectorAll, edgeSelectorNone:
	case edgeSelectorIDs:
		for index, id := range selector.ids {
			if id >= edgeCount {
				return nil, fmt.Errorf(
					"igraph: edge ID at selector index %d is %d, outside [0, %d)",
					index, id, edgeCount,
				)
			}
		}
	case edgeSelectorPairs:
		resolved := make([]int, len(selector.pairs))
		for index, pair := range selector.pairs {
			if err := validateEdge(pair, vertexCount, index); err != nil {
				return nil, err
			}
			var edgeID C.igraph_int_t
			if code := C.go_igraph_get_eid(
				graph,
				&edgeID,
				C.igraph_int_t(pair.From),
				C.igraph_int_t(pair.To),
				booltoint(selector.directed),
				booltoint(false),
			); code != C.IGRAPH_SUCCESS {
				return nil, igraphError("validate edge pair", int(code))
			}
			if edgeID < 0 {
				return nil, fmt.Errorf(
					"igraph: edge pair %d (%d, %d) does not exist",
					index, pair.From, pair.To,
				)
			}
			resolvedID, err := igraphIntToInt(edgeID, "resolved edge ID")
			if err != nil {
				return nil, err
			}
			resolved[index] = resolvedID
		}
		selector = EdgeSelector{kind: edgeSelectorIDs, ids: resolved}
	default:
		return nil, fmt.Errorf("igraph: invalid edge selector kind: %d", selector.kind)
	}
	if selector.empty() {
		return []int{}, nil
	}

	return materializeEdgeIDs(graph, selector)
}

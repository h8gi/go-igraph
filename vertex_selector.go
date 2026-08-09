package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
import "C"

import "fmt"

type vertexSelectorKind uint8

const (
	vertexSelectorAll vertexSelectorKind = iota
	vertexSelectorNone
	vertexSelectorIDs
	vertexSelectorRange
)

// VertexSelector describes a reusable vertex selection without retaining a
// graph or any C resource. Its zero value selects all vertices.
//
// Explicit IDs preserve caller order and duplicates. VertexRange uses a
// half-open [start, end) interval. Negative IDs and malformed ranges are
// rejected when constructed; graph-specific upper bounds are checked whenever
// a selector is materialized for a graph.
type VertexSelector struct {
	kind  vertexSelectorKind
	ids   []int
	start int
	end   int
}

// AllVertices returns a selector for every vertex in vertex ID order.
func AllVertices() VertexSelector {
	return VertexSelector{}
}

// NoVertices returns an empty selector.
func NoVertices() VertexSelector {
	return VertexSelector{kind: vertexSelectorNone}
}

// VertexIDs returns a selector that preserves the supplied order and
// duplicates. The input is copied and can be reused or changed immediately.
func VertexIDs(ids ...int) (VertexSelector, error) {
	result := VertexSelector{kind: vertexSelectorIDs, ids: append([]int{}, ids...)}
	for index, id := range result.ids {
		if id < 0 {
			return VertexSelector{}, fmt.Errorf("igraph: vertex ID at index %d must be non-negative: %d", index, id)
		}
	}
	return result, nil
}

// VertexRange returns a selector for the half-open interval [start, end).
func VertexRange(start, end int) (VertexSelector, error) {
	if start < 0 {
		return VertexSelector{}, fmt.Errorf("igraph: vertex range start must be non-negative: %d", start)
	}
	if end < start {
		return VertexSelector{}, fmt.Errorf("igraph: vertex range end %d is before start %d", end, start)
	}
	if _, err := intToIgraphInt(start, "vertex range start"); err != nil {
		return VertexSelector{}, err
	}
	if _, err := intToIgraphInt(end, "vertex range end"); err != nil {
		return VertexSelector{}, err
	}
	return VertexSelector{kind: vertexSelectorRange, start: start, end: end}, nil
}

// cVertexSelector holds an immediate selector or an owned regular selector for
// explicit IDs. The regular selector copies its backing vector into C-owned
// memory so C iterators never retain a pointer into Go memory.
type cVertexSelector struct {
	value C.igraph_vs_t
	owned bool
}

//igraph:internal igraph_vss_all
//igraph:internal igraph_vss_none
//igraph:internal igraph_vs_vector_copy
//igraph:internal igraph_vss_range
func newCVertexSelector(selector VertexSelector) (*cVertexSelector, error) {
	result := &cVertexSelector{}
	switch selector.kind {
	case vertexSelectorAll:
		result.value = C.igraph_vss_all()
	case vertexSelectorNone:
		result.value = C.igraph_vss_none()
	case vertexSelectorIDs:
		backing, err := newIntVector(selector.ids)
		if err != nil {
			return nil, err
		}
		code := C.go_igraph_vs_vector_copy(&result.value, &backing.value)
		backing.close()
		if code != C.IGRAPH_SUCCESS {
			return nil, igraphError("copy explicit vertex selector", int(code))
		}
		result.owned = true
	case vertexSelectorRange:
		result.value = C.igraph_vss_range(
			C.igraph_int_t(selector.start),
			C.igraph_int_t(selector.end),
		)
	default:
		return nil, fmt.Errorf("igraph: invalid vertex selector kind: %d", selector.kind)
	}
	return result, nil
}

//igraph:internal igraph_vs_destroy
func (selector *cVertexSelector) close() {
	if selector.owned {
		C.igraph_vs_destroy(&selector.value)
	}
}

// vertexIDs materializes a selector as an independently owned Go slice. It is
// the common validation boundary for later algorithms that consume selectors.
func (g *Graph) vertexIDs(selector VertexSelector) ([]int, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}

	vertexCount := int(C.igraph_vcount(&g.graph))
	if err := validateVertexSelector(selector, vertexCount); err != nil {
		return nil, err
	}

	return materializeVertexIDs(&g.graph, selector)
}

// validateVertexSelector checks the selector against a graph's current vertex
// count. Callers that already hold the graph mutex use this helper to keep
// selector validation and the C operation in the same critical section.
func validateVertexSelector(selector VertexSelector, vertexCount int) error {
	switch selector.kind {
	case vertexSelectorAll, vertexSelectorNone:
	case vertexSelectorIDs:
		for index, id := range selector.ids {
			if id < 0 {
				return fmt.Errorf("igraph: vertex ID at selector index %d must be non-negative: %d", index, id)
			}
			if id >= vertexCount {
				return fmt.Errorf(
					"igraph: vertex ID at selector index %d is %d, outside [0, %d)",
					index, id, vertexCount,
				)
			}
		}
	case vertexSelectorRange:
		if selector.start < 0 {
			return fmt.Errorf("igraph: vertex range start must be non-negative: %d", selector.start)
		}
		if selector.end < selector.start {
			return fmt.Errorf("igraph: vertex range end %d is before start %d", selector.end, selector.start)
		}
		if selector.end > vertexCount {
			return fmt.Errorf(
				"igraph: vertex range [%d, %d) exceeds vertex count %d",
				selector.start, selector.end, vertexCount,
			)
		}
	default:
		return fmt.Errorf("igraph: invalid vertex selector kind: %d", selector.kind)
	}
	return nil
}

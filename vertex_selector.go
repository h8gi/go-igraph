package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
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

// cVertexSelector holds an immediate igraph_vs_t and, for explicit IDs, owns
// the backing integer vector borrowed by that selector. Immediate selectors
// must not be passed to igraph_vs_destroy; close releases only the backing
// vector after the selector's final use.
type cVertexSelector struct {
	value   C.igraph_vs_t
	backing *intVector
}

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
		result.backing = backing
		result.value = C.igraph_vss_vector(&backing.value)
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

func (selector *cVertexSelector) close() {
	if selector.backing != nil {
		selector.backing.close()
	}
}

// vertexIDs materializes a selector as an independently owned Go slice. It is
// the common validation boundary for later algorithms that consume selectors.
//
//igraph:internal igraph_vss_all
//igraph:internal igraph_vss_none
//igraph:internal igraph_vss_vector
//igraph:internal igraph_vss_range
//igraph:internal igraph_vs_as_vector
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
	switch selector.kind {
	case vertexSelectorAll, vertexSelectorNone:
	case vertexSelectorIDs:
		for index, id := range selector.ids {
			if id >= vertexCount {
				return nil, fmt.Errorf(
					"igraph: vertex ID at selector index %d is %d, outside [0, %d)",
					index, id, vertexCount,
				)
			}
		}
	case vertexSelectorRange:
		if selector.end > vertexCount {
			return nil, fmt.Errorf(
				"igraph: vertex range [%d, %d) exceeds vertex count %d",
				selector.start, selector.end, vertexCount,
			)
		}
	default:
		return nil, fmt.Errorf("igraph: invalid vertex selector kind: %d", selector.kind)
	}

	cSelector, err := newCVertexSelector(selector)
	if err != nil {
		return nil, err
	}
	defer cSelector.close()
	result, err := newIntVector(nil)
	if err != nil {
		return nil, err
	}
	defer result.close()
	if code := C.igraph_vs_as_vector(&g.graph, cSelector.value, &result.value); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("materialize vertex selector", int(code))
	}
	return result.slice()
}

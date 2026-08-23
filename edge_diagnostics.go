package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "edge_diagnostics_cgo.h"
import "C"

// HasLoopEdges reports whether the graph contains at least one self-loop.
//
//igraph:bind igraph_has_loop
func (g *Graph) HasLoopEdges() (bool, error) {
	if g == nil {
		return false, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return false, ErrClosed
	}
	var result C.igraph_bool_t
	if code := C.go_igraph_has_loop(&g.graph, &result); code != C.IGRAPH_SUCCESS {
		return false, igraphError("check loop edges", int(code))
	}
	return result != booltoint(false), nil
}

// LoopEdgeCount returns the number of self-loop edges in the graph.
//
//igraph:bind igraph_count_loops
func (g *Graph) LoopEdgeCount() (int, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return 0, ErrClosed
	}
	var result C.igraph_int_t
	if code := C.go_igraph_count_loops(&g.graph, &result); code != C.IGRAPH_SUCCESS {
		return 0, igraphError("count loop edges", int(code))
	}
	return igraphIntToInt(result, "loop edge count")
}

// LoopEdgeFlags reports whether each selected edge is a self-loop. The
// selector is borrowed only for the call. The returned non-nil slice preserves
// materialized selector order and duplicates, is Go-owned, and remains valid
// after the graph is closed.
//
//igraph:bind igraph_is_loop
func (g *Graph) LoopEdgeFlags(edges EdgeSelector) ([]bool, error) {
	return g.loopEdgeFlags(edges, nil)
}

// HasMultipleEdges reports whether the graph contains parallel edges. In a
// directed graph, edges are parallel only when their endpoint orientation is
// the same.
//
//igraph:bind igraph_has_multiple
func (g *Graph) HasMultipleEdges() (bool, error) {
	if g == nil {
		return false, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return false, ErrClosed
	}
	return operatorGraphHasMultiple(&g.graph)
}

// EdgeMultiplicities returns, for each selected edge, the total number of
// edges with the same endpoints. The count includes the selected edge itself,
// so every valid edge has multiplicity at least one. Endpoint orientation is
// significant in directed graphs. The selector is borrowed only for the call;
// the returned non-nil slice preserves order and duplicates and is Go-owned.
//
//igraph:bind igraph_count_multiple
func (g *Graph) EdgeMultiplicities(edges EdgeSelector) ([]int, error) {
	return g.edgeMultiplicities(edges, nil)
}

func (g *Graph) edgeMultiplicities(edges EdgeSelector, adapters *edgeDiagnosticAdapters) ([]int, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	selected, err := materializeSelectedEdgeIDs(&g.graph, edges)
	if err != nil {
		return nil, err
	}
	if len(selected) == 0 {
		return []int{}, nil
	}
	selector, err := newCEdgeSelector(EdgeSelector{kind: edgeSelectorIDs, ids: selected})
	if err != nil {
		return nil, err
	}
	defer selector.close()
	resolved := resolveEdgeDiagnosticAdapters(adapters)
	result, err := resolved.newInt(nil)
	if err != nil {
		return nil, err
	}
	defer resolved.closeInt(result)
	if code := resolved.countMultiple(g, result, selector); code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("count edge multiplicities", code)
	}
	return resolved.intSlice(result)
}

// MultipleEdgeFlags reports which selected edges are later occurrences within
// their parallel-edge group. For each endpoint pair, the lowest edge ID is
// false and every later parallel edge is true. This differs intentionally from
// testing whether EdgeMultiplicities is greater than one. The returned non-nil
// Go-owned slice preserves materialized selector order and duplicates.
//
//igraph:bind igraph_is_multiple
func (g *Graph) MultipleEdgeFlags(edges EdgeSelector) ([]bool, error) {
	return g.multipleEdgeFlags(edges, nil)
}

// HasMutualEdges reports whether the graph contains an edge whose reverse is
// also present. IncludeLoops controls whether directed self-loops count as
// mutual. Every edge in an undirected graph is mutual by definition, so a
// non-empty undirected graph returns true regardless of IncludeLoops.
//
//igraph:bind igraph_has_mutual
func (g *Graph) HasMutualEdges(includeLoops bool) (bool, error) {
	if g == nil {
		return false, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return false, ErrClosed
	}
	var result C.igraph_bool_t
	if code := C.go_igraph_has_mutual(&g.graph, &result, booltoint(includeLoops)); code != C.IGRAPH_SUCCESS {
		return false, igraphError("check mutual edges", int(code))
	}
	return result != booltoint(false), nil
}

// MutualEdgeFlags reports whether each selected edge has a reverse edge.
// IncludeLoops controls whether directed self-loops count as mutual. Edge
// multiplicity is ignored: when either orientation exists, all matching edges
// in both orientations are mutual. Every selected edge in an undirected graph
// is mutual. The returned non-nil Go-owned slice preserves selector order and
// duplicates.
//
//igraph:bind igraph_is_mutual
func (g *Graph) MutualEdgeFlags(edges EdgeSelector, includeLoops bool) ([]bool, error) {
	return g.mutualEdgeFlags(edges, includeLoops, nil)
}

type edgeDiagnosticAdapters struct {
	newBool       func([]bool) (*boolVector, error)
	newInt        func([]int) (*intVector, error)
	closeBool     func(*boolVector)
	closeInt      func(*intVector)
	boolSlice     func(*boolVector) ([]bool, error)
	intSlice      func(*intVector) ([]int, error)
	loop          func(*Graph, *boolVector, *cEdgeSelector) int
	multiple      func(*Graph, *boolVector, *cEdgeSelector) int
	mutual        func(*Graph, *boolVector, *cEdgeSelector, bool) int
	countMultiple func(*Graph, *intVector, *cEdgeSelector) int
}

func defaultEdgeDiagnosticAdapters() edgeDiagnosticAdapters {
	return edgeDiagnosticAdapters{
		newBool:   newBoolVector,
		newInt:    newIntVector,
		closeBool: (*boolVector).close,
		closeInt:  (*intVector).close,
		boolSlice: (*boolVector).slice,
		intSlice:  (*intVector).slice,
		loop: func(g *Graph, result *boolVector, selector *cEdgeSelector) int {
			return int(C.go_igraph_is_loop(&g.graph, &result.value, selector.value))
		},
		multiple: func(g *Graph, result *boolVector, selector *cEdgeSelector) int {
			return int(C.go_igraph_is_multiple(&g.graph, &result.value, selector.value))
		},
		mutual: func(g *Graph, result *boolVector, selector *cEdgeSelector, includeLoops bool) int {
			return int(C.go_igraph_is_mutual(
				&g.graph, &result.value, selector.value, booltoint(includeLoops),
			))
		},
		countMultiple: func(g *Graph, result *intVector, selector *cEdgeSelector) int {
			return int(C.go_igraph_count_multiple(&g.graph, &result.value, selector.value))
		},
	}
}

func resolveEdgeDiagnosticAdapters(adapters *edgeDiagnosticAdapters) edgeDiagnosticAdapters {
	if adapters == nil {
		return defaultEdgeDiagnosticAdapters()
	}
	return *adapters
}

func (g *Graph) loopEdgeFlags(edges EdgeSelector, adapters *edgeDiagnosticAdapters) ([]bool, error) {
	resolved := resolveEdgeDiagnosticAdapters(adapters)
	return g.selectedEdgeBoolResult(edges, "check loop edges", resolved, func(result *boolVector, selector *cEdgeSelector) int {
		return resolved.loop(g, result, selector)
	})
}

func (g *Graph) multipleEdgeFlags(edges EdgeSelector, adapters *edgeDiagnosticAdapters) ([]bool, error) {
	resolved := resolveEdgeDiagnosticAdapters(adapters)
	return g.selectedEdgeBoolResult(edges, "check multiple edges", resolved, func(result *boolVector, selector *cEdgeSelector) int {
		return resolved.multiple(g, result, selector)
	})
}

func (g *Graph) mutualEdgeFlags(edges EdgeSelector, includeLoops bool, adapters *edgeDiagnosticAdapters) ([]bool, error) {
	resolved := resolveEdgeDiagnosticAdapters(adapters)
	return g.selectedEdgeBoolResult(edges, "check mutual edges", resolved, func(result *boolVector, selector *cEdgeSelector) int {
		return resolved.mutual(g, result, selector, includeLoops)
	})
}

func (g *Graph) selectedEdgeBoolResult(
	edges EdgeSelector,
	operation string,
	adapters edgeDiagnosticAdapters,
	call func(*boolVector, *cEdgeSelector) int,
) ([]bool, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	selected, err := materializeSelectedEdgeIDs(&g.graph, edges)
	if err != nil {
		return nil, err
	}
	if len(selected) == 0 {
		return []bool{}, nil
	}
	selector, err := newCEdgeSelector(EdgeSelector{kind: edgeSelectorIDs, ids: selected})
	if err != nil {
		return nil, err
	}
	defer selector.close()
	result, err := adapters.newBool(nil)
	if err != nil {
		return nil, err
	}
	defer adapters.closeBool(result)
	if code := call(result, selector); code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError(operation, code)
	}
	return adapters.boolSlice(result)
}

// The single-edge multiplicity entry point is deliberately represented by
// EdgeMultiplicities so callers use one selector-ordered ownership contract.
//
//igraph:unsupported igraph_count_multiple_1

package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "chordal_cgo.h"
import "C"

import "fmt"

// MaximumCardinalityOrder is a Go-owned maximum-cardinality ordering.
// Vertices lists vertex IDs in visit order. PositionByVertex is its inverse:
// PositionByVertex[v] is the index of v in Vertices.
type MaximumCardinalityOrder struct {
	Vertices         []int
	PositionByVertex []int
}

// ChordalityOptions controls chordality analysis. Ordering may be nil to have
// igraph compute one; otherwise it is borrowed for the call and must be a full
// vertex permutation. Complete requests an independently owned chordal graph.
type ChordalityOptions struct {
	Ordering []int
	Complete bool
}

// ChordalityResult contains a chordality decision and its Go-owned fill-in
// edges. Each fill edge is normalized so From < To. Completion is nil unless
// requested; when present it is independently owned and must be closed.
type ChordalityResult struct {
	Chordal    bool
	FillEdges  []Edge
	Completion *Graph
}

// MaximumCardinalityOrder returns a deterministic ordering. Edge directions,
// loops and parallel multiplicity do not affect adjacency for this operation.
//
//igraph:bind igraph_maximum_cardinality_search
func (g *Graph) MaximumCardinalityOrder() (MaximumCardinalityOrder, error) {
	if g == nil {
		return MaximumCardinalityOrder{}, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return MaximumCardinalityOrder{}, ErrClosed
	}
	alpha, err := newIntVector(nil)
	if err != nil {
		return MaximumCardinalityOrder{}, err
	}
	defer alpha.close()
	inverse, err := newIntVector(nil)
	if err != nil {
		return MaximumCardinalityOrder{}, err
	}
	defer inverse.close()
	if code := C.go_igraph_maximum_cardinality_search(&g.graph, &alpha.value, &inverse.value); code != C.IGRAPH_SUCCESS {
		return MaximumCardinalityOrder{}, igraphError("calculate maximum-cardinality order", int(code))
	}
	ranks, err := alpha.slice()
	if err != nil {
		return MaximumCardinalityOrder{}, err
	}
	byRank, err := inverse.slice()
	if err != nil {
		return MaximumCardinalityOrder{}, err
	}
	return maximumCardinalityOrderFromSlices(ranks, byRank)
}

func maximumCardinalityOrderFromSlices(ranks, byRank []int) (MaximumCardinalityOrder, error) {
	n := len(ranks)
	if len(byRank) != n {
		return MaximumCardinalityOrder{}, fmt.Errorf("igraph: maximum-cardinality order length mismatch: %d ranks, %d inverse ranks", n, len(byRank))
	}
	vertices := make([]int, n)
	positions := make([]int, n)
	seen := make([]bool, n)
	for position := range n {
		vertex := byRank[n-1-position]
		if vertex < 0 || vertex >= n || seen[vertex] {
			return MaximumCardinalityOrder{}, fmt.Errorf("igraph: invalid maximum-cardinality vertex %d at position %d", vertex, position)
		}
		if ranks[vertex] != n-1-position {
			return MaximumCardinalityOrder{}, fmt.Errorf("igraph: inconsistent maximum-cardinality rank for vertex %d", vertex)
		}
		seen[vertex] = true
		vertices[position] = vertex
		positions[vertex] = position
	}
	return MaximumCardinalityOrder{Vertices: vertices, PositionByVertex: positions}, nil
}

func orderingVectors(order []int, vertexCount int) (*intVector, *intVector, error) {
	return orderingVectorsWithInitializer(order, vertexCount, newIntVector)
}

func orderingVectorsWithInitializer(order []int, vertexCount int, initialize func([]int) (*intVector, error)) (*intVector, *intVector, error) {
	if len(order) != vertexCount {
		return nil, nil, fmt.Errorf("igraph: ordering length %d does not match vertex count %d", len(order), vertexCount)
	}
	ranks := make([]int, vertexCount)
	byRank := make([]int, vertexCount)
	seen := make([]bool, vertexCount)
	for position, vertex := range order {
		if vertex < 0 || vertex >= vertexCount || seen[vertex] {
			return nil, nil, fmt.Errorf("igraph: ordering must be a vertex permutation; invalid vertex %d at position %d", vertex, position)
		}
		seen[vertex] = true
		rank := vertexCount - 1 - position
		ranks[vertex], byRank[rank] = rank, vertex
	}
	alpha, err := initialize(ranks)
	if err != nil {
		return nil, nil, err
	}
	inverse, err := initialize(byRank)
	if err != nil {
		alpha.close()
		return nil, nil, err
	}
	return alpha, inverse, nil
}

// Chordality determines whether the graph is chordal and returns a chordal
// fill-in. Directed edges are treated as undirected. Ordering is borrowed only
// for the call; all returned values are independently owned.
//
//igraph:bind igraph_is_chordal
func (g *Graph) Chordality(options ChordalityOptions) (ChordalityResult, error) {
	if g == nil {
		return ChordalityResult{}, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return ChordalityResult{}, ErrClosed
	}
	var alpha, inverse *intVector
	var err error
	if options.Ordering != nil {
		alpha, inverse, err = orderingVectors(options.Ordering, int(C.igraph_vcount(&g.graph)))
		if err != nil {
			return ChordalityResult{}, err
		}
		defer alpha.close()
		defer inverse.close()
	}
	fill, err := newIntVector(nil)
	if err != nil {
		return ChordalityResult{}, err
	}
	defer fill.close()
	var alphaPtr, inversePtr *C.igraph_vector_int_t
	if alpha != nil {
		alphaPtr, inversePtr = &alpha.value, &inverse.value
	}
	var chordal C.igraph_bool_t
	var completed C.igraph_t
	var completedPtr *C.igraph_t
	if options.Complete {
		completedPtr = &completed
	}
	if code := C.go_igraph_is_chordal(&g.graph, alphaPtr, inversePtr, &chordal, &fill.value, completedPtr); code != C.IGRAPH_SUCCESS {
		return ChordalityResult{}, igraphError("analyze graph chordality", int(code))
	}
	endpoints, err := fill.slice()
	if err != nil {
		if completedPtr != nil {
			C.igraph_destroy(&completed)
		}
		return ChordalityResult{}, err
	}
	if len(endpoints)%2 != 0 {
		if completedPtr != nil {
			C.igraph_destroy(&completed)
		}
		return ChordalityResult{}, fmt.Errorf("igraph: chordal fill-in returned odd endpoint count %d", len(endpoints))
	}
	edges := make([]Edge, len(endpoints)/2)
	for i := range edges {
		from, to := endpoints[2*i], endpoints[2*i+1]
		if from > to {
			from, to = to, from
		}
		edges[i] = Edge{From: from, To: to}
	}
	result := ChordalityResult{Chordal: chordal != booltoint(false), FillEdges: edges}
	if completedPtr != nil {
		result.Completion = adoptInitializedGraph(&completed)
	}
	return result, nil
}

// IsPerfect reports whether an undirected simple graph is perfect. Directed,
// looped and parallel-edge graphs are rejected before the perfectness test.
//
//igraph:bind igraph_is_perfect
func (g *Graph) IsPerfect() (bool, error) {
	return g.isPerfect(nil)
}

type perfectGraphAdapters struct {
	isSimple  func(*Graph) (bool, int)
	isPerfect func(*Graph) (bool, int)
}

func defaultPerfectGraphAdapters() perfectGraphAdapters {
	return perfectGraphAdapters{
		isSimple: func(graph *Graph) (bool, int) {
			var result C.igraph_bool_t
			code := C.go_igraph_is_simple_for_perfect(&graph.graph, &result)
			return result != booltoint(false), int(code)
		},
		isPerfect: func(graph *Graph) (bool, int) {
			var result C.igraph_bool_t
			code := C.go_igraph_is_perfect(&graph.graph, &result)
			return result != booltoint(false), int(code)
		},
	}
}

func resolvedPerfectGraphAdapters(adapters *perfectGraphAdapters) perfectGraphAdapters {
	if adapters == nil {
		return defaultPerfectGraphAdapters()
	}
	return *adapters
}

func (g *Graph) isPerfect(adapters *perfectGraphAdapters) (bool, error) {
	if g == nil {
		return false, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return false, ErrClosed
	}
	if C.igraph_is_directed(&g.graph) != booltoint(false) {
		return false, fmt.Errorf("igraph: perfect graph analysis requires an undirected graph")
	}
	resolved := resolvedPerfectGraphAdapters(adapters)
	simple, code := resolved.isSimple(g)
	if code != int(C.IGRAPH_SUCCESS) {
		return false, igraphError("validate perfect graph input", code)
	}
	if !simple {
		return false, fmt.Errorf("igraph: perfect graph analysis requires a simple graph")
	}
	perfect, code := resolved.isPerfect(g)
	if code != int(C.IGRAPH_SUCCESS) {
		return false, igraphError("test whether graph is perfect", code)
	}
	return perfect, nil
}

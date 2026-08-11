package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "cycle_cgo.h"
import "C"

import (
	"errors"
	"fmt"
)

// Cycle is a Go-owned cycle witness. Vertices and Edges have equal lengths and
// are aligned in traversal order: Edges[i] connects Vertices[i] to the next
// vertex, wrapping at the end. Both slices are non-nil, including when no
// cycle was found, and remain valid and mutable after the graph is closed.
type Cycle struct {
	Vertices []int
	Edges    []int
}

// GirthResult contains the length and vertex witness of a shortest cycle.
// Vertices is Go-owned and non-nil. For an acyclic graph, Length is positive
// infinity and Vertices is empty.
type GirthResult struct {
	Length   float64
	Vertices []int
}

// IsAcyclic reports whether the graph has no cycles. Directed graphs are
// checked for directed cycles; undirected graphs are checked for undirected
// cycles. Self-loops are cycles in both cases.
//
//igraph:bind igraph_is_acyclic
func (g *Graph) IsAcyclic() (bool, error) {
	return g.cyclePredicate("check acyclicity", cyclePredicateAcyclic)
}

// IsDAG reports whether the graph is a directed acyclic graph. It returns
// false, without error, for every undirected graph. A directed self-loop is a
// cycle.
//
//igraph:bind igraph_is_dag
func (g *Graph) IsDAG() (bool, error) {
	return g.cyclePredicate("check directed acyclicity", cyclePredicateDAG)
}

type cyclePredicateKind uint8

const (
	cyclePredicateAcyclic cyclePredicateKind = iota
	cyclePredicateDAG
)

func (g *Graph) cyclePredicate(action string, kind cyclePredicateKind) (bool, error) {
	if g == nil {
		return false, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return false, ErrClosed
	}
	var result C.igraph_bool_t
	var code C.igraph_error_t
	switch kind {
	case cyclePredicateAcyclic:
		code = C.go_igraph_is_acyclic(&g.graph, &result)
	case cyclePredicateDAG:
		code = C.go_igraph_is_dag(&g.graph, &result)
	default:
		return false, errors.New("igraph: invalid cycle predicate")
	}
	if code != C.IGRAPH_SUCCESS {
		return false, igraphError(action, int(code))
	}
	return result != booltoint(false), nil
}

// TopologicalSort returns one topological ordering of a directed graph. Mode
// must be DirectionOut or DirectionIn; the latter returns the reverse
// orientation. The graph and mode are borrowed only for the synchronous call,
// and the returned slice is Go-owned. No stable ordering is promised between
// otherwise valid choices.
//
// Undirected graphs, DirectionAll, and directed graphs with a non-loop cycle
// return errors. Pinned igraph 1.0.1 ignores self-loops here, so a graph whose
// only cycles are self-loops can be sorted even though IsAcyclic and IsDAG
// return false.
//
//igraph:bind igraph_topological_sorting
func (g *Graph) TopologicalSort(mode DirectionMode) ([]int, error) {
	return g.topologicalSort(mode, nil)
}

// FindCycle returns one cycle reachable under mode. DirectionOut,
// DirectionIn, and DirectionAll are accepted; direction is ignored for an
// undirected graph. If no cycle exists, both result slices are non-nil and
// empty. The graph and mode are borrowed only for the synchronous call.
//
//igraph:bind igraph_find_cycle
func (g *Graph) FindCycle(mode DirectionMode) (Cycle, error) {
	return g.findCycle(mode, nil)
}

// Girth returns the length and vertex witness of a shortest cycle. Pinned
// igraph 1.0.1 ignores edge direction, self-loops, and cycles formed by two
// parallel edges. An acyclic graph therefore returns positive infinity and a
// non-nil empty witness. The returned slice is Go-owned and remains valid
// after graph closure.
//
//igraph:bind igraph_girth
func (g *Graph) Girth() (GirthResult, error) {
	return g.girth(nil)
}

type cycleVectorInitializer func() (*intVector, error)
type cycleVectorCloser func(*intVector)
type cycleVectorConverter func(*intVector) ([]int, error)

func defaultCycleVectorInitializer() (*intVector, error) { return newIntVector(nil) }
func defaultCycleVectorCloser(vector *intVector)         { vector.close() }
func defaultCycleVectorConverter(vector *intVector) ([]int, error) {
	return vector.slice()
}

func newCycleVectorPair(
	initialize cycleVectorInitializer,
	closeVector cycleVectorCloser,
) (*intVector, *intVector, error) {
	vertices, err := initialize()
	if err != nil {
		return nil, nil, err
	}
	edges, err := initialize()
	if err != nil {
		closeVector(vertices)
		return nil, nil, err
	}
	return vertices, edges, nil
}

type topologicalSortAdapters struct {
	initialize cycleVectorInitializer
	close      cycleVectorCloser
	call       func(*Graph, *intVector, DirectionMode) int
	convert    cycleVectorConverter
}

func defaultTopologicalSortAdapters() topologicalSortAdapters {
	return topologicalSortAdapters{
		initialize: defaultCycleVectorInitializer,
		close:      defaultCycleVectorCloser,
		call: func(g *Graph, result *intVector, mode DirectionMode) int {
			cMode, _ := mode.cValue()
			return int(C.go_igraph_topological_sorting(&g.graph, &result.value, cMode))
		},
		convert: defaultCycleVectorConverter,
	}
}

func (g *Graph) topologicalSort(mode DirectionMode, adapters *topologicalSortAdapters) ([]int, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	if _, err := topologicalDirection(mode); err != nil {
		return nil, err
	}
	if C.igraph_is_directed(&g.graph) == booltoint(false) {
		return nil, errors.New("igraph: topological sorting requires a directed graph")
	}
	resolved := defaultTopologicalSortAdapters()
	if adapters != nil {
		resolved = *adapters
	}
	result, err := resolved.initialize()
	if err != nil {
		return nil, err
	}
	defer resolved.close(result)
	if code := resolved.call(g, result, mode); code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("topologically sort graph", code)
	}
	return resolved.convert(result)
}

func topologicalDirection(mode DirectionMode) (C.igraph_neimode_t, error) {
	cMode, err := mode.cValue()
	if err != nil {
		return 0, err
	}
	if mode == DirectionAll {
		return 0, errors.New("igraph: topological sorting does not accept DirectionAll")
	}
	return cMode, nil
}

type findCycleAdapters struct {
	initialize cycleVectorInitializer
	close      cycleVectorCloser
	call       func(*Graph, *intVector, *intVector, DirectionMode) int
	convert    cycleVectorConverter
}

func defaultFindCycleAdapters() findCycleAdapters {
	return findCycleAdapters{
		initialize: defaultCycleVectorInitializer,
		close:      defaultCycleVectorCloser,
		call: func(g *Graph, vertices, edges *intVector, mode DirectionMode) int {
			cMode, _ := mode.cValue()
			return int(C.go_igraph_find_cycle(&g.graph, &vertices.value, &edges.value, cMode))
		},
		convert: defaultCycleVectorConverter,
	}
}

func (g *Graph) findCycle(mode DirectionMode, adapters *findCycleAdapters) (Cycle, error) {
	if g == nil {
		return Cycle{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Cycle{}, ErrClosed
	}
	if _, err := mode.cValue(); err != nil {
		return Cycle{}, err
	}
	resolved := defaultFindCycleAdapters()
	if adapters != nil {
		resolved = *adapters
	}
	vertices, edges, err := newCycleVectorPair(resolved.initialize, resolved.close)
	if err != nil {
		return Cycle{}, err
	}
	defer resolved.close(vertices)
	defer resolved.close(edges)
	if code := resolved.call(g, vertices, edges, mode); code != int(C.IGRAPH_SUCCESS) {
		return Cycle{}, igraphError("find cycle", code)
	}
	vertexIDs, err := resolved.convert(vertices)
	if err != nil {
		return Cycle{}, err
	}
	edgeIDs, err := resolved.convert(edges)
	if err != nil {
		return Cycle{}, err
	}
	return newPredecessorAlignedCycle(vertexIDs, edgeIDs)
}

func newPredecessorAlignedCycle(vertexIDs, edgeIDs []int) (Cycle, error) {
	if len(vertexIDs) != len(edgeIDs) {
		return Cycle{}, fmt.Errorf("igraph: cycle vertex length %d does not match edge length %d", len(vertexIDs), len(edgeIDs))
	}
	// Pinned igraph returns each edge immediately before the corresponding
	// vertex. Rotate the edge IDs so the public Cycle contract instead stores
	// the edge leaving Vertices[i] at Edges[i].
	if len(edgeIDs) > 1 {
		first := edgeIDs[0]
		copy(edgeIDs, edgeIDs[1:])
		edgeIDs[len(edgeIDs)-1] = first
	}
	return newCycle(vertexIDs, edgeIDs)
}

func newCycle(vertexIDs, edgeIDs []int) (Cycle, error) {
	if len(vertexIDs) != len(edgeIDs) {
		return Cycle{}, fmt.Errorf("igraph: cycle vertex length %d does not match edge length %d", len(vertexIDs), len(edgeIDs))
	}
	return Cycle{Vertices: vertexIDs, Edges: edgeIDs}, nil
}

type girthAdapters struct {
	initialize cycleVectorInitializer
	close      cycleVectorCloser
	call       func(*Graph, *intVector) (float64, int)
	convert    cycleVectorConverter
}

func defaultGirthAdapters() girthAdapters {
	return girthAdapters{
		initialize: defaultCycleVectorInitializer,
		close:      defaultCycleVectorCloser,
		call: func(g *Graph, vertices *intVector) (float64, int) {
			var length C.igraph_real_t
			code := C.go_igraph_girth(&g.graph, &length, &vertices.value)
			return float64(length), int(code)
		},
		convert: defaultCycleVectorConverter,
	}
}

func (g *Graph) girth(adapters *girthAdapters) (GirthResult, error) {
	if g == nil {
		return GirthResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return GirthResult{}, ErrClosed
	}
	resolved := defaultGirthAdapters()
	if adapters != nil {
		resolved = *adapters
	}
	vertices, err := resolved.initialize()
	if err != nil {
		return GirthResult{}, err
	}
	defer resolved.close(vertices)
	length, code := resolved.call(g, vertices)
	if code != int(C.IGRAPH_SUCCESS) {
		return GirthResult{}, igraphError("calculate girth", code)
	}
	vertexIDs, err := resolved.convert(vertices)
	if err != nil {
		return GirthResult{}, err
	}
	return GirthResult{Length: length, Vertices: vertexIDs}, nil
}

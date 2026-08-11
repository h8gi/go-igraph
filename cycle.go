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

type cycleAnalysisStage uint8

const (
	cycleAnalysisAfterFirstVectorInit cycleAnalysisStage = iota
	cycleAnalysisAfterSecondVectorInit
	cycleAnalysisBeforeUpstream
	cycleAnalysisBeforeFirstConversion
	cycleAnalysisBeforeSecondConversion
)

type cycleAnalysisFailureHook func(cycleAnalysisStage) error

func runCycleAnalysisHook(hook cycleAnalysisFailureHook, stage cycleAnalysisStage) error {
	if hook == nil {
		return nil
	}
	if err := hook(stage); err != nil {
		return fmt.Errorf("igraph: injected cycle-analysis failure at stage %d: %w", stage, err)
	}
	return nil
}

func (g *Graph) topologicalSort(mode DirectionMode, hook cycleAnalysisFailureHook) ([]int, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	cMode, err := topologicalDirection(mode)
	if err != nil {
		return nil, err
	}
	if C.igraph_is_directed(&g.graph) == booltoint(false) {
		return nil, errors.New("igraph: topological sorting requires a directed graph")
	}
	result, err := newIntVector(nil)
	if err != nil {
		return nil, err
	}
	defer result.close()
	if err := runCycleAnalysisHook(hook, cycleAnalysisAfterFirstVectorInit); err != nil {
		return nil, err
	}
	if err := runCycleAnalysisHook(hook, cycleAnalysisBeforeUpstream); err != nil {
		return nil, err
	}
	if code := C.go_igraph_topological_sorting(&g.graph, &result.value, cMode); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("topologically sort graph", int(code))
	}
	if err := runCycleAnalysisHook(hook, cycleAnalysisBeforeFirstConversion); err != nil {
		return nil, err
	}
	return result.slice()
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

func (g *Graph) findCycle(mode DirectionMode, hook cycleAnalysisFailureHook) (Cycle, error) {
	if g == nil {
		return Cycle{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Cycle{}, ErrClosed
	}
	cMode, err := mode.cValue()
	if err != nil {
		return Cycle{}, err
	}
	vertices, err := newIntVector(nil)
	if err != nil {
		return Cycle{}, err
	}
	defer vertices.close()
	if err := runCycleAnalysisHook(hook, cycleAnalysisAfterFirstVectorInit); err != nil {
		return Cycle{}, err
	}
	edges, err := newIntVector(nil)
	if err != nil {
		return Cycle{}, err
	}
	defer edges.close()
	if err := runCycleAnalysisHook(hook, cycleAnalysisAfterSecondVectorInit); err != nil {
		return Cycle{}, err
	}
	if err := runCycleAnalysisHook(hook, cycleAnalysisBeforeUpstream); err != nil {
		return Cycle{}, err
	}
	if code := C.go_igraph_find_cycle(&g.graph, &vertices.value, &edges.value, cMode); code != C.IGRAPH_SUCCESS {
		return Cycle{}, igraphError("find cycle", int(code))
	}
	if err := runCycleAnalysisHook(hook, cycleAnalysisBeforeFirstConversion); err != nil {
		return Cycle{}, err
	}
	vertexIDs, err := vertices.slice()
	if err != nil {
		return Cycle{}, err
	}
	if err := runCycleAnalysisHook(hook, cycleAnalysisBeforeSecondConversion); err != nil {
		return Cycle{}, err
	}
	edgeIDs, err := edges.slice()
	if err != nil {
		return Cycle{}, err
	}
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
	return Cycle{Vertices: vertexIDs, Edges: edgeIDs}, nil
}

func (g *Graph) girth(hook cycleAnalysisFailureHook) (GirthResult, error) {
	if g == nil {
		return GirthResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return GirthResult{}, ErrClosed
	}
	vertices, err := newIntVector(nil)
	if err != nil {
		return GirthResult{}, err
	}
	defer vertices.close()
	if err := runCycleAnalysisHook(hook, cycleAnalysisAfterFirstVectorInit); err != nil {
		return GirthResult{}, err
	}
	if err := runCycleAnalysisHook(hook, cycleAnalysisBeforeUpstream); err != nil {
		return GirthResult{}, err
	}
	var length C.igraph_real_t
	if code := C.go_igraph_girth(&g.graph, &length, &vertices.value); code != C.IGRAPH_SUCCESS {
		return GirthResult{}, igraphError("calculate girth", int(code))
	}
	if err := runCycleAnalysisHook(hook, cycleAnalysisBeforeFirstConversion); err != nil {
		return GirthResult{}, err
	}
	vertexIDs, err := vertices.slice()
	if err != nil {
		return GirthResult{}, err
	}
	return GirthResult{Length: float64(length), Vertices: vertexIDs}, nil
}

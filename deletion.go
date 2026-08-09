package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
import "C"

import (
	"fmt"
	"reflect"
)

type deletionStage uint8

const (
	deletionBeforeClone deletionStage = iota
	deletionBeforeEdgeSnapshot
	deletionBeforeSelectorInit
	deletionBeforeFirstMappingInit
	deletionBeforeSecondMappingInit
	deletionAtMutation
	deletionAfterMutation
)

type deletionFailureHook func(stage deletionStage) error

// DeleteEdges removes the selected edges and returns the vertex and edge ID
// mappings from the graph before deletion to the graph after deletion. The
// selector is borrowed and fully materialized while the graph lock is held,
// before any mutation begins. Selector order does not affect the result and
// duplicate selections remove an edge only once. An empty selector is a no-op.
//
// The returned mapping is Go-owned and remains valid after the graph is
// closed. Vertex IDs are unchanged. Retained edges keep their relative order;
// removed edges map to RemovedID. The mutation is atomic: validation,
// initialization, upstream, and result-conversion failures leave the original
// graph unchanged.
//
//igraph:bind igraph_delete_edges
func (g *Graph) DeleteEdges(edges EdgeSelector) (GraphIDMapping, error) {
	return g.deleteEdges(edges, nil)
}

func (g *Graph) deleteEdges(
	edges EdgeSelector,
	hook deletionFailureHook,
) (GraphIDMapping, error) {
	if g == nil {
		return GraphIDMapping{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return GraphIDMapping{}, ErrClosed
	}

	selectedIDs, err := materializeSelectedEdgeIDs(&g.graph, edges)
	if err != nil {
		return GraphIDMapping{}, err
	}
	vertexCount := int(C.igraph_vcount(&g.graph))
	edgeCount := int(C.igraph_ecount(&g.graph))
	vertexMapping, err := identityIDMapping(vertexCount)
	if err != nil {
		return GraphIDMapping{}, err
	}
	edgeMapping, err := deletionIDMapping(edgeCount, selectedIDs)
	if err != nil {
		return GraphIDMapping{}, err
	}
	result := GraphIDMapping{Vertices: vertexMapping, Edges: edgeMapping}
	if len(selectedIDs) == 0 {
		return result, nil
	}

	var replacement C.igraph_t
	if err := runDeletionHook(hook, deletionBeforeClone); err != nil {
		return GraphIDMapping{}, err
	}
	if code := C.go_igraph_copy(&replacement, &g.graph); code != C.IGRAPH_SUCCESS {
		return GraphIDMapping{}, igraphError("copy graph for edge deletion", int(code))
	}
	committed := false
	defer func() {
		if !committed {
			C.igraph_destroy(&replacement)
		}
	}()

	if err := runDeletionHook(hook, deletionBeforeSelectorInit); err != nil {
		return GraphIDMapping{}, err
	}
	selector := EdgeSelector{kind: edgeSelectorIDs, ids: selectedIDs}
	cSelector, err := newCEdgeSelector(selector)
	if err != nil {
		return GraphIDMapping{}, err
	}
	defer cSelector.close()
	// A failure injected at deletionAtMutation models an upstream wrapper
	// returning an error without committing the replacement graph.
	if err := runDeletionHook(hook, deletionAtMutation); err != nil {
		return GraphIDMapping{}, err
	}
	if code := C.go_igraph_delete_edges(&replacement, cSelector.value); code != C.IGRAPH_SUCCESS {
		return GraphIDMapping{}, igraphError("delete edges", int(code))
	}
	if err := runDeletionHook(hook, deletionAfterMutation); err != nil {
		return GraphIDMapping{}, err
	}
	if err := validateDeletionCount("edge", int(C.igraph_ecount(&replacement)), len(edgeMapping.NewToOld)); err != nil {
		return GraphIDMapping{}, err
	}

	replaceInitializedGraph(&g.graph, &replacement)
	committed = true
	return result, nil
}

// DeleteVertices removes the selected vertices and all incident edges. It
// returns vertex and edge mappings from the graph before deletion to the graph
// after deletion. The selector borrowing, materialization, duplicate, empty,
// ownership, and atomicity rules match DeleteEdges.
//
// Remaining vertices and edges keep their relative order. Every removed
// vertex and every edge incident to a removed vertex maps to RemovedID. Loops,
// parallel edges, and edge direction do not change these rules.
//
//igraph:bind igraph_delete_vertices_map
func (g *Graph) DeleteVertices(vertices VertexSelector) (GraphIDMapping, error) {
	return g.deleteVertices(vertices, nil)
}

func (g *Graph) deleteVertices(
	vertices VertexSelector,
	hook deletionFailureHook,
) (GraphIDMapping, error) {
	if g == nil {
		return GraphIDMapping{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return GraphIDMapping{}, ErrClosed
	}

	vertexCount := int(C.igraph_vcount(&g.graph))
	if err := validateVertexSelector(vertices, vertexCount); err != nil {
		return GraphIDMapping{}, err
	}
	selectedIDs, err := materializeVertexIDs(&g.graph, vertices)
	if err != nil {
		return GraphIDMapping{}, fmt.Errorf("igraph: materialize vertex deletion selector: %w", err)
	}
	if len(selectedIDs) == 0 {
		verticesIdentity, err := identityIDMapping(vertexCount)
		if err != nil {
			return GraphIDMapping{}, err
		}
		edgesIdentity, err := identityIDMapping(int(C.igraph_ecount(&g.graph)))
		if err != nil {
			return GraphIDMapping{}, err
		}
		return GraphIDMapping{Vertices: verticesIdentity, Edges: edgesIdentity}, nil
	}
	if err := runDeletionHook(hook, deletionBeforeEdgeSnapshot); err != nil {
		return GraphIDMapping{}, err
	}
	edgesBefore, err := edgeSlice(&g.graph)
	if err != nil {
		return GraphIDMapping{}, err
	}

	var replacement C.igraph_t
	if err := runDeletionHook(hook, deletionBeforeClone); err != nil {
		return GraphIDMapping{}, err
	}
	if code := C.go_igraph_copy(&replacement, &g.graph); code != C.IGRAPH_SUCCESS {
		return GraphIDMapping{}, igraphError("copy graph for vertex deletion", int(code))
	}
	committed := false
	defer func() {
		if !committed {
			C.igraph_destroy(&replacement)
		}
	}()

	if err := runDeletionHook(hook, deletionBeforeSelectorInit); err != nil {
		return GraphIDMapping{}, err
	}
	selector := VertexSelector{kind: vertexSelectorIDs, ids: selectedIDs}
	cSelector, err := newCVertexSelector(selector)
	if err != nil {
		return GraphIDMapping{}, err
	}
	defer cSelector.close()

	if err := runDeletionHook(hook, deletionBeforeFirstMappingInit); err != nil {
		return GraphIDMapping{}, err
	}
	oldToNew, err := newIntVector(nil)
	if err != nil {
		return GraphIDMapping{}, err
	}
	defer oldToNew.close()
	if err := runDeletionHook(hook, deletionBeforeSecondMappingInit); err != nil {
		return GraphIDMapping{}, err
	}
	newToOld, err := newIntVector(nil)
	if err != nil {
		return GraphIDMapping{}, err
	}
	defer newToOld.close()

	// A failure injected at deletionAtMutation models an upstream wrapper
	// returning an error without committing the replacement graph.
	if err := runDeletionHook(hook, deletionAtMutation); err != nil {
		return GraphIDMapping{}, err
	}
	if code := C.go_igraph_delete_vertices_map(
		&replacement, cSelector.value, &oldToNew.value, &newToOld.value,
	); code != C.IGRAPH_SUCCESS {
		return GraphIDMapping{}, igraphError("delete vertices", int(code))
	}
	if err := runDeletionHook(hook, deletionAfterMutation); err != nil {
		return GraphIDMapping{}, err
	}

	oldToNewSlice, err := oldToNew.slice()
	if err != nil {
		return GraphIDMapping{}, err
	}
	newToOldSlice, err := newToOld.slice()
	if err != nil {
		return GraphIDMapping{}, err
	}
	vertexMapping, err := newIDMapping(oldToNewSlice, len(newToOldSlice))
	if err != nil {
		return GraphIDMapping{}, fmt.Errorf("igraph: convert vertex deletion mapping: %w", err)
	}
	if err := validateReverseDeletionMapping(vertexMapping.NewToOld, newToOldSlice); err != nil {
		return GraphIDMapping{}, err
	}
	edgeMapping, err := vertexDeletionEdgeMapping(edgesBefore, vertexMapping)
	if err != nil {
		return GraphIDMapping{}, err
	}
	if err := validateDeletionCount("edge after vertex", int(C.igraph_ecount(&replacement)), len(edgeMapping.NewToOld)); err != nil {
		return GraphIDMapping{}, err
	}

	replaceInitializedGraph(&g.graph, &replacement)
	committed = true
	return GraphIDMapping{Vertices: vertexMapping, Edges: edgeMapping}, nil
}

func deletionIDMapping(count int, deletedIDs []int) (IDMapping, error) {
	if count < 0 {
		return IDMapping{}, fmt.Errorf("igraph: deletion ID count must be non-negative: %d", count)
	}
	deleted := make([]bool, count)
	for _, id := range deletedIDs {
		if id < 0 || id >= count {
			return IDMapping{}, fmt.Errorf("igraph: deleted ID %d out of range [0, %d)", id, count)
		}
		deleted[id] = true
	}
	oldToNew := make([]int, count)
	newID := 0
	for oldID := range oldToNew {
		if deleted[oldID] {
			oldToNew[oldID] = RemovedID
			continue
		}
		oldToNew[oldID] = newID
		newID++
	}
	return newIDMapping(oldToNew, newID)
}

func vertexDeletionEdgeMapping(edges []Edge, vertices IDMapping) (IDMapping, error) {
	deletedEdges := make([]int, 0)
	for edgeID, edge := range edges {
		if edge.From < 0 || edge.From >= len(vertices.OldToNew) {
			return IDMapping{}, fmt.Errorf("igraph: edge %d source %d has no vertex mapping", edgeID, edge.From)
		}
		if edge.To < 0 || edge.To >= len(vertices.OldToNew) {
			return IDMapping{}, fmt.Errorf("igraph: edge %d target %d has no vertex mapping", edgeID, edge.To)
		}
		if vertices.OldToNew[edge.From] == RemovedID || vertices.OldToNew[edge.To] == RemovedID {
			deletedEdges = append(deletedEdges, edgeID)
		}
	}
	return deletionIDMapping(len(edges), deletedEdges)
}

func edgeSlice(graph *C.igraph_t) ([]Edge, error) {
	if graph == nil {
		return nil, fmt.Errorf("igraph: graph is nil")
	}
	result, err := newIntVector(nil)
	if err != nil {
		return nil, err
	}
	defer result.close()
	if code := C.go_igraph_get_edgelist(graph, &result.value, booltoint(false)); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("get edge list before deletion", int(code))
	}
	values, err := result.slice()
	if err != nil {
		return nil, err
	}
	edges := make([]Edge, len(values)/2)
	for index := range edges {
		edges[index] = Edge{From: values[2*index], To: values[2*index+1]}
	}
	return edges, nil
}

func validateDeletionCount(element string, actual, mapped int) error {
	if actual != mapped {
		return fmt.Errorf("igraph: %s count after deletion is %d, mapping describes %d", element, actual, mapped)
	}
	return nil
}

func validateReverseDeletionMapping(expected, actual []int) error {
	if !reflect.DeepEqual(expected, actual) {
		return fmt.Errorf("igraph: inconsistent reverse vertex mapping: got %v, want %v", actual, expected)
	}
	return nil
}

func replaceInitializedGraph(destination, replacement *C.igraph_t) {
	previous := *destination
	*destination = *replacement
	*replacement = C.igraph_t{}
	C.igraph_destroy(&previous)
}

func runDeletionHook(hook deletionFailureHook, stage deletionStage) error {
	if hook == nil {
		return nil
	}
	if err := hook(stage); err != nil {
		return fmt.Errorf("igraph: injected deletion failure at stage %d: %w", stage, err)
	}
	return nil
}

package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "subgraph_cgo.h"
import "C"

import (
	"errors"
	"fmt"
	"sort"
)

// InducedSubgraphResult contains an independently owned graph and the exact
// vertex-ID mapping returned by upstream igraph. No edge mapping is exposed
// because upstream does not return result edge-ID correspondence or a result
// edge-order mapping.
//
// Graph remains usable after the source graph is closed. Vertices and all of
// its slices are non-nil, Go-owned values and remain valid after either graph
// is closed.
type InducedSubgraphResult struct {
	Graph    *Graph
	Vertices IDMapping
}

// EdgeSubgraphResult contains an independently owned graph and its vertex-ID
// mapping. No edge mapping is exposed because upstream igraph does not return
// result edge-ID correspondence or a result edge-order mapping.
//
// Graph remains usable after the source graph is closed. Vertices and all of
// its slices are non-nil, Go-owned values and remain valid after either graph
// is closed.
type EdgeSubgraphResult struct {
	Graph    *Graph
	Vertices IDMapping
}

// InducedSubgraph returns the subgraph induced by vertices and every edge whose
// endpoints are both selected. The selector is borrowed and fully materialized
// while the source graph is locked; it is never retained. Duplicate vertex IDs
// are considered once, selector order is ignored, and retained vertices appear
// in increasing source-ID order, as defined by upstream igraph.
//
// The returned graph is independently owned and must be closed by the caller.
// An empty selection returns a valid empty graph and a non-nil mapping.
//
//igraph:bind igraph_induced_subgraph_map
func (g *Graph) InducedSubgraph(vertices VertexSelector) (InducedSubgraphResult, error) {
	if g == nil {
		return InducedSubgraphResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return InducedSubgraphResult{}, ErrClosed
	}

	oldVertexCount, err := igraphIntToInt(C.igraph_vcount(&g.graph), "source vertex count")
	if err != nil {
		return InducedSubgraphResult{}, err
	}
	if err := validateVertexSelector(vertices, oldVertexCount); err != nil {
		return InducedSubgraphResult{}, err
	}
	selectedIDs, err := materializeVertexIDs(&g.graph, vertices)
	if err != nil {
		return InducedSubgraphResult{}, fmt.Errorf("igraph: materialize induced-subgraph selector: %w", err)
	}
	materialized, err := VertexIDs(selectedIDs...)
	if err != nil {
		return InducedSubgraphResult{}, err
	}
	selector, err := newCVertexSelector(materialized)
	if err != nil {
		return InducedSubgraphResult{}, err
	}
	defer selector.close()

	return collectInducedSubgraph(oldVertexCount, inducedSubgraphOperations{
		newVector:   func() (*intVector, error) { return newIntVector(nil) },
		closeVector: (*intVector).close,
		query: func(mapping, inverse *intVector) (*Graph, int, error) {
			var value C.igraph_t
			if code := C.go_igraph_induced_subgraph_map(
				&g.graph, &value, selector.value, &mapping.value, &inverse.value,
			); code != C.IGRAPH_SUCCESS {
				return nil, 0, igraphError("create induced subgraph", int(code))
			}
			newVertexCount, err := igraphIntToInt(C.igraph_vcount(&value), "induced-subgraph vertex count")
			if err != nil {
				C.igraph_destroy(&value)
				return nil, 0, err
			}
			return adoptInitializedGraph(&value), newVertexCount, nil
		},
		vectorSlice: func(vector *intVector) ([]int, error) { return vector.slice() },
		closeGraph:  func(graph *Graph) { _ = graph.Close() },
	})
}

type inducedSubgraphOperations struct {
	newVector   func() (*intVector, error)
	closeVector func(*intVector)
	query       func(*intVector, *intVector) (*Graph, int, error)
	vectorSlice func(*intVector) ([]int, error)
	closeGraph  func(*Graph)
}

func collectInducedSubgraph(oldVertexCount int, operations inducedSubgraphOperations) (result InducedSubgraphResult, err error) {
	mappingVector, err := operations.newVector()
	if err != nil {
		return InducedSubgraphResult{}, err
	}
	defer operations.closeVector(mappingVector)
	inverseVector, err := operations.newVector()
	if err != nil {
		return InducedSubgraphResult{}, err
	}
	defer operations.closeVector(inverseVector)

	graph, newVertexCount, err := operations.query(mappingVector, inverseVector)
	if err != nil {
		if graph != nil {
			operations.closeGraph(graph)
		}
		return InducedSubgraphResult{}, err
	}
	if graph == nil {
		return InducedSubgraphResult{}, errors.New("igraph: induced-subgraph query returned a nil graph")
	}
	succeeded := false
	defer func() {
		if !succeeded {
			operations.closeGraph(graph)
		}
	}()

	oldToNew, err := operations.vectorSlice(mappingVector)
	if err != nil {
		return InducedSubgraphResult{}, err
	}
	newToOld, err := operations.vectorSlice(inverseVector)
	if err != nil {
		return InducedSubgraphResult{}, err
	}
	if len(oldToNew) != oldVertexCount {
		return InducedSubgraphResult{}, fmt.Errorf(
			"igraph: induced-subgraph mapping length %d does not match source vertex count %d",
			len(oldToNew), oldVertexCount,
		)
	}
	mapping, err := newIDMapping(oldToNew, newVertexCount)
	if err != nil {
		return InducedSubgraphResult{}, err
	}
	if !equalIntSlices(mapping.NewToOld, newToOld) {
		return InducedSubgraphResult{}, errors.New("igraph: induced-subgraph inverse mapping is inconsistent")
	}

	result = InducedSubgraphResult{Graph: graph, Vertices: mapping}
	succeeded = true
	return result, nil
}

// EdgeSubgraph returns a graph containing the selected edges. The selector is
// borrowed and fully materialized while the source graph is locked, then
// normalized so duplicate edge IDs are retained once and selector order is
// ignored. Edges are supplied to upstream in increasing source-ID order.
//
// If deleteIsolatedVertices is false, every source vertex is retained with the
// same ID. If true, vertices not incident on a retained edge are removed and
// the exact vertex mapping returned by upstream vertex deletion is exposed.
// The edge subgraph is first created with every source vertex so that its
// initial vertex IDs are identical to the source IDs; isolated vertices are
// then deleted from that independent result. The returned graph is
// independently owned and must be closed by the caller. No edge mapping is
// returned because upstream exposes neither result edge-ID correspondence nor
// a result edge-order mapping.
//
//igraph:bind igraph_subgraph_from_edges
func (g *Graph) EdgeSubgraph(edges EdgeSelector, deleteIsolatedVertices bool) (EdgeSubgraphResult, error) {
	if g == nil {
		return EdgeSubgraphResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return EdgeSubgraphResult{}, ErrClosed
	}

	selectedIDs, err := materializeSelectedEdgeIDs(&g.graph, edges)
	if err != nil {
		return EdgeSubgraphResult{}, fmt.Errorf("igraph: materialize edge-subgraph selector: %w", err)
	}
	selectedIDs = sortedUniqueIDs(selectedIDs)
	vertexCount, err := igraphIntToInt(C.igraph_vcount(&g.graph), "source vertex count")
	if err != nil {
		return EdgeSubgraphResult{}, err
	}
	materialized, err := EdgeIDs(selectedIDs...)
	if err != nil {
		return EdgeSubgraphResult{}, err
	}
	selector, err := newCEdgeSelector(materialized)
	if err != nil {
		return EdgeSubgraphResult{}, err
	}
	defer selector.close()

	return collectEdgeSubgraph(vertexCount, deleteIsolatedVertices, edgeSubgraphOperations{
		query: func() (*Graph, error) {
			var value C.igraph_t
			if code := C.go_igraph_subgraph_from_edges(
				&g.graph, &value, selector.value, booltoint(false),
			); code != C.IGRAPH_SUCCESS {
				return nil, igraphError("create edge subgraph", int(code))
			}
			return adoptInitializedGraph(&value), nil
		},
		identityMapping:  identityIDMapping,
		isolatedVertices: edgeSubgraphIsolatedVertices,
		deleteVertices: func(graph *Graph, vertices VertexSelector) (GraphIDMapping, error) {
			return graph.DeleteVertices(vertices)
		},
		vertexCount: (*Graph).VertexCount,
		closeGraph:  func(graph *Graph) { _ = graph.Close() },
	})
}

type edgeSubgraphOperations struct {
	query            func() (*Graph, error)
	identityMapping  func(int) (IDMapping, error)
	isolatedVertices func(*Graph) (VertexSelector, error)
	deleteVertices   func(*Graph, VertexSelector) (GraphIDMapping, error)
	vertexCount      func(*Graph) (int, error)
	closeGraph       func(*Graph)
}

func collectEdgeSubgraph(
	sourceVertexCount int,
	deleteIsolatedVertices bool,
	operations edgeSubgraphOperations,
) (result EdgeSubgraphResult, err error) {
	graph, err := operations.query()
	if err != nil {
		if graph != nil {
			operations.closeGraph(graph)
		}
		return EdgeSubgraphResult{}, err
	}
	if graph == nil {
		return EdgeSubgraphResult{}, errors.New("igraph: edge-subgraph query returned a nil graph")
	}
	succeeded := false
	defer func() {
		if !succeeded {
			operations.closeGraph(graph)
		}
	}()

	var mapping IDMapping
	if !deleteIsolatedVertices {
		mapping, err = operations.identityMapping(sourceVertexCount)
	} else {
		var isolated VertexSelector
		isolated, err = operations.isolatedVertices(graph)
		if err == nil {
			var deletionMapping GraphIDMapping
			deletionMapping, err = operations.deleteVertices(graph, isolated)
			mapping = deletionMapping.Vertices
		}
	}
	if err != nil {
		return EdgeSubgraphResult{}, err
	}
	if len(mapping.OldToNew) != sourceVertexCount {
		return EdgeSubgraphResult{}, fmt.Errorf(
			"igraph: edge-subgraph mapping length %d does not match source vertex count %d",
			len(mapping.OldToNew), sourceVertexCount,
		)
	}
	vertexCount, err := operations.vertexCount(graph)
	if err != nil {
		return EdgeSubgraphResult{}, err
	}
	if vertexCount != len(mapping.NewToOld) {
		return EdgeSubgraphResult{}, fmt.Errorf(
			"igraph: edge-subgraph vertex count %d does not match inverse mapping length %d",
			vertexCount, len(mapping.NewToOld),
		)
	}
	result = EdgeSubgraphResult{Graph: graph, Vertices: mapping}
	succeeded = true
	return result, nil
}

func edgeSubgraphIsolatedVertices(graph *Graph) (VertexSelector, error) {
	degrees, err := graph.Degree(AllVertices(), DegreeOptions{
		Direction:  DirectionAll,
		CountLoops: true,
	})
	if err != nil {
		return VertexSelector{}, err
	}
	isolated := make([]int, 0)
	for vertexID, degree := range degrees {
		if degree == 0 {
			isolated = append(isolated, vertexID)
		}
	}
	return VertexIDs(isolated...)
}

// DecomposeOptions controls component decomposition. Its zero value returns
// every weakly connected component, including isolated vertices. Strong mode
// applies only to directed graphs. MinimumVertices filters out smaller
// components. MaximumComponents limits the number of remaining components;
// zero means unlimited. Negative values are invalid.
type DecomposeOptions struct {
	Connectedness     ConnectednessMode
	MinimumVertices   int
	MaximumComponents int
}

// Decompose returns independently owned component graphs. Components smaller
// than MinimumVertices are filtered before MaximumComponents is applied.
// Component ordering is defined by upstream igraph and is not sorted by this
// binding. The options are borrowed only for the call and are not retained.
// The returned slice is non-nil and each graph must be closed independently.
// Every result remains usable after the source or any sibling is closed.
//
//igraph:bind igraph_decompose
func (g *Graph) Decompose(options DecomposeOptions) ([]*Graph, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	mode, maximum, minimum, err := validateDecomposeOptions(options)
	if err != nil {
		return nil, err
	}

	return collectDecomposition(decompositionOperations{
		newList:   newGraphList,
		closeList: func(list *graphList) { list.close() },
		query: func(list *graphList) error {
			if code := C.go_igraph_decompose(
				&g.graph, &list.value, mode, maximum, minimum,
			); code != C.IGRAPH_SUCCESS {
				return igraphError("decompose graph", int(code))
			}
			return nil
		},
		takeGraphs: func(list *graphList) ([]*Graph, error) { return list.takeGraphs() },
	})
}

type decompositionOperations struct {
	newList    func() (*graphList, error)
	closeList  func(*graphList)
	query      func(*graphList) error
	takeGraphs func(*graphList) ([]*Graph, error)
}

func collectDecomposition(operations decompositionOperations) ([]*Graph, error) {
	list, err := operations.newList()
	if err != nil {
		return nil, err
	}
	defer operations.closeList(list)
	if err := operations.query(list); err != nil {
		return nil, err
	}
	graphs, err := operations.takeGraphs(list)
	if err != nil {
		return nil, err
	}
	if graphs == nil {
		return nil, errors.New("igraph: decomposition returned a nil graph slice")
	}
	return graphs, nil
}

func validateDecomposeOptions(options DecomposeOptions) (C.igraph_connectedness_t, C.igraph_int_t, C.igraph_int_t, error) {
	mode, err := options.Connectedness.cValue()
	if err != nil {
		return 0, 0, 0, err
	}
	if options.MinimumVertices < 0 {
		return 0, 0, 0, fmt.Errorf("igraph: minimum component size must be non-negative: %d", options.MinimumVertices)
	}
	if options.MaximumComponents < 0 {
		return 0, 0, 0, fmt.Errorf("igraph: maximum component count must be non-negative: %d", options.MaximumComponents)
	}
	minimum, err := intToIgraphInt(options.MinimumVertices, "minimum component size")
	if err != nil {
		return 0, 0, 0, err
	}
	maximum := C.igraph_int_t(-1)
	if options.MaximumComponents > 0 {
		maximum, err = intToIgraphInt(options.MaximumComponents, "maximum component count")
		if err != nil {
			return 0, 0, 0, err
		}
	}
	return mode, maximum, minimum, nil
}

func sortedUniqueIDs(ids []int) []int {
	result := append([]int{}, ids...)
	sort.Ints(result)
	write := 0
	for _, id := range result {
		if write == 0 || result[write-1] != id {
			result[write] = id
			write++
		}
	}
	return result[:write]
}

func equalIntSlices(left, right []int) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

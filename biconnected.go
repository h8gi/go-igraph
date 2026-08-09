package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
import "C"

import (
	"errors"
	"fmt"
)

// ArticulationPoints returns the zero-based vertex IDs whose removal increases
// the number of weakly connected components. Edge directions are ignored.
// Loops do not affect the result, and parallel edges are preserved when
// deciding connectivity. The order is defined by upstream igraph.
//
// The returned non-nil slice is Go-owned and remains valid after the graph is
// closed. Empty, singleton, and edgeless graphs have no articulation points.
//
//igraph:bind igraph_articulation_points
func (g *Graph) ArticulationPoints() ([]int, error) {
	return g.cutStructureIDs("find articulation points", func(graph *C.igraph_t, result *C.igraph_vector_int_t) C.igraph_error_t {
		return C.go_igraph_articulation_points(graph, result)
	})
}

// Bridges returns the zero-based edge IDs whose removal increases the number
// of weakly connected components. Edge directions are ignored. Loops are never
// bridges, and neither edge in a parallel pair is a bridge. The order is
// defined by upstream igraph.
//
// The returned non-nil slice is Go-owned and remains valid after the graph is
// closed. Empty and edgeless graphs have no bridges.
//
//igraph:bind igraph_bridges
func (g *Graph) Bridges() ([]int, error) {
	return g.cutStructureIDs("find bridges", func(graph *C.igraph_t, result *C.igraph_vector_int_t) C.igraph_error_t {
		return C.go_igraph_bridges(graph, result)
	})
}

func (g *Graph) cutStructureIDs(
	action string,
	query func(*C.igraph_t, *C.igraph_vector_int_t) C.igraph_error_t,
) ([]int, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	return collectCutStructureIDs(cutStructureOperations{
		newVector: func() (*intVector, error) { return newIntVector(nil) },
		close:     (*intVector).close,
		query: func(result *intVector) error {
			if code := query(&g.graph, &result.value); code != C.IGRAPH_SUCCESS {
				return igraphError(action, int(code))
			}
			return nil
		},
		slice: func(result *intVector) ([]int, error) { return result.slice() },
	})
}

type cutStructureOperations struct {
	newVector func() (*intVector, error)
	close     func(*intVector)
	query     func(*intVector) error
	slice     func(*intVector) ([]int, error)
}

func collectCutStructureIDs(operations cutStructureOperations) ([]int, error) {
	result, err := operations.newVector()
	if err != nil {
		return nil, err
	}
	defer operations.close(result)
	if err := operations.query(result); err != nil {
		return nil, err
	}
	return operations.slice(result)
}

// BiconnectedComponents is a Go-owned decomposition into maximal
// biconnected subgraphs. ComponentEdges and ComponentVertices use the same
// component index, and Count equals both outer-slice lengths. ArticulationPoints
// contains the cut vertices for the whole graph.
//
// Component order, edge order, vertex order, and articulation-point order are
// defined by upstream igraph. Edge and vertex IDs are zero-based indexes into
// the source graph. Each non-loop edge belongs to exactly one component;
// vertices may occur in more than one component. Isolated vertices and loops
// are not components. A single non-loop edge is a degenerate biconnected
// component under igraph's convention. Every slice, including empty inner and
// outer slices, is non-nil and Go-owned.
type BiconnectedComponents struct {
	Count              int
	ComponentEdges     [][]int
	ComponentVertices  [][]int
	ArticulationPoints []int
}

// BiconnectedComponents returns the graph's biconnected decomposition. A
// directed graph is treated as undirected. Disconnected graphs are decomposed
// component by component; empty, singleton, and edgeless graphs return Count
// zero and non-nil empty collections. Returned data remains valid after the
// source graph is closed.
//
//igraph:bind igraph_biconnected_components
func (g *Graph) BiconnectedComponents() (BiconnectedComponents, error) {
	if g == nil {
		return BiconnectedComponents{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return BiconnectedComponents{}, ErrClosed
	}
	vertexCount := int(C.igraph_vcount(&g.graph))
	edgeCount := int(C.igraph_ecount(&g.graph))
	return collectBiconnectedComponents(vertexCount, edgeCount, biconnectedOperations{
		newList:     newIntVectorList,
		newVector:   func() (*intVector, error) { return newIntVector(nil) },
		closeList:   (*intVectorList).close,
		closeVector: (*intVector).close,
		query: func(edges, vertices *intVectorList, points *intVector) (int, error) {
			var count C.igraph_int_t
			if code := C.go_igraph_biconnected_components(
				&g.graph, &count, &edges.value, &vertices.value, &points.value,
			); code != C.IGRAPH_SUCCESS {
				return 0, igraphError("find biconnected components", int(code))
			}
			return igraphIntToInt(count, "biconnected component count")
		},
		listSlices:  func(list *intVectorList) ([][]int, error) { return list.slices() },
		vectorSlice: func(vector *intVector) ([]int, error) { return vector.slice() },
	})
}

type biconnectedOperations struct {
	newList     func() (*intVectorList, error)
	newVector   func() (*intVector, error)
	closeList   func(*intVectorList)
	closeVector func(*intVector)
	query       func(*intVectorList, *intVectorList, *intVector) (int, error)
	listSlices  func(*intVectorList) ([][]int, error)
	vectorSlice func(*intVector) ([]int, error)
}

func collectBiconnectedComponents(vertexCount, edgeCount int, operations biconnectedOperations) (BiconnectedComponents, error) {
	edges, err := operations.newList()
	if err != nil {
		return BiconnectedComponents{}, err
	}
	defer operations.closeList(edges)
	vertices, err := operations.newList()
	if err != nil {
		return BiconnectedComponents{}, err
	}
	defer operations.closeList(vertices)
	points, err := operations.newVector()
	if err != nil {
		return BiconnectedComponents{}, err
	}
	defer operations.closeVector(points)

	count, err := operations.query(edges, vertices, points)
	if err != nil {
		return BiconnectedComponents{}, err
	}
	result := BiconnectedComponents{Count: count}
	result.ComponentEdges, err = operations.listSlices(edges)
	if err == nil {
		result.ComponentVertices, err = operations.listSlices(vertices)
	}
	if err == nil {
		result.ArticulationPoints, err = operations.vectorSlice(points)
	}
	if err != nil {
		return BiconnectedComponents{}, err
	}
	if err := validateBiconnectedComponents(result, vertexCount, edgeCount); err != nil {
		return BiconnectedComponents{}, err
	}
	return result, nil
}

func validateBiconnectedComponents(result BiconnectedComponents, vertexCount, edgeCount int) error {
	if result.ComponentEdges == nil {
		return errors.New("igraph: biconnected component edge collection is nil")
	}
	if result.ComponentVertices == nil {
		return errors.New("igraph: biconnected component vertex collection is nil")
	}
	if result.ArticulationPoints == nil {
		return errors.New("igraph: biconnected articulation-point collection is nil")
	}
	if result.Count != len(result.ComponentEdges) || result.Count != len(result.ComponentVertices) {
		return fmt.Errorf(
			"igraph: biconnected component count %d does not match edge-list length %d and vertex-list length %d",
			result.Count, len(result.ComponentEdges), len(result.ComponentVertices),
		)
	}
	for componentID := 0; componentID < result.Count; componentID++ {
		if result.ComponentEdges[componentID] == nil || result.ComponentVertices[componentID] == nil {
			return fmt.Errorf("igraph: biconnected component %d contains a nil collection", componentID)
		}
		for _, edgeID := range result.ComponentEdges[componentID] {
			if edgeID < 0 || edgeID >= edgeCount {
				return fmt.Errorf("igraph: biconnected component %d edge ID %d out of range [0, %d)", componentID, edgeID, edgeCount)
			}
		}
		for _, vertexID := range result.ComponentVertices[componentID] {
			if vertexID < 0 || vertexID >= vertexCount {
				return fmt.Errorf("igraph: biconnected component %d vertex ID %d out of range [0, %d)", componentID, vertexID, vertexCount)
			}
		}
	}
	for _, vertexID := range result.ArticulationPoints {
		if vertexID < 0 || vertexID >= vertexCount {
			return fmt.Errorf("igraph: articulation point ID %d out of range [0, %d)", vertexID, vertexCount)
		}
	}
	return nil
}

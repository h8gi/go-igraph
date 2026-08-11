package igraph

/*
#include <igraph.h>
#include "connectivity_cgo.h"
*/
import "C"
import (
	"fmt"
)

// VertexConnectivityNeighbors defines how neighbor vertices are handled in vertex connectivity calculations.
type VertexConnectivityNeighbors uint8

const (
	VConnNeigError VertexConnectivityNeighbors = iota
	VConnNeigIgnore
	VConnNeigNumberOfNodes
)

func (mode VertexConnectivityNeighbors) cValue() (C.igraph_vconn_nei_t, error) {
	switch mode {
	case VConnNeigError:
		return C.IGRAPH_VCONN_NEI_ERROR, nil
	case VConnNeigIgnore:
		return C.IGRAPH_VCONN_NEI_IGNORE, nil
	case VConnNeigNumberOfNodes:
		return C.IGRAPH_VCONN_NEI_NUMBER_OF_NODES, nil
	default:
		return 0, fmt.Errorf("igraph: invalid vertex connectivity neighbors mode: %d", mode)
	}
}

// EdgeConnectivity calculates the global edge connectivity of the graph.
// The checks parameter controls whether structural checks are performed by C igraph.
// Returns a Go-owned integer count.
//
//igraph:bind igraph_edge_connectivity
func (g *Graph) EdgeConnectivity(checks bool) (int, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return 0, ErrClosed
	}

	var res C.igraph_int_t
	code := C.go_igraph_edge_connectivity(&g.graph, &res, booltoint(checks))
	if code != C.IGRAPH_SUCCESS {
		return 0, igraphError("igraph_edge_connectivity", int(code))
	}
	return igraphIntToInt(res, "edge connectivity")
}

// STEdgeConnectivity calculates the edge connectivity between a source and a target vertex.
// Source and target vertex IDs are validated. Returns a Go-owned integer count.
//
//igraph:bind igraph_st_edge_connectivity
func (g *Graph) STEdgeConnectivity(source, target int) (int, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return 0, ErrClosed
	}

	src, tgt, err := validateSourceTarget(g, source, target)
	if err != nil {
		return 0, err
	}

	var res C.igraph_int_t
	code := C.go_igraph_st_edge_connectivity(&g.graph, &res, src, tgt)
	if code != C.IGRAPH_SUCCESS {
		return 0, igraphError("igraph_st_edge_connectivity", int(code))
	}
	return igraphIntToInt(res, "st edge connectivity")
}

// VertexConnectivity calculates the global vertex connectivity of the graph.
// The checks parameter controls whether structural checks are performed by C igraph.
// Returns a Go-owned integer count.
//
//igraph:bind igraph_vertex_connectivity
func (g *Graph) VertexConnectivity(checks bool) (int, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return 0, ErrClosed
	}

	var res C.igraph_int_t
	code := C.go_igraph_vertex_connectivity(&g.graph, &res, booltoint(checks))
	if code != C.IGRAPH_SUCCESS {
		return 0, igraphError("igraph_vertex_connectivity", int(code))
	}
	return igraphIntToInt(res, "vertex connectivity")
}

// STVertexConnectivity calculates the vertex connectivity between a source and a target vertex.
// Source and target vertex IDs are validated. Returns a Go-owned integer count.
//
//igraph:bind igraph_st_vertex_connectivity
func (g *Graph) STVertexConnectivity(source, target int, neighbors VertexConnectivityNeighbors) (int, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return 0, ErrClosed
	}

	src, tgt, err := validateSourceTarget(g, source, target)
	if err != nil {
		return 0, err
	}

	cNeighbors, err := neighbors.cValue()
	if err != nil {
		return 0, err
	}

	var res C.igraph_int_t
	code := C.go_igraph_st_vertex_connectivity(&g.graph, &res, src, tgt, cNeighbors)
	if code != C.IGRAPH_SUCCESS {
		return 0, igraphError("igraph_st_vertex_connectivity", int(code))
	}
	return igraphIntToInt(res, "st vertex connectivity")
}

// EdgeDisjointPaths calculates the maximum number of edge-disjoint paths between source and target.
// Source and target vertex IDs are validated. Returns a Go-owned integer count.
//
//igraph:bind igraph_edge_disjoint_paths
func (g *Graph) EdgeDisjointPaths(source, target int) (int, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return 0, ErrClosed
	}

	src, tgt, err := validateSourceTarget(g, source, target)
	if err != nil {
		return 0, err
	}

	var res C.igraph_int_t
	code := C.go_igraph_edge_disjoint_paths(&g.graph, &res, src, tgt)
	if code != C.IGRAPH_SUCCESS {
		return 0, igraphError("igraph_edge_disjoint_paths", int(code))
	}
	return igraphIntToInt(res, "edge disjoint paths")
}

// VertexDisjointPaths calculates the maximum number of vertex-disjoint paths between source and target.
// Source and target vertex IDs are validated. Returns a Go-owned integer count.
//
//igraph:bind igraph_vertex_disjoint_paths
func (g *Graph) VertexDisjointPaths(source, target int) (int, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return 0, ErrClosed
	}

	src, tgt, err := validateSourceTarget(g, source, target)
	if err != nil {
		return 0, err
	}

	var res C.igraph_int_t
	code := C.go_igraph_vertex_disjoint_paths(&g.graph, &res, src, tgt)
	if code != C.IGRAPH_SUCCESS {
		return 0, igraphError("igraph_vertex_disjoint_paths", int(code))
	}
	return igraphIntToInt(res, "vertex disjoint paths")
}

// Adhesion calculates the edge connectivity (adhesion) of a graph.
// The checks parameter controls whether structural checks are performed by C igraph.
// Returns a Go-owned integer count.
//
//igraph:bind igraph_adhesion
func (g *Graph) Adhesion(checks bool) (int, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return 0, ErrClosed
	}

	var res C.igraph_int_t
	code := C.go_igraph_adhesion(&g.graph, &res, booltoint(checks))
	if code != C.IGRAPH_SUCCESS {
		return 0, igraphError("igraph_adhesion", int(code))
	}
	return igraphIntToInt(res, "adhesion")
}

// Cohesion calculates the vertex connectivity (cohesion) of a graph.
// The checks parameter controls whether structural checks are performed by C igraph.
// Returns a Go-owned integer count.
//
//igraph:bind igraph_cohesion
func (g *Graph) Cohesion(checks bool) (int, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return 0, ErrClosed
	}

	var res C.igraph_int_t
	code := C.go_igraph_cohesion(&g.graph, &res, booltoint(checks))
	if code != C.IGRAPH_SUCCESS {
		return 0, igraphError("igraph_cohesion", int(code))
	}
	return igraphIntToInt(res, "cohesion")
}

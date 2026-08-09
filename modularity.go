package igraph

/*
#cgo pkg-config: igraph
#include <igraph.h>
#include "algorithm_cgo.h"
#include "modularity_cgo.h"
*/
import "C"

import (
	"fmt"
)

// NeiMode controls how neighbor direction is interpreted in graph calculations.
type NeiMode = DirectionMode

const (
	NeiOut = DirectionOut
	NeiIn  = DirectionIn
	NeiAll = DirectionAll
)

// Coreness calculates the k-coreness of each vertex in the graph.
//
// The graph parameter is borrowed; the returned slice is Go-owned and remains
// valid after subsequent graph operations or Close.
//
//igraph:bind igraph_coreness
func (g *Graph) Coreness(mode NeiMode) ([]int, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if err := g.checkClosed(); err != nil {
		return nil, err
	}

	cMode, err := mode.cValue()
	if err != nil {
		return nil, err
	}

	coresVec, err := newIntVector(nil)
	if err != nil {
		return nil, err
	}
	defer coresVec.close()

	if code := C.go_igraph_coreness(&g.graph, &coresVec.value, cMode); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("calculate coreness", int(code))
	}

	return coresVec.slice()
}

// Trussness calculates the k-trussness of each edge in the graph.
//
// The graph parameter is borrowed; the returned slice is Go-owned and remains
// valid after subsequent graph operations or Close.
//
//igraph:bind igraph_trussness
func (g *Graph) Trussness() ([]int, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if err := g.checkClosed(); err != nil {
		return nil, err
	}

	trussVec, err := newIntVector(nil)
	if err != nil {
		return nil, err
	}
	defer trussVec.close()

	if code := C.go_igraph_trussness(&g.graph, &trussVec.value); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("calculate trussness", int(code))
	}

	return trussVec.slice()
}

// Modularity calculates the modularity score of a graph partition defined by membership.
//
// Graph, membership, and weights inputs are borrowed; the returned score is Go-owned.
// Nil weights select an unweighted calculation.
//
//igraph:bind igraph_modularity
func (g *Graph) Modularity(membership []int, weights []float64, resolution float64) (float64, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if err := g.checkClosed(); err != nil {
		return 0, err
	}

	vcount := int(C.igraph_vcount(&g.graph))
	if len(membership) != vcount {
		return 0, fmt.Errorf("igraph: membership length %d does not match vertex count %d", len(membership), vcount)
	}

	weightsVec, err := newOptionalEdgeWeights(weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return 0, err
	}
	if weightsVec != nil {
		defer weightsVec.close()
	}

	memVec, err := newIntVector(membership)
	if err != nil {
		return 0, err
	}
	defer memVec.close()

	var mod C.igraph_real_t
	directed := C.igraph_is_directed(&g.graph)

	if code := C.go_igraph_modularity(
		&g.graph, &memVec.value, edgeWeightPointer(weightsVec), C.igraph_real_t(resolution), directed, &mod,
	); code != C.IGRAPH_SUCCESS {
		return 0, igraphError("calculate modularity", int(code))
	}

	return float64(mod), nil
}

// ModularityMatrix calculates the modularity matrix of a graph.
//
// Graph and weights inputs are borrowed; the returned Matrix is Go-owned.
// Nil weights select an unweighted calculation.
//
//igraph:bind igraph_modularity_matrix
func (g *Graph) ModularityMatrix(weights []float64, resolution float64) (*Matrix, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if err := g.checkClosed(); err != nil {
		return nil, err
	}

	weightsVec, err := newOptionalEdgeWeights(weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return nil, err
	}
	if weightsVec != nil {
		defer weightsVec.close()
	}

	var cmat cMatrix
	if code := C.go_igraph_matrix_init(&cmat.value, 0, 0); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("initialize modularity matrix", int(code))
	}
	defer cmat.close()

	directed := C.igraph_is_directed(&g.graph)
	if code := C.go_igraph_modularity_matrix(
		&g.graph, edgeWeightPointer(weightsVec), C.igraph_real_t(resolution), &cmat.value, directed,
	); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("calculate modularity matrix", int(code))
	}

	resMatrix, err := cmat.matrix()
	if err != nil {
		return nil, err
	}
	return &resMatrix, nil
}

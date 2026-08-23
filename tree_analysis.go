package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "tree_analysis_cgo.h"
import "C"

import (
	"fmt"
	"math"
)

// TreeTestResult reports whether a graph is a tree and, when it is, its root.
// Root is NoParent when the graph is not a tree. DirectionAll ignores edge
// directions and reports vertex 0 as the root; DirectionOut and DirectionIn
// require a consistently oriented tree and report its unique root. The empty
// graph is not a tree, while a singleton is. The result is Go-owned.
type TreeTestResult struct {
	IsTree bool
	Root   int
}

// ForestTestResult reports whether a graph is a forest and its component
// roots when it is. Roots follows upstream component-discovery order. In
// DirectionAll, one arbitrary root is reported per component; directed modes
// report the zero in- or out-degree roots. The empty graph is a forest. Roots
// is a non-nil Go-owned slice, empty for non-forests and the empty graph.
type ForestTestResult struct {
	IsForest bool
	Roots    []int
}

// UnfoldTreeResult contains an independently owned tree and a Go-owned mapping
// from each tree vertex to its source-graph vertex. Graph must be closed.
type UnfoldTreeResult struct {
	Graph          *Graph
	SourceVertices []int
}

// IsTree tests the graph using direction to interpret directed edges.
//
//igraph:bind igraph_is_tree
func (g *Graph) IsTree(direction DirectionMode) (TreeTestResult, error) {
	if g == nil {
		return TreeTestResult{}, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return TreeTestResult{}, ErrClosed
	}
	mode, err := direction.cValue()
	if err != nil {
		return TreeTestResult{}, err
	}
	var result C.igraph_bool_t
	root := C.igraph_int_t(NoParent)
	if code := C.go_igraph_is_tree(&g.graph, &result, &root, mode); code != C.IGRAPH_SUCCESS {
		return TreeTestResult{}, igraphError("test whether graph is a tree", int(code))
	}
	if result == booltoint(false) {
		return TreeTestResult{Root: NoParent}, nil
	}
	return TreeTestResult{IsTree: true, Root: int(root)}, nil
}

// IsForest tests the graph using direction to interpret directed edges.
//
//igraph:bind igraph_is_forest
func (g *Graph) IsForest(direction DirectionMode) (ForestTestResult, error) {
	if g == nil {
		return ForestTestResult{}, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return ForestTestResult{}, ErrClosed
	}
	mode, err := direction.cValue()
	if err != nil {
		return ForestTestResult{}, err
	}
	roots, err := newIntVector(nil)
	if err != nil {
		return ForestTestResult{}, err
	}
	defer roots.close()
	var result C.igraph_bool_t
	if code := C.go_igraph_is_forest(&g.graph, &result, &roots.value, mode); code != C.IGRAPH_SUCCESS {
		return ForestTestResult{}, igraphError("test whether graph is a forest", int(code))
	}
	if result == booltoint(false) {
		return ForestTestResult{Roots: []int{}}, nil
	}
	values, err := roots.slice()
	if err != nil {
		return ForestTestResult{}, err
	}
	return ForestTestResult{IsForest: true, Roots: values}, nil
}

// MinimumSpanningForest returns a minimum spanning forest. weights may be nil
// for an unweighted forest; otherwise it is borrowed for the call and must
// contain one finite value per edge. Negative weights are supported. Directed
// edges are treated as undirected; disconnected input produces one minimum
// spanning tree per component. The returned edge IDs follow the source graph,
// are non-nil and Go-owned.
//
//igraph:bind igraph_minimum_spanning_tree
func (g *Graph) MinimumSpanningForest(weights []float64) ([]int, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return nil, ErrClosed
	}
	if weights != nil && len(weights) != int(C.igraph_ecount(&g.graph)) {
		return nil, fmt.Errorf("igraph: weight count %d does not match edge count %d", len(weights), int(C.igraph_ecount(&g.graph)))
	}
	var cWeights *realVector
	var err error
	if weights != nil {
		cWeights, err = newRealVector(weights)
		if err != nil {
			return nil, err
		}
		defer cWeights.close()
		for index, weight := range weights {
			if math.IsNaN(weight) || math.IsInf(weight, 0) {
				return nil, fmt.Errorf("igraph: weight at index %d must be finite: %v", index, weight)
			}
		}
	}
	result, err := newIntVector(nil)
	if err != nil {
		return nil, err
	}
	defer result.close()
	var ptr *C.igraph_vector_t
	method := C.igraph_mst_algorithm_t(C.IGRAPH_MST_UNWEIGHTED)
	if cWeights != nil {
		ptr = &cWeights.value
		method = C.igraph_mst_algorithm_t(C.IGRAPH_MST_AUTOMATIC)
	}
	if code := C.go_igraph_minimum_spanning_tree(&g.graph, &result.value, ptr, method); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("calculate minimum spanning forest", int(code))
	}
	values, err := result.slice()
	if err != nil {
		return nil, err
	}
	if values == nil {
		values = []int{}
	}
	return values, nil
}

// UnfoldTree unfolds graph traversal from roots into an independently owned
// tree (or forest for disconnected input). Roots are traversed in the supplied
// order and are borrowed for the call; every root must be a valid vertex.
//
//igraph:bind igraph_unfold_tree
func (g *Graph) UnfoldTree(roots []int, direction DirectionMode) (UnfoldTreeResult, error) {
	if g == nil {
		return UnfoldTreeResult{}, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return UnfoldTreeResult{}, ErrClosed
	}
	mode, err := direction.cValue()
	if err != nil {
		return UnfoldTreeResult{}, err
	}
	vertexCount := int(C.igraph_vcount(&g.graph))
	for i, root := range roots {
		if root < 0 || root >= vertexCount {
			return UnfoldTreeResult{}, fmt.Errorf("igraph: root at index %d out of range [0, %d): %d", i, vertexCount, root)
		}
	}
	cRoots, err := newIntVector(roots)
	if err != nil {
		return UnfoldTreeResult{}, err
	}
	defer cRoots.close()
	mapping, err := newIntVector(nil)
	if err != nil {
		return UnfoldTreeResult{}, err
	}
	defer mapping.close()
	var tree C.igraph_t
	if code := C.go_igraph_unfold_tree(&g.graph, &tree, mode, &cRoots.value, &mapping.value); code != C.IGRAPH_SUCCESS {
		return UnfoldTreeResult{}, igraphError("unfold graph into tree", int(code))
	}
	values, err := mapping.slice()
	if err != nil {
		C.igraph_destroy(&tree)
		return UnfoldTreeResult{}, err
	}
	return UnfoldTreeResult{Graph: adoptInitializedGraph(&tree), SourceVertices: values}, nil
}

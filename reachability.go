package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
// #include "reachability_cgo.h"
import "C"

import "fmt"

// ReachabilityResult is a Go-owned directed reachability decomposition.
// Membership is indexed by vertex ID, Sizes by component ID, and Reachable by
// source vertex ID. Each reachable set includes its source vertex.
type ReachabilityResult struct {
	Membership     []int
	Sizes          []int
	ComponentCount int
	Reachable      [][]int
}

// NeighborhoodGraphResult pairs an independently owned induced neighborhood
// graph with its selected root and original source vertex IDs. Graph must be
// closed by the caller.
type NeighborhoodGraphResult struct {
	Graph          *Graph
	Root           int
	SourceVertices []int
}

// Reachability returns reachability sets and strongly or weakly connected
// component metadata according to direction. The result is entirely Go-owned.
//
//igraph:bind igraph_reachability
func (g *Graph) Reachability(direction DirectionMode) (ReachabilityResult, error) {
	if g == nil {
		return ReachabilityResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return ReachabilityResult{}, ErrClosed
	}
	mode, err := direction.cValue()
	if err != nil {
		return ReachabilityResult{}, err
	}
	membership, err := newIntVector(nil)
	if err != nil {
		return ReachabilityResult{}, err
	}
	defer membership.close()
	sizes, err := newIntVector(nil)
	if err != nil {
		return ReachabilityResult{}, err
	}
	defer sizes.close()
	reachable, err := newIntVectorList()
	if err != nil {
		return ReachabilityResult{}, err
	}
	defer reachable.close()
	var componentCount C.igraph_int_t
	code := C.go_igraph_reachability(
		&g.graph, &membership.value, &sizes.value, &componentCount,
		&reachable.value, mode,
	)
	if code != C.IGRAPH_SUCCESS {
		return ReachabilityResult{}, igraphError("calculate reachability", int(code))
	}
	members, err := membership.slice()
	if err != nil {
		return ReachabilityResult{}, err
	}
	componentSizes, err := sizes.slice()
	if err != nil {
		return ReachabilityResult{}, err
	}
	sets, err := reachable.slices()
	if err != nil {
		return ReachabilityResult{}, err
	}
	if len(sets) != int(componentCount) {
		return ReachabilityResult{}, fmt.Errorf("igraph: reachability component result length mismatch")
	}
	vertexSets := make([][]int, len(members))
	for vertex, component := range members {
		if component < 0 || component >= len(sets) {
			return ReachabilityResult{}, fmt.Errorf("igraph: invalid reachability component for vertex %d: %d", vertex, component)
		}
		vertexSets[vertex] = append([]int(nil), sets[component]...)
		if vertexSets[vertex] == nil {
			vertexSets[vertex] = []int{}
		}
	}
	return ReachabilityResult{
		Membership: members, Sizes: componentSizes,
		ComponentCount: int(componentCount), Reachable: vertexSets,
	}, nil
}

// ReachableCounts returns the number of vertices reachable from every vertex,
// including the source vertex itself. The returned slice is Go-owned.
//
//igraph:bind igraph_count_reachable
func (g *Graph) ReachableCounts(direction DirectionMode) ([]int, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	mode, err := direction.cValue()
	if err != nil {
		return nil, err
	}
	counts, err := newIntVector(nil)
	if err != nil {
		return nil, err
	}
	defer counts.close()
	if code := C.go_igraph_count_reachable(&g.graph, &counts.value, mode); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("count reachable vertices", int(code))
	}
	return counts.slice()
}

// NeighborhoodGraphs returns one independently owned induced graph per root,
// preserving selector order and duplicates. SourceVertices maps every result
// vertex ID back to the source graph. Each Graph must be closed independently.
//
//igraph:bind igraph_neighborhood_graphs
func (g *Graph) NeighborhoodGraphs(vertices VertexSelector, options NeighborhoodOptions) ([]NeighborhoodGraphResult, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	selector, order, minDistance, mode, err := g.prepareNeighborhood(vertices, options)
	if err != nil {
		return nil, err
	}
	defer selector.close()
	roots, err := materializeVertexIDs(&g.graph, vertices)
	if err != nil {
		return nil, err
	}
	provenance, err := newIntVectorList()
	if err != nil {
		return nil, err
	}
	defer provenance.close()
	if code := C.go_igraph_neighborhood(
		&g.graph, &provenance.value, selector.value, order, mode, minDistance,
	); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("calculate neighborhood provenance", int(code))
	}
	sourceVertices, err := provenance.slices()
	if err != nil {
		return nil, err
	}
	graphs, err := newGraphList()
	if err != nil {
		return nil, err
	}
	defer graphs.close()
	if code := C.go_igraph_neighborhood_graphs(
		&g.graph, &graphs.value, selector.value, order, mode, minDistance,
	); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("create neighborhood graphs", int(code))
	}
	owned, err := graphs.takeGraphs()
	if err != nil {
		return nil, err
	}
	if len(owned) != len(roots) || len(sourceVertices) != len(roots) {
		for _, graph := range owned {
			_ = graph.Close()
		}
		return nil, fmt.Errorf("igraph: neighborhood result length mismatch")
	}
	result := make([]NeighborhoodGraphResult, len(owned))
	for i := range owned {
		result[i] = NeighborhoodGraphResult{Graph: owned[i], Root: roots[i], SourceVertices: sourceVertices[i]}
	}
	return result, nil
}

// TransitiveClosure returns an independently owned graph containing an edge
// for every reachable distinct ordered vertex pair. The caller must close it.
//
//igraph:bind igraph_transitive_closure
func (g *Graph) TransitiveClosure() (*Graph, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	var closure C.igraph_t
	if code := C.go_igraph_transitive_closure(&g.graph, &closure); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("create transitive closure", int(code))
	}
	return adoptInitializedGraph(&closure), nil
}

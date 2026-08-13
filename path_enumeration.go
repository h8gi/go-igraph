package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
import "C"

import (
	"errors"
	"fmt"
)

// PathsResult is a Go-owned path collection. Paths is always non-nil.
// Truncated is meaningful for explicitly bounded enumeration operations.
type PathsResult struct {
	Paths     []Path
	Truncated bool
}

// SimplePathOptions controls bounded simple-path enumeration. MaxResults must
// be positive, so potentially exponential work is never requested without an
// explicit result bound. MinEdges and MaxEdges are optional inclusive bounds;
// nil means no bound on that side.
type SimplePathOptions struct {
	Direction  DirectionMode
	MinEdges   *int
	MaxEdges   *int
	MaxResults int
}

// ShortestPaths returns one shortest path from source to each selected target.
// Target order and duplicates are preserved. Inputs are borrowed only for the
// synchronous call; all results are Go-owned and survive graph closure.
//
//igraph:bind igraph_get_shortest_paths
func (g *Graph) ShortestPaths(source int, targets VertexSelector, options PathOptions) ([]Path, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	mode, weights, selector, targetPositions, err := g.preparePathCollection(source, targets, options)
	if err != nil {
		return nil, err
	}
	defer selector.close()
	if weights != nil {
		defer weights.close()
	}
	vertices, edges, err := newPathVectorLists()
	if err != nil {
		return nil, err
	}
	defer vertices.close()
	defer edges.close()
	code := C.go_igraph_get_shortest_paths(&g.graph, edgeWeightPointer(weights), &vertices.value, &edges.value, C.igraph_int_t(source), selector.value, mode)
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("calculate shortest paths", int(code))
	}
	paths, err := convertPathLists(vertices, edges, true)
	if err != nil {
		return nil, err
	}
	return expandByPositions(paths, targetPositions), nil
}

// KShortestPaths returns at most k loopless paths in nondecreasing total
// weight order. k must be positive. If fewer paths exist, all are returned.
// Non-nil weights must be finite and non-negative.
//
//igraph:bind igraph_get_k_shortest_paths
func (g *Graph) KShortestPaths(source, target, k int, options PathOptions) ([]Path, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	if k <= 0 {
		return nil, fmt.Errorf("igraph: k-shortest path count must be positive: %d", k)
	}
	if _, err := intToIgraphInt(k, "k-shortest path count"); err != nil {
		return nil, err
	}
	mode, err := options.Direction.cValue()
	if err != nil {
		return nil, err
	}
	vc := int(C.igraph_vcount(&g.graph))
	if err := validateVertexID(source, vc); err != nil {
		return nil, fmt.Errorf("igraph: invalid path source: %w", err)
	}
	if err := validateVertexID(target, vc); err != nil {
		return nil, fmt.Errorf("igraph: invalid path target: %w", err)
	}
	for index, weight := range options.Weights {
		if weight < 0 {
			return nil, fmt.Errorf("igraph: k-shortest path weight at index %d must be non-negative: %v", index, weight)
		}
	}
	weights, err := newOptionalEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return nil, err
	}
	if weights != nil {
		defer weights.close()
	}
	vertices, edges, err := newPathVectorLists()
	if err != nil {
		return nil, err
	}
	defer vertices.close()
	defer edges.close()
	code := C.go_igraph_get_k_shortest_paths(&g.graph, edgeWeightPointer(weights), &vertices.value, &edges.value, C.igraph_int_t(k), C.igraph_int_t(source), C.igraph_int_t(target), mode)
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("calculate k shortest paths", int(code))
	}
	return convertPathLists(vertices, edges, true)
}

// SimplePaths enumerates at most MaxResults simple vertex paths. Upstream
// ignores loops and collapses parallel-edge choices, so Edges contains one
// matching edge ID for each consecutive vertex pair and is not a promise to
// enumerate distinct parallel-edge realizations.
//
//igraph:bind igraph_get_all_simple_paths
func (g *Graph) SimplePaths(source int, targets VertexSelector, options SimplePathOptions) (PathsResult, error) {
	if g == nil {
		return PathsResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return PathsResult{}, ErrClosed
	}
	params, err := prepareSimplePathOptions(options)
	if err != nil {
		return PathsResult{}, err
	}
	vc := int(C.igraph_vcount(&g.graph))
	if err := validateVertexID(source, vc); err != nil {
		return PathsResult{}, fmt.Errorf("igraph: invalid path source: %w", err)
	}
	if err := validateVertexSelector(targets, vc); err != nil {
		return PathsResult{}, fmt.Errorf("igraph: invalid target selector: %w", err)
	}
	selector, err := newCVertexSelector(targets)
	if err != nil {
		return PathsResult{}, err
	}
	defer selector.close()
	list, err := newIntVectorList()
	if err != nil {
		return PathsResult{}, err
	}
	defer list.close()
	mode, _ := options.Direction.cValue()
	code := C.go_igraph_get_all_simple_paths(&g.graph, &list.value, C.igraph_int_t(source), selector.value, mode, C.igraph_int_t(params.min), C.igraph_int_t(params.max), C.igraph_int_t(params.requested))
	if code != C.IGRAPH_SUCCESS {
		return PathsResult{}, igraphError("enumerate simple paths", int(code))
	}
	vertexPaths, err := list.slices()
	if err != nil {
		return PathsResult{}, err
	}
	truncated := len(vertexPaths) > options.MaxResults
	if truncated {
		vertexPaths = vertexPaths[:options.MaxResults]
	}
	paths := make([]Path, len(vertexPaths))
	for i, vertices := range vertexPaths {
		edges, err := g.edgePathForVertices(vertices, options.Direction)
		if err != nil {
			return PathsResult{}, err
		}
		paths[i] = Path{Vertices: vertices, Edges: edges, Found: true}
	}
	return PathsResult{Paths: paths, Truncated: truncated}, nil
}

type simplePathParameters struct{ min, max, requested int }

func prepareSimplePathOptions(options SimplePathOptions) (simplePathParameters, error) {
	if _, err := options.Direction.cValue(); err != nil {
		return simplePathParameters{}, err
	}
	if options.MaxResults <= 0 {
		return simplePathParameters{}, fmt.Errorf("igraph: simple-path maximum results must be positive: %d", options.MaxResults)
	}
	if options.MaxResults == int(^uint(0)>>1) {
		return simplePathParameters{}, errors.New("igraph: simple-path maximum results plus one is out of range")
	}
	requested := options.MaxResults + 1
	if _, err := intToIgraphInt(requested, "simple-path requested results"); err != nil {
		return simplePathParameters{}, err
	}
	min, err := optionalPathLength(options.MinEdges, "minimum")
	if err != nil {
		return simplePathParameters{}, err
	}
	max, err := optionalPathLength(options.MaxEdges, "maximum")
	if err != nil {
		return simplePathParameters{}, err
	}
	if min >= 0 && max >= 0 && min > max {
		return simplePathParameters{}, fmt.Errorf("igraph: simple-path minimum length %d exceeds maximum length %d", min, max)
	}
	return simplePathParameters{min: min, max: max, requested: requested}, nil
}

func optionalPathLength(value *int, name string) (int, error) {
	if value == nil {
		return -1, nil
	}
	if *value < 0 {
		return 0, fmt.Errorf("igraph: simple-path %s length must be non-negative: %d", name, *value)
	}
	if _, err := intToIgraphInt(*value, "simple-path "+name+" length"); err != nil {
		return 0, err
	}
	return *value, nil
}

func (g *Graph) preparePathCollection(source int, targets VertexSelector, options PathOptions) (C.igraph_neimode_t, *realVector, *cVertexSelector, []int, error) {
	mode, err := options.Direction.cValue()
	if err != nil {
		return 0, nil, nil, nil, err
	}
	vc := int(C.igraph_vcount(&g.graph))
	if err := validateVertexID(source, vc); err != nil {
		return 0, nil, nil, nil, fmt.Errorf("igraph: invalid path source: %w", err)
	}
	if err := validateVertexSelector(targets, vc); err != nil {
		return 0, nil, nil, nil, fmt.Errorf("igraph: invalid target selector: %w", err)
	}
	targetIDs, err := materializeVertexIDs(&g.graph, targets)
	if err != nil {
		return 0, nil, nil, nil, fmt.Errorf("igraph: materialize target selector: %w", err)
	}
	uniqueTargets, targetPositions := deduplicateIDs(targetIDs)
	uniqueSelector, err := VertexIDs(uniqueTargets...)
	if err != nil {
		return 0, nil, nil, nil, err
	}
	selector, err := newCVertexSelector(uniqueSelector)
	if err != nil {
		return 0, nil, nil, nil, err
	}
	weights, err := newOptionalEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		selector.close()
		return 0, nil, nil, nil, err
	}
	return mode, weights, selector, targetPositions, nil
}

func newPathVectorLists() (*intVectorList, *intVectorList, error) {
	vertices, err := newIntVectorList()
	if err != nil {
		return nil, nil, err
	}
	edges, err := newIntVectorList()
	if err != nil {
		vertices.close()
		return nil, nil, err
	}
	return vertices, edges, nil
}

func convertPathLists(vertices, edges *intVectorList, includeUnreachable bool) ([]Path, error) {
	vertexSlices, err := vertices.slices()
	if err != nil {
		return nil, err
	}
	edgeSlices, err := edges.slices()
	if err != nil {
		return nil, err
	}
	if len(vertexSlices) != len(edgeSlices) {
		return nil, fmt.Errorf("igraph: path vertex/edge result length mismatch: %d and %d", len(vertexSlices), len(edgeSlices))
	}
	result := make([]Path, len(vertexSlices))
	for i := range result {
		result[i] = Path{Vertices: vertexSlices[i], Edges: edgeSlices[i], Found: !includeUnreachable || len(vertexSlices[i]) != 0}
	}
	return result, nil
}

func (g *Graph) edgePathForVertices(vertices []int, direction DirectionMode) ([]int, error) {
	edges := make([]int, 0, max(0, len(vertices)-1))
	directed := C.igraph_is_directed(&g.graph) != booltoint(false)
	for i := 0; i+1 < len(vertices); i++ {
		from, to := vertices[i], vertices[i+1]
		if directed && direction == DirectionIn {
			from, to = to, from
		}
		var edge C.igraph_int_t
		code := C.go_igraph_get_eid(&g.graph, &edge, C.igraph_int_t(from), C.igraph_int_t(to), booltoint(directed && direction != DirectionAll), booltoint(false))
		if code != C.IGRAPH_SUCCESS {
			return nil, igraphError("resolve simple-path edge", int(code))
		}
		if edge < 0 {
			return nil, fmt.Errorf("igraph: no edge for simple-path step %d -> %d", vertices[i], vertices[i+1])
		}
		edges = append(edges, int(edge))
	}
	return edges, nil
}

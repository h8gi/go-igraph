package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
import "C"

import "fmt"

// BFSOptions configures a breadth-first traversal. Roots is required and is
// considered in caller order; it is borrowed only for the duration of the
// call. A root reached from an earlier root is not started as a new tree.
//
// Restriction limits the traversal to the selected vertices. Its zero value,
// AllVertices(), imposes no restriction. Every root must be selected by
// Restriction. When TraverseUnreachable is true, new trees are started in
// vertex ID order until every selected vertex has been visited.
type BFSOptions struct {
	Roots               []int
	Direction           DirectionMode
	TraverseUnreachable bool
	Restriction         VertexSelector
}

// BFSResult contains Go-owned traversal results. Order is the discovery
// sequence and contains only visited vertex IDs. Parents and Distances are
// indexed by vertex ID and therefore have one entry per graph vertex.
//
// Parents contains -1 for the root of each search tree and -2 for an unvisited
// vertex. Distances contains the depth in the returned parent forest, or -1
// when the vertex was not visited.
type BFSResult struct {
	Order     []int
	Parents   []int
	Distances []int
}

// DFSOptions configures a depth-first traversal. Root must identify a graph
// vertex. When TraverseUnreachable is true, new trees are started in vertex ID
// order after the root's reachable component has been visited.
type DFSOptions struct {
	Root                int
	Direction           DirectionMode
	TraverseUnreachable bool
}

// DFSResult contains Go-owned traversal results. Order and FinishOrder contain
// only visited vertex IDs, in discovery and subtree-completion order,
// respectively. Parents and Distances are indexed by vertex ID and therefore
// have one entry per graph vertex.
//
// Parents contains -1 for the root of each search tree and -2 for an unvisited
// vertex. Distances contains the depth in the returned parent forest, or -1
// when the vertex was not visited.
type DFSResult struct {
	Order       []int
	FinishOrder []int
	Parents     []int
	Distances   []int
}

// BreadthFirstSearch performs a breadth-first traversal without invoking Go
// callbacks. It returns an error when Roots is empty, including for an empty
// graph. All returned slices are independent of the graph and remain valid
// after Close.
//
//igraph:bind igraph_bfs
func (g *Graph) BreadthFirstSearch(options BFSOptions) (BFSResult, error) {
	if g == nil {
		return BFSResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return BFSResult{}, ErrClosed
	}

	cMode, err := options.Direction.cValue()
	if err != nil {
		return BFSResult{}, err
	}
	vertexCount := int(C.igraph_vcount(&g.graph))
	if len(options.Roots) == 0 {
		return BFSResult{}, fmt.Errorf("igraph: breadth-first search requires at least one root")
	}
	for index, root := range options.Roots {
		if err := validateVertexID(root, vertexCount); err != nil {
			return BFSResult{}, fmt.Errorf("igraph: breadth-first root at index %d: %w", index, err)
		}
	}
	if err := validateVertexSelector(options.Restriction, vertexCount); err != nil {
		return BFSResult{}, fmt.Errorf("igraph: breadth-first restriction: %w", err)
	}
	restrictedIDs, err := materializeVertexIDs(&g.graph, options.Restriction)
	if err != nil {
		return BFSResult{}, err
	}
	if err := validateRootsRestricted(options.Roots, restrictedIDs); err != nil {
		return BFSResult{}, err
	}

	inputs, err := newTraversalIntVectors(options.Roots, restrictedIDs)
	if err != nil {
		return BFSResult{}, err
	}
	defer closeIntVectors(inputs)
	outputs, err := newTraversalIntVectors(nil, nil, nil)
	if err != nil {
		return BFSResult{}, err
	}
	defer closeIntVectors(outputs)

	if code := C.go_igraph_bfs(
		&g.graph,
		0,
		&inputs[0].value,
		cMode,
		booltoint(options.TraverseUnreachable),
		&inputs[1].value,
		&outputs[0].value,
		nil,
		&outputs[1].value,
		nil,
		nil,
		&outputs[2].value,
		nil,
		nil,
	); code != C.IGRAPH_SUCCESS {
		return BFSResult{}, igraphError("perform breadth-first search", int(code))
	}

	values, err := traversalSlices(outputs)
	if err != nil {
		return BFSResult{}, err
	}
	return BFSResult{
		Order:     trimTraversalOrder(values[0]),
		Parents:   values[1],
		Distances: values[2],
	}, nil
}

// DepthFirstSearch performs a depth-first traversal without invoking Go
// callbacks. Root must identify a vertex, so an empty graph always returns an
// error. All returned slices are independent of the graph and remain valid
// after Close.
//
//igraph:bind igraph_dfs
func (g *Graph) DepthFirstSearch(options DFSOptions) (DFSResult, error) {
	if g == nil {
		return DFSResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return DFSResult{}, ErrClosed
	}

	cMode, err := options.Direction.cValue()
	if err != nil {
		return DFSResult{}, err
	}
	if err := validateVertexID(options.Root, int(C.igraph_vcount(&g.graph))); err != nil {
		return DFSResult{}, fmt.Errorf("igraph: depth-first root: %w", err)
	}
	outputs, err := newTraversalIntVectors(nil, nil, nil)
	if err != nil {
		return DFSResult{}, err
	}
	defer closeIntVectors(outputs)

	if code := C.go_igraph_dfs(
		&g.graph,
		C.igraph_int_t(options.Root),
		cMode,
		booltoint(options.TraverseUnreachable),
		&outputs[0].value,
		&outputs[1].value,
		&outputs[2].value,
		nil,
		nil,
		nil,
		nil,
	); code != C.IGRAPH_SUCCESS {
		return DFSResult{}, igraphError("perform depth-first search", int(code))
	}

	values, err := traversalSlices(outputs)
	if err != nil {
		return DFSResult{}, err
	}
	order := trimTraversalOrder(values[0])
	finishOrder := trimTraversalOrder(values[1])
	parents := values[2]
	// igraph 1.0.1 can overwrite the distance of a previously discovered
	// vertex when TraverseUnreachable starts later trees. Deriving depths from
	// the returned forest keeps the public result internally consistent.
	distances := traversalDistances(order, parents)
	return DFSResult{
		Order:       order,
		FinishOrder: finishOrder,
		Parents:     parents,
		Distances:   distances,
	}, nil
}

func validateRootsRestricted(roots, restricted []int) error {
	selected := make(map[int]struct{}, len(restricted))
	for _, vertex := range restricted {
		selected[vertex] = struct{}{}
	}
	for index, root := range roots {
		if _, ok := selected[root]; !ok {
			return fmt.Errorf(
				"igraph: breadth-first root at index %d (%d) is not selected by the restriction",
				index, root,
			)
		}
	}
	return nil
}

func newTraversalIntVectors(values ...[]int) ([]*intVector, error) {
	result := make([]*intVector, 0, len(values))
	for _, value := range values {
		vector, err := newIntVector(value)
		if err != nil {
			closeIntVectors(result)
			return nil, err
		}
		result = append(result, vector)
	}
	return result, nil
}

func closeIntVectors(vectors []*intVector) {
	for _, vector := range vectors {
		vector.close()
	}
}

func traversalSlices(vectors []*intVector) ([][]int, error) {
	result := make([][]int, len(vectors))
	for index, vector := range vectors {
		values, err := vector.slice()
		if err != nil {
			return nil, err
		}
		result[index] = values
	}
	return result, nil
}

func trimTraversalOrder(values []int) []int {
	for index, value := range values {
		if value < 0 {
			return values[:index]
		}
	}
	return values
}

func traversalDistances(order, parents []int) []int {
	result := make([]int, len(parents))
	for index := range result {
		result[index] = -1
	}
	for _, vertex := range order {
		parent := parents[vertex]
		if parent == -1 {
			result[vertex] = 0
		} else {
			result[vertex] = result[parent] + 1
		}
	}
	return result
}

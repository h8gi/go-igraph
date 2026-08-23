package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "tree_construction_cgo.h"
import "C"

import "fmt"

// NoParent marks a root in a parent-vector tree or forest encoding.
const NoParent = -1

// NewTreeFromPrufer constructs the unique undirected labelled tree encoded by
// prufer. A sequence of length n-2 represents a tree on n vertices, so an
// empty sequence constructs the two-vertex tree. Vertex IDs and sequence
// values are zero-based. The input is borrowed only for the call. The returned
// graph is independently owned and must be closed.
//
//igraph:bind igraph_from_prufer
func NewTreeFromPrufer(prufer []int) (*Graph, error) {
	return newTreeFromPrufer(prufer, nil)
}

// PruferSequence returns the Prüfer encoding of a tree on at least two
// vertices. Edge directions are ignored. Non-tree graphs, including empty and
// singleton graphs, are rejected. The returned non-nil slice is Go-owned and
// remains valid after g is closed.
//
//igraph:bind igraph_to_prufer
func (g *Graph) PruferSequence() ([]int, error) {
	return g.pruferSequence(nil)
}

// NewTreeFromParents constructs a rooted tree or forest whose vertices appear
// in parent-vector order. parents[v] is the parent vertex of v; NoParent (-1)
// marks each root. Other negative values, out-of-range IDs, self-parenting, and
// cycles are rejected. TreeOut directs parent-to-child edges, TreeIn directs
// child-to-parent edges, and TreeUndirected creates undirected edges. Empty and
// singleton parent vectors are valid. The input is borrowed only for the call.
// The returned graph is independently owned and must be closed.
//
//igraph:bind igraph_tree_from_parent_vector
func NewTreeFromParents(parents []int, mode TreeMode) (*Graph, error) {
	return newTreeFromParents(parents, mode, nil)
}

// NewSymmetricTree constructs a breadth-first tree whose vertices at level i
// each have branches[i] children. An empty slice constructs a singleton tree;
// every branch count must be positive. The input is borrowed only for the call.
// The returned graph is independently owned and must be closed.
//
//igraph:bind igraph_symmetric_tree
func NewSymmetricTree(branches []int, mode TreeMode) (*Graph, error) {
	return newSymmetricTree(branches, mode, nil)
}

// NewRegularTree constructs a regular rooted tree of the given positive
// height in which every non-leaf vertex has total degree degree. Degree must be
// at least two. Vertices are ordered breadth-first from root 0. The returned
// graph is independently owned and must be closed.
//
//igraph:bind igraph_regular_tree
func NewRegularTree(height, degree int, mode TreeMode) (*Graph, error) {
	return newRegularTree(height, degree, mode, nil)
}

type treeConstructionCallResult struct {
	graph C.igraph_t
	code  int
}

type treeConstructionAdapters struct {
	newInt        func([]int) (*intVector, error)
	closeInt      func(*intVector)
	vectorSlice   func(*intVector) ([]int, error)
	fromPrufer    func(*intVector) treeConstructionCallResult
	toPrufer      func(*Graph, *intVector) int
	fromParents   func(*intVector, TreeMode) treeConstructionCallResult
	symmetricTree func(*intVector, TreeMode) treeConstructionCallResult
	regularTree   func(int, int, TreeMode) treeConstructionCallResult
}

func defaultTreeConstructionAdapters() treeConstructionAdapters {
	return treeConstructionAdapters{
		newInt:      newIntVector,
		closeInt:    (*intVector).close,
		vectorSlice: func(vector *intVector) ([]int, error) { return vector.slice() },
		fromPrufer: func(prufer *intVector) treeConstructionCallResult {
			var graph C.igraph_t
			code := C.go_igraph_from_prufer(&graph, &prufer.value)
			return treeConstructionCallResult{graph: graph, code: int(code)}
		},
		toPrufer: func(graph *Graph, prufer *intVector) int {
			return int(C.go_igraph_to_prufer(&graph.graph, &prufer.value))
		},
		fromParents: func(parents *intVector, mode TreeMode) treeConstructionCallResult {
			cMode, _ := mode.cValue()
			var graph C.igraph_t
			code := C.go_igraph_tree_from_parent_vector(&graph, &parents.value, cMode)
			return treeConstructionCallResult{graph: graph, code: int(code)}
		},
		symmetricTree: func(branches *intVector, mode TreeMode) treeConstructionCallResult {
			cMode, _ := mode.cValue()
			var graph C.igraph_t
			code := C.go_igraph_symmetric_tree(&graph, &branches.value, cMode)
			return treeConstructionCallResult{graph: graph, code: int(code)}
		},
		regularTree: func(height, degree int, mode TreeMode) treeConstructionCallResult {
			cMode, _ := mode.cValue()
			var graph C.igraph_t
			code := C.go_igraph_regular_tree(&graph, C.igraph_int_t(height), C.igraph_int_t(degree), cMode)
			return treeConstructionCallResult{graph: graph, code: int(code)}
		},
	}
}

func resolvedTreeConstructionAdapters(adapters *treeConstructionAdapters) treeConstructionAdapters {
	if adapters == nil {
		return defaultTreeConstructionAdapters()
	}
	return *adapters
}

func newTreeFromPrufer(prufer []int, adapters *treeConstructionAdapters) (*Graph, error) {
	if len(prufer) > int(^uint(0)>>1)-2 {
		return nil, fmt.Errorf("igraph: Prüfer sequence is too large")
	}
	vertexCount := len(prufer) + 2
	if err := validateConstructorSize("Prüfer tree vertex count", vertexCount); err != nil {
		return nil, err
	}
	for index, vertex := range prufer {
		if vertex < 0 || vertex >= vertexCount {
			return nil, fmt.Errorf("igraph: Prüfer vertex at index %d out of range [0, %d): %d", index, vertexCount, vertex)
		}
	}
	resolved := resolvedTreeConstructionAdapters(adapters)
	vector, err := resolved.newInt(prufer)
	if err != nil {
		return nil, err
	}
	defer resolved.closeInt(vector)
	call := resolved.fromPrufer(vector)
	if call.code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("construct tree from Prüfer sequence", call.code)
	}
	return adoptInitializedGraph(&call.graph), nil
}

func (g *Graph) pruferSequence(adapters *treeConstructionAdapters) ([]int, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return nil, ErrClosed
	}
	resolved := resolvedTreeConstructionAdapters(adapters)
	vector, err := resolved.newInt(nil)
	if err != nil {
		return nil, err
	}
	defer resolved.closeInt(vector)
	if code := resolved.toPrufer(g, vector); code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("convert tree to Prüfer sequence", code)
	}
	return resolved.vectorSlice(vector)
}

func validateParentVector(parents []int) error {
	states := make([]uint8, len(parents))
	for vertex, parent := range parents {
		if parent < NoParent || parent >= len(parents) {
			return fmt.Errorf("igraph: parent of vertex %d is invalid: %d", vertex, parent)
		}
		if parent == vertex {
			return fmt.Errorf("igraph: vertex %d cannot be its own parent", vertex)
		}
	}
	for start := range parents {
		vertex := start
		for vertex != NoParent && states[vertex] == 0 {
			states[vertex] = 1
			vertex = parents[vertex]
		}
		if vertex != NoParent && states[vertex] == 1 {
			return fmt.Errorf("igraph: parent vector contains a cycle through vertex %d", vertex)
		}
		vertex = start
		for vertex != NoParent && states[vertex] == 1 {
			states[vertex] = 2
			vertex = parents[vertex]
		}
	}
	return nil
}

func newTreeFromParents(parents []int, mode TreeMode, adapters *treeConstructionAdapters) (*Graph, error) {
	if _, err := mode.cValue(); err != nil {
		return nil, err
	}
	if err := validateConstructorSize("parent vector length", len(parents)); err != nil {
		return nil, err
	}
	if err := validateParentVector(parents); err != nil {
		return nil, err
	}
	resolved := resolvedTreeConstructionAdapters(adapters)
	vector, err := resolved.newInt(parents)
	if err != nil {
		return nil, err
	}
	defer resolved.closeInt(vector)
	call := resolved.fromParents(vector, mode)
	if call.code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("construct tree from parent vector", call.code)
	}
	return adoptInitializedGraph(&call.graph), nil
}

func symmetricTreeVertexCount(branches []int) (int, error) {
	maximum := int(^uint(0) >> 1)
	total, level := 1, 1
	for index, branchCount := range branches {
		if branchCount <= 0 {
			return 0, fmt.Errorf("igraph: branch count at level %d must be positive: %d", index, branchCount)
		}
		if level > maximum/branchCount {
			return 0, fmt.Errorf("igraph: symmetric tree vertex count overflows int")
		}
		level *= branchCount
		if total > maximum-level {
			return 0, fmt.Errorf("igraph: symmetric tree vertex count overflows int")
		}
		total += level
	}
	if err := validateConstructorSize("symmetric tree vertex count", total); err != nil {
		return 0, err
	}
	return total, nil
}

func newSymmetricTree(branches []int, mode TreeMode, adapters *treeConstructionAdapters) (*Graph, error) {
	if _, err := mode.cValue(); err != nil {
		return nil, err
	}
	if _, err := symmetricTreeVertexCount(branches); err != nil {
		return nil, err
	}
	resolved := resolvedTreeConstructionAdapters(adapters)
	vector, err := resolved.newInt(branches)
	if err != nil {
		return nil, err
	}
	defer resolved.closeInt(vector)
	call := resolved.symmetricTree(vector, mode)
	if call.code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("construct symmetric tree", call.code)
	}
	return adoptInitializedGraph(&call.graph), nil
}

func newRegularTree(height, degree int, mode TreeMode, adapters *treeConstructionAdapters) (*Graph, error) {
	if height < 1 {
		return nil, fmt.Errorf("igraph: regular tree height must be positive: %d", height)
	}
	if degree < 2 {
		return nil, fmt.Errorf("igraph: regular tree degree must be at least two: %d", degree)
	}
	if _, err := mode.cValue(); err != nil {
		return nil, err
	}
	if err := validateConstructorSize("regular tree height", height); err != nil {
		return nil, err
	}
	if err := validateConstructorSize("regular tree degree", degree); err != nil {
		return nil, err
	}
	if _, err := regularTreeVertexCount(height, degree); err != nil {
		return nil, fmt.Errorf("igraph: regular tree size is invalid: %w", err)
	}
	resolved := resolvedTreeConstructionAdapters(adapters)
	call := resolved.regularTree(height, degree, mode)
	if call.code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("construct regular tree", call.code)
	}
	return adoptInitializedGraph(&call.graph), nil
}

func regularTreeVertexCount(height, degree int) (int, error) {
	maximum := int(^uint(0) >> 1)
	if degree == 2 {
		if height > (maximum-1)/2 {
			return 0, fmt.Errorf("regular tree vertex count overflows int")
		}
		total := 1 + 2*height
		if err := validateConstructorSize("regular tree vertex count", total); err != nil {
			return 0, err
		}
		return total, nil
	}
	total, level := 1, 1
	for index := 0; index < height; index++ {
		branchCount := degree - 1
		if index == 0 {
			branchCount = degree
		}
		if level > maximum/branchCount {
			return 0, fmt.Errorf("regular tree vertex count overflows int")
		}
		level *= branchCount
		if total > maximum-level {
			return 0, fmt.Errorf("regular tree vertex count overflows int")
		}
		total += level
	}
	if err := validateConstructorSize("regular tree vertex count", total); err != nil {
		return 0, err
	}
	return total, nil
}

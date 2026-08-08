package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
import "C"

import (
	"fmt"
	"runtime"
)

// StarMode controls the direction of edges in a star graph.
type StarMode uint8

const (
	StarOut StarMode = iota
	StarIn
	StarUndirected
	StarMutual
)

func (mode StarMode) cValue() (C.igraph_star_mode_t, error) {
	switch mode {
	case StarOut:
		return C.IGRAPH_STAR_OUT, nil
	case StarIn:
		return C.IGRAPH_STAR_IN, nil
	case StarUndirected:
		return C.IGRAPH_STAR_UNDIRECTED, nil
	case StarMutual:
		return C.IGRAPH_STAR_MUTUAL, nil
	default:
		return 0, fmt.Errorf("igraph: invalid star mode: %d", mode)
	}
}

// TreeMode controls the direction of edges in a tree graph.
type TreeMode uint8

const (
	TreeOut TreeMode = iota
	TreeIn
	TreeUndirected
)

func (mode TreeMode) cValue() (C.igraph_tree_mode_t, error) {
	switch mode {
	case TreeOut:
		return C.IGRAPH_TREE_OUT, nil
	case TreeIn:
		return C.IGRAPH_TREE_IN, nil
	case TreeUndirected:
		return C.IGRAPH_TREE_UNDIRECTED, nil
	default:
		return 0, fmt.Errorf("igraph: invalid tree mode: %d", mode)
	}
}

// NewFull constructs a complete graph. When loops is true, every vertex also
// has a self-loop.
//
//igraph:bind igraph_full
func NewFull(vertexCount int, directed, loops bool) (*Graph, error) {
	if err := validateConstructorSize("vertex count", vertexCount); err != nil {
		return nil, err
	}
	g := &Graph{}
	code := C.igraph_full(&g.graph, C.igraph_int_t(vertexCount), booltoint(directed), booltoint(loops))
	return finishGraphConstruction(g, "initialize full graph", int(code))
}

// NewRing constructs a circular graph. Mutual adds reverse edges to a directed
// ring and has no effect on an undirected ring.
//
//igraph:bind igraph_ring
func NewRing(vertexCount int, directed, mutual bool) (*Graph, error) {
	if err := validateConstructorSize("vertex count", vertexCount); err != nil {
		return nil, err
	}
	g := &Graph{}
	code := C.igraph_ring(&g.graph, C.igraph_int_t(vertexCount), booltoint(directed), booltoint(mutual), booltoint(true))
	return finishGraphConstruction(g, "initialize ring graph", int(code))
}

// NewPath constructs a path graph. Mutual adds reverse edges to a directed
// path and has no effect on an undirected path.
//
//igraph:bind igraph_path_graph
func NewPath(vertexCount int, directed, mutual bool) (*Graph, error) {
	if err := validateConstructorSize("vertex count", vertexCount); err != nil {
		return nil, err
	}
	g := &Graph{}
	code := C.igraph_path_graph(&g.graph, C.igraph_int_t(vertexCount), booltoint(directed), booltoint(mutual))
	return finishGraphConstruction(g, "initialize path graph", int(code))
}

// NewStar constructs a star graph around center. A star must contain at least
// one vertex, and center must identify one of its vertices.
//
//igraph:bind igraph_star
func NewStar(vertexCount, center int, mode StarMode) (*Graph, error) {
	if vertexCount <= 0 {
		return nil, fmt.Errorf("igraph: star vertex count must be positive: %d", vertexCount)
	}
	if center < 0 || center >= vertexCount {
		return nil, fmt.Errorf("igraph: star center %d out of range [0, %d)", center, vertexCount)
	}
	cMode, err := mode.cValue()
	if err != nil {
		return nil, err
	}
	g := &Graph{}
	code := C.igraph_star(&g.graph, C.igraph_int_t(vertexCount), cMode, C.igraph_int_t(center))
	return finishGraphConstruction(g, "initialize star graph", int(code))
}

// NewKaryTree constructs a breadth-first k-ary tree. Children must be
// positive. The graph may contain zero vertices.
//
//igraph:bind igraph_kary_tree
func NewKaryTree(vertexCount, children int, mode TreeMode) (*Graph, error) {
	if err := validateConstructorSize("vertex count", vertexCount); err != nil {
		return nil, err
	}
	if children <= 0 {
		return nil, fmt.Errorf("igraph: tree child count must be positive: %d", children)
	}
	cMode, err := mode.cValue()
	if err != nil {
		return nil, err
	}
	g := &Graph{}
	code := C.igraph_kary_tree(&g.graph, C.igraph_int_t(vertexCount), C.igraph_int_t(children), cMode)
	return finishGraphConstruction(g, "initialize k-ary tree", int(code))
}

// NewHypercube constructs an n-dimensional hypercube.
//
//igraph:bind igraph_hypercube
func NewHypercube(dimension int, directed bool) (*Graph, error) {
	if err := validateConstructorSize("hypercube dimension", dimension); err != nil {
		return nil, err
	}
	g := &Graph{}
	code := C.igraph_hypercube(&g.graph, C.igraph_int_t(dimension), booltoint(directed))
	return finishGraphConstruction(g, "initialize hypercube", int(code))
}

func validateConstructorSize(name string, value int) error {
	if value < 0 {
		return fmt.Errorf("igraph: %s must be non-negative: %d", name, value)
	}
	if int(C.igraph_int_t(value)) != value {
		return fmt.Errorf("igraph: %s is too large: %d", name, value)
	}
	return nil
}

func finishGraphConstruction(g *Graph, operation string, code int) (*Graph, error) {
	if code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError(operation, code)
	}
	runtime.SetFinalizer(g, (*Graph).finalize)
	return g, nil
}

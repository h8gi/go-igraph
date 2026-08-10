package igraph

/*
#include <stdint.h>
#include <igraph.h>
#include "layout_cgo.h"
*/
import "C"
import "fmt"

// DegMode controls edge direction interpretation in layout algorithms.
// It is an alias for DirectionMode (DegOut, DegIn, DegAll).
type DegMode = DirectionMode

const (
	DegOut = DirectionOut
	DegIn  = DirectionIn
	DegAll = DirectionAll
)

// LayoutRandomOptions specifies options for generating a random 2D layout.
type LayoutRandomOptions struct {
	// Seed is an optional seed for thread-safe reproducible execution.
	Seed *uint64
}

// SugiyamaOptions controls parameters for the Sugiyama layered layout.
type SugiyamaOptions struct {
	// HGap is the horizontal gap between vertices.
	HGap float64
	// VGap is the vertical gap between layers.
	VGap float64
	// MaxIter is the maximum number of iterations.
	MaxIter int
	// Weights specifies optional edge weights for crossing reduction.
	Weights []float64
}

// newOrderVertexSelector validates order slice and converts it to cVertexSelector.
// If order is nil, natural vertex ordering is used.
func newOrderVertexSelector(order []int, numVertices int) (*cVertexSelector, error) {
	if order == nil {
		return newCVertexSelector(AllVertices())
	}
	if len(order) != numVertices {
		return nil, fmt.Errorf("igraph: order slice length (%d) does not match vertex count (%d)", len(order), numVertices)
	}
	selector, err := VertexIDs(order...)
	if err != nil {
		return nil, err
	}
	if err := validateVertexSelector(selector, numVertices); err != nil {
		return nil, err
	}
	return newCVertexSelector(selector)
}

// LayoutCircle computes a 2D circular layout for the graph.
// If order is nil, natural vertex ordering is used.
// Public input slices are borrowed and returned values are Go-owned.
//
//igraph:bind igraph_layout_circle
func (g *Graph) LayoutCircle(order []int) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}

	numVertices := int(C.igraph_vcount(&g.graph))
	cOrder, err := newOrderVertexSelector(order, numVertices)
	if err != nil {
		return Matrix{}, err
	}
	defer cOrder.close()

	cMat, err := newCMatrix(Matrix{})
	if err != nil {
		return Matrix{}, err
	}
	defer cMat.close()

	code := C.go_igraph_layout_circle(&g.graph, &cMat.value, cOrder.value)
	if code != C.IGRAPH_SUCCESS {
		return Matrix{}, igraphError("calculate circle layout", int(code))
	}
	return cMat.matrix()
}

// LayoutStar computes a 2D star layout with the given center vertex.
// If order is nil, natural vertex ordering is used.
// Public input slices are borrowed and returned values are Go-owned.
//
//igraph:bind igraph_layout_star
func (g *Graph) LayoutStar(center int, order []int) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}

	numVertices := int(C.igraph_vcount(&g.graph))
	if numVertices == 0 || center < 0 || center >= numVertices {
		return Matrix{}, fmt.Errorf("igraph: center vertex ID %d is out of bounds for graph with %d vertices", center, numVertices)
	}

	var cOrderPtr *C.igraph_vector_int_t
	if order != nil {
		if len(order) != numVertices {
			return Matrix{}, fmt.Errorf("igraph: order slice length (%d) does not match vertex count (%d)", len(order), numVertices)
		}
		for i, id := range order {
			if id < 0 || id >= numVertices {
				return Matrix{}, fmt.Errorf("igraph: vertex ID at order index %d is %d, outside [0, %d)", i, id, numVertices)
			}
		}
		cOrder, err := newIntVector(order)
		if err != nil {
			return Matrix{}, err
		}
		defer cOrder.close()
		cOrderPtr = &cOrder.value
	}

	cMat, err := newCMatrix(Matrix{})
	if err != nil {
		return Matrix{}, err
	}
	defer cMat.close()

	cCenter, err := intToIgraphInt(center, "center vertex ID")
	if err != nil {
		return Matrix{}, err
	}

	code := C.go_igraph_layout_star(&g.graph, &cMat.value, cCenter, cOrderPtr)
	if code != C.IGRAPH_SUCCESS {
		return Matrix{}, igraphError("calculate star layout", int(code))
	}
	return cMat.matrix()
}

// LayoutGrid computes a 2D regular grid layout with the specified row width.
// Public input slices are borrowed and returned values are Go-owned.
//
//igraph:bind igraph_layout_grid
func (g *Graph) LayoutGrid(width int) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}

	if width < 0 {
		return Matrix{}, fmt.Errorf("igraph: width must be non-negative: %d", width)
	}

	cMat, err := newCMatrix(Matrix{})
	if err != nil {
		return Matrix{}, err
	}
	defer cMat.close()

	cWidth, err := intToIgraphInt(width, "grid width")
	if err != nil {
		return Matrix{}, err
	}

	code := C.go_igraph_layout_grid(&g.graph, &cMat.value, cWidth)
	if code != C.IGRAPH_SUCCESS {
		return Matrix{}, igraphError("calculate grid layout", int(code))
	}
	return cMat.matrix()
}

// LayoutRandom computes a 2D random layout where coordinates are uniformly distributed.
// LayoutRandomOptions allows specifying a random seed for thread-safe reproducible execution.
// Public input slices are borrowed and returned values are Go-owned.
//
//igraph:bind igraph_layout_random
func (g *Graph) LayoutRandom(options LayoutRandomOptions) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}

	cMat, err := newCMatrix(Matrix{})
	if err != nil {
		return Matrix{}, err
	}
	defer cMat.close()

	errRNG := withRNG(options.Seed, func() error {
		code := C.go_igraph_layout_random(&g.graph, &cMat.value)
		if code != C.IGRAPH_SUCCESS {
			return igraphError("calculate random layout", int(code))
		}
		return nil
	})
	if errRNG != nil {
		return Matrix{}, errRNG
	}
	return cMat.matrix()
}

// LayoutReingoldTilford computes a 2D tree layout using the Reingold-Tilford algorithm.
// If roots is nil, root vertices are selected automatically.
// Public input slices are borrowed and returned values are Go-owned.
//
//igraph:bind igraph_layout_reingold_tilford
func (g *Graph) LayoutReingoldTilford(mode DegMode, roots []int) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}

	cMode, err := mode.cValue()
	if err != nil {
		return Matrix{}, err
	}

	numVertices := int(C.igraph_vcount(&g.graph))
	var cRootsPtr *C.igraph_vector_int_t
	if roots != nil {
		for i, id := range roots {
			if id < 0 || id >= numVertices {
				return Matrix{}, fmt.Errorf("igraph: root vertex ID at index %d is %d, outside [0, %d)", i, id, numVertices)
			}
		}
		cRoots, err := newIntVector(roots)
		if err != nil {
			return Matrix{}, err
		}
		defer cRoots.close()
		cRootsPtr = &cRoots.value
	}

	cMat, err := newCMatrix(Matrix{})
	if err != nil {
		return Matrix{}, err
	}
	defer cMat.close()

	code := C.go_igraph_layout_reingold_tilford(&g.graph, &cMat.value, cMode, cRootsPtr, nil)
	if code != C.IGRAPH_SUCCESS {
		return Matrix{}, igraphError("calculate Reingold-Tilford layout", int(code))
	}
	return cMat.matrix()
}

// LayoutReingoldTilfordCircular computes a 2D circular tree layout using the Reingold-Tilford algorithm.
// If roots is nil, root vertices are selected automatically.
// Public input slices are borrowed and returned values are Go-owned.
//
//igraph:bind igraph_layout_reingold_tilford_circular
func (g *Graph) LayoutReingoldTilfordCircular(mode DegMode, roots []int) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}

	cMode, err := mode.cValue()
	if err != nil {
		return Matrix{}, err
	}

	numVertices := int(C.igraph_vcount(&g.graph))
	var cRootsPtr *C.igraph_vector_int_t
	if roots != nil {
		for i, id := range roots {
			if id < 0 || id >= numVertices {
				return Matrix{}, fmt.Errorf("igraph: root vertex ID at index %d is %d, outside [0, %d)", i, id, numVertices)
			}
		}
		cRoots, err := newIntVector(roots)
		if err != nil {
			return Matrix{}, err
		}
		defer cRoots.close()
		cRootsPtr = &cRoots.value
	}

	cMat, err := newCMatrix(Matrix{})
	if err != nil {
		return Matrix{}, err
	}
	defer cMat.close()

	code := C.go_igraph_layout_reingold_tilford_circular(&g.graph, &cMat.value, cMode, cRootsPtr, nil)
	if code != C.IGRAPH_SUCCESS {
		return Matrix{}, igraphError("calculate circular Reingold-Tilford layout", int(code))
	}
	return cMat.matrix()
}

// LayoutBipartite computes a 2D layout for bipartite graphs.
// types specifies the partition for each vertex (true for one type, false for the other).
// Public input slices are borrowed and returned values are Go-owned.
//
//igraph:bind igraph_layout_bipartite
func (g *Graph) LayoutBipartite(types []bool, hgap float64, vgap float64, maxIter int) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}

	numVertices := int(C.igraph_vcount(&g.graph))
	if len(types) != numVertices {
		return Matrix{}, fmt.Errorf("igraph: types slice length (%d) does not match vertex count (%d)", len(types), numVertices)
	}

	if maxIter < 0 {
		return Matrix{}, fmt.Errorf("igraph: maxIter must be non-negative: %d", maxIter)
	}

	cTypes, err := newBoolVector(types)
	if err != nil {
		return Matrix{}, err
	}
	defer cTypes.close()

	cMat, err := newCMatrix(Matrix{})
	if err != nil {
		return Matrix{}, err
	}
	defer cMat.close()

	cMaxIter, err := intToIgraphInt(maxIter, "maxIter")
	if err != nil {
		return Matrix{}, err
	}

	code := C.go_igraph_layout_bipartite(
		&g.graph,
		&cTypes.value,
		&cMat.value,
		C.igraph_real_t(hgap),
		C.igraph_real_t(vgap),
		cMaxIter,
	)
	if code != C.IGRAPH_SUCCESS {
		return Matrix{}, igraphError("calculate bipartite layout", int(code))
	}
	return cMat.matrix()
}

// LayoutSugiyama computes a 2D Sugiyama layered layout for directed acyclic graphs.
// If layers is non-nil, it specifies explicit 0-based layer assignments per vertex.
// Public input slices are borrowed and returned values are Go-owned.
//
//igraph:bind igraph_layout_sugiyama
func (g *Graph) LayoutSugiyama(layers []int, options SugiyamaOptions) (Matrix, error) {
	if g == nil {
		return Matrix{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return Matrix{}, ErrClosed
	}

	numVertices := int(C.igraph_vcount(&g.graph))
	var cLayersPtr *C.igraph_vector_int_t
	if layers != nil {
		if len(layers) != numVertices {
			return Matrix{}, fmt.Errorf("igraph: layers slice length (%d) does not match vertex count (%d)", len(layers), numVertices)
		}
		for i, l := range layers {
			if l < 0 {
				return Matrix{}, fmt.Errorf("igraph: layer assignment at index %d must be non-negative: %d", i, l)
			}
		}
		cLayers, err := newIntVector(layers)
		if err != nil {
			return Matrix{}, err
		}
		defer cLayers.close()
		cLayersPtr = &cLayers.value
	}

	if options.MaxIter < 0 {
		return Matrix{}, fmt.Errorf("igraph: maxIter must be non-negative: %d", options.MaxIter)
	}

	weightsVec, err := newOptionalEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return Matrix{}, err
	}
	if weightsVec != nil {
		defer weightsVec.close()
	}

	cMat, err := newCMatrix(Matrix{})
	if err != nil {
		return Matrix{}, err
	}
	defer cMat.close()

	cMaxIter, err := intToIgraphInt(options.MaxIter, "maxIter")
	if err != nil {
		return Matrix{}, err
	}

	code := C.go_igraph_layout_sugiyama(
		&g.graph,
		&cMat.value,
		nil,
		cLayersPtr,
		C.igraph_real_t(options.HGap),
		C.igraph_real_t(options.VGap),
		cMaxIter,
		edgeWeightPointer(weightsVec),
	)
	if code != C.IGRAPH_SUCCESS {
		return Matrix{}, igraphError("calculate Sugiyama layout", int(code))
	}
	return cMat.matrix()
}

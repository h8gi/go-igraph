package igraph

/*
#include <stdint.h>
#include <igraph.h>
#include "layout_cgo.h"
*/
import "C"
import "fmt"

// LayoutRandomOptions specifies options for generating a random 2D layout.
type LayoutRandomOptions struct {
	// Seed is an optional seed for thread-safe reproducible execution.
	Seed *uint64
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

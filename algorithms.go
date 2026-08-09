package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
//
// static igraph_int_t go_igraph_vector_int_list_size(
//     const igraph_vector_int_list_t *list) {
//   return igraph_vector_int_list_size(list);
// }
//
// static const igraph_vector_int_t *go_igraph_vector_int_list_get(
//     const igraph_vector_int_list_t *list, igraph_int_t pos) {
//   return igraph_vector_int_list_get_ptr(list, pos);
// }
import "C"

import (
	"fmt"
)

// DegreeOptions controls how Degree interprets edge direction and self-loops.
// CountLoops uses the standard graph-theoretic convention: a self-loop counts
// twice toward an undirected degree. On a directed graph it counts once toward
// in-degree and once toward out-degree, and therefore twice with DirectionAll.
type DegreeOptions struct {
	Direction  DirectionMode
	CountLoops bool
}

// Degree returns the degree of each selected vertex in selector order.
// Explicit selector duplicates produce duplicate result entries. The returned
// slice is Go-owned and remains valid after the graph is closed. The selector
// and options are read only for the duration of the call and are not retained.
//
//igraph:bind igraph_degree
func (g *Graph) Degree(vertices VertexSelector, options DegreeOptions) ([]int, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}

	mode, err := options.Direction.cValue()
	if err != nil {
		return nil, err
	}
	if err := validateVertexSelector(vertices, int(C.igraph_vcount(&g.graph))); err != nil {
		return nil, err
	}
	selector, err := newCVertexSelector(vertices)
	if err != nil {
		return nil, err
	}
	defer selector.close()
	result, err := newIntVector(nil)
	if err != nil {
		return nil, err
	}
	defer result.close()

	loops := C.igraph_loops_t(C.IGRAPH_NO_LOOPS)
	if options.CountLoops {
		loops = C.igraph_loops_t(C.IGRAPH_LOOPS)
	}
	if code := C.igraph_degree(&g.graph, &result.value, selector.value, mode, loops); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("calculate degree", int(code))
	}
	return result.slice()
}

// NeighborhoodOptions specifies an inclusive distance interval and edge
// direction for neighborhood queries. Order is the maximum distance and
// MinDistance is the minimum; both must be non-negative and MinDistance must
// not exceed Order. A MinDistance of zero includes each selected vertex in its
// own neighborhood.
type NeighborhoodOptions struct {
	Order       int
	MinDistance int
	Direction   DirectionMode
}

// NeighborhoodSizes returns the number of vertices in each selected vertex's
// bounded neighborhood, in selector order. Explicit selector duplicates
// produce duplicate result entries. The returned slice is Go-owned. The
// selector and options are read only for the duration of the call and are not
// retained.
//
//igraph:bind igraph_neighborhood_size
func (g *Graph) NeighborhoodSizes(vertices VertexSelector, options NeighborhoodOptions) ([]int, error) {
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
	result, err := newIntVector(nil)
	if err != nil {
		return nil, err
	}
	defer result.close()

	if code := C.igraph_neighborhood_size(
		&g.graph, &result.value, selector.value, order, mode, minDistance,
	); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("calculate neighborhood sizes", int(code))
	}
	return result.slice()
}

// Neighborhoods returns one bounded neighborhood for each selected vertex, in
// selector order. Both the outer slice and every inner slice are independently
// Go-owned and remain valid after subsequent graph operations or Close. The
// selector and options are read only for the duration of the call and are not
// retained.
//
//igraph:bind igraph_neighborhood
func (g *Graph) Neighborhoods(vertices VertexSelector, options NeighborhoodOptions) ([][]int, error) {
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
	result, err := newIntVectorList()
	if err != nil {
		return nil, err
	}
	defer result.close()

	if code := C.igraph_neighborhood(
		&g.graph, &result.value, selector.value, order, mode, minDistance,
	); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("calculate neighborhoods", int(code))
	}
	return result.slices()
}

func (g *Graph) prepareNeighborhood(
	vertices VertexSelector,
	options NeighborhoodOptions,
) (*cVertexSelector, C.igraph_int_t, C.igraph_int_t, C.igraph_neimode_t, error) {
	mode, err := options.Direction.cValue()
	if err != nil {
		return nil, 0, 0, 0, err
	}
	if options.Order < 0 {
		return nil, 0, 0, 0, fmt.Errorf("igraph: neighborhood order must be non-negative: %d", options.Order)
	}
	if options.MinDistance < 0 {
		return nil, 0, 0, 0, fmt.Errorf("igraph: neighborhood minimum distance must be non-negative: %d", options.MinDistance)
	}
	if options.MinDistance > options.Order {
		return nil, 0, 0, 0, fmt.Errorf(
			"igraph: neighborhood minimum distance %d exceeds order %d",
			options.MinDistance, options.Order,
		)
	}
	// igraph_int_t is a signed 64-bit integer in the pinned upstream release,
	// so every non-negative Go int is representable.
	order := C.igraph_int_t(options.Order)
	minDistance := C.igraph_int_t(options.MinDistance)
	if err := validateVertexSelector(vertices, int(C.igraph_vcount(&g.graph))); err != nil {
		return nil, 0, 0, 0, err
	}
	selector, err := newCVertexSelector(vertices)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	return selector, order, minDistance, mode, nil
}

// intVectorList owns an initialized C list and all integer vectors it contains.
type intVectorList struct {
	value C.igraph_vector_int_list_t
}

//igraph:internal igraph_vector_int_list_init
func newIntVectorList() (*intVectorList, error) {
	result := &intVectorList{}
	if code := C.igraph_vector_int_list_init(&result.value, 0); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("initialize integer vector list", int(code))
	}
	return result, nil
}

//igraph:internal igraph_vector_int_list_size
//igraph:internal igraph_vector_int_list_get_ptr
func (list *intVectorList) slices() ([][]int, error) {
	// The list has one entry per selected vertex, so its size was necessarily
	// representable as a Go slice length at the selection boundary.
	size := int(C.go_igraph_vector_int_list_size(&list.value))
	result := make([][]int, size)
	var err error
	for i := range result {
		vector := C.go_igraph_vector_int_list_get(&list.value, C.igraph_int_t(i))
		result[i], err = intVectorSlice(vector)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

//igraph:internal igraph_vector_int_list_destroy
func (list *intVectorList) close() {
	C.igraph_vector_int_list_destroy(&list.value)
}

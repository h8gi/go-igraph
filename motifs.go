package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "motifs_cgo.h"
//
// static igraph_real_t go_igraph_vector_motif_get(
//     const igraph_vector_t *vector, igraph_int_t pos) {
//   return VECTOR(*vector)[pos];
// }
//
// static igraph_int_t go_igraph_vector_int_motif_get(
//     const igraph_vector_int_t *vector, igraph_int_t pos) {
//   return VECTOR(*vector)[pos];
// }
import "C"

import (
	"fmt"
)

// DyadCensusResult contains the result of dyad census: the count of mutual,
// asymmetric, and null dyads in a directed graph. All fields are non-negative.
type DyadCensusResult struct {
	Mutual     int64
	Asymmetric int64
	Null       int64
}

// DyadCensus computes the dyad census of a directed graph, counting mutual
// (both directions), asymmetric (one direction), and null (no connection) dyads.
// The result is Go-owned.
//
// For undirected graphs, all dyads are counted as null. The result remains
// valid after the graph is closed.
//
//igraph:bind igraph_dyad_census
func (g *Graph) DyadCensus() (DyadCensusResult, error) {
	if g == nil {
		return DyadCensusResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return DyadCensusResult{}, ErrClosed
	}

	var mutual, asymmetric, null C.igraph_real_t
	code := C.go_igraph_dyad_census(&g.graph, &mutual, &asymmetric, &null)
	if code != C.IGRAPH_SUCCESS {
		return DyadCensusResult{}, igraphError("dyad census", int(code))
	}

	return DyadCensusResult{
		Mutual:     int64(mutual),
		Asymmetric: int64(asymmetric),
		Null:       int64(null),
	}, nil
}

// TriadCensus computes the triad census of a graph, returning a 16-element
// slice corresponding to Davis-Leinhardt isomorphism classes. The returned
// slice is Go-owned, non-nil, and has exactly 16 elements. Each element
// represents the count of triads in one isomorphism class. The result remains
// valid after the graph is closed.
//
// For undirected graphs, a subset of the 16 classes have non-zero counts.
//
//igraph:bind igraph_triad_census
func (g *Graph) TriadCensus() ([]int64, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}

	var res C.igraph_vector_t
	code := C.igraph_vector_init(&res, 16)
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("init triad census result", int(code))
	}
	defer C.igraph_vector_destroy(&res)

	code = C.go_igraph_triad_census(&g.graph, &res)
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("triad census", int(code))
	}

	// Convert vector to int64 slice
	size := int(C.igraph_vector_size(&res))
	if size != 16 {
		return nil, fmt.Errorf("igraph: triad census result has unexpected size: %d (expected 16)", size)
	}

	result := make([]int64, 16)
	for i := 0; i < 16; i++ {
		result[i] = int64(C.go_igraph_vector_motif_get(&res, C.igraph_int_t(i)))
	}
	return result, nil
}

// TrianglesCount returns the total number of triangles in the graph. All edges
// in an undirected graph are equally counted. For directed graphs, a triangle
// is defined as three vertices with edges in all directions (a directed cycle
// of length 3), and the count is divided by 3 due to the cyclic symmetry.
//
//igraph:bind igraph_count_triangles
func (g *Graph) TrianglesCount() (int64, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return 0, ErrClosed
	}

	var res C.igraph_real_t
	code := C.go_igraph_count_triangles(&g.graph, &res)
	if code != C.IGRAPH_SUCCESS {
		return 0, igraphError("count triangles", int(code))
	}

	return int64(res), nil
}

// TrianglesList returns all triangles in the graph as a slice of 3-vertex
// indices. Each triangle is represented as a 3-element array [v1, v2, v3]
// where v1 < v2 < v3. The returned slice is Go-owned and non-nil, including
// when no triangles exist. The result remains valid after the graph is closed.
//
// For directed graphs, only directed cycles of length 3 are listed.
//
//igraph:bind igraph_list_triangles
func (g *Graph) TrianglesList() ([][3]int, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}

	var res C.igraph_vector_int_t
	code := C.igraph_vector_int_init(&res, 0)
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("init triangles list result", int(code))
	}
	defer C.igraph_vector_int_destroy(&res)

	code = C.go_igraph_list_triangles(&g.graph, &res)
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("list triangles", int(code))
	}

	// Convert vector_int to slice of [3]int
	size := int(C.igraph_vector_int_size(&res))
	if size%3 != 0 {
		return nil, fmt.Errorf("igraph: triangles list has non-multiple-of-3 size: %d", size)
	}

	triangles := make([][3]int, size/3)
	for i := 0; i < len(triangles); i++ {
		triangles[i][0] = int(C.go_igraph_vector_int_motif_get(&res, C.igraph_int_t(i*3)))
		triangles[i][1] = int(C.go_igraph_vector_int_motif_get(&res, C.igraph_int_t(i*3+1)))
		triangles[i][2] = int(C.go_igraph_vector_int_motif_get(&res, C.igraph_int_t(i*3+2)))
	}

	return triangles, nil
}

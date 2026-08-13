package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "motifs_cgo.h"
import "C"

import (
	"fmt"
	"math"
)

const maximumExactMotifCount = float64(1 << 53)

// DyadCensusResult contains the three Holland-Leinhardt dyad classes. Counts
// are Go-owned non-negative exact integers and remain valid after graph closure.
type DyadCensusResult struct {
	Mutual     int64
	Asymmetric int64
	Null       int64
}

// DyadCensus classifies every unordered vertex pair as mutual, asymmetric, or
// null. Loops and parallel-edge multiplicity are ignored. In an undirected
// graph, connected pairs are mutual and Asymmetric is always zero.
//
//igraph:bind igraph_dyad_census
func (g *Graph) DyadCensus() (DyadCensusResult, error) {
	return g.dyadCensus(nil)
}

// TriadCensus returns the Davis-Leinhardt census in its standard 16-class
// order. The non-nil slice and its exact integer counts are Go-owned and remain
// valid after graph closure. Directed graphs use directed classes; undirected
// graphs occupy the corresponding subset of classes.
//
//igraph:bind igraph_triad_census
func (g *Graph) TriadCensus() ([]int64, error) {
	return g.triadCensus(nil)
}

// AdjacentTrianglesCount returns one triangle count per selected vertex in
// materialized selector order, including duplicates. The selector is borrowed
// only for the synchronous call. Edge directions, multiplicities, and loops
// are ignored; the returned non-nil slice is Go-owned.
//
//igraph:bind igraph_count_adjacent_triangles
func (g *Graph) AdjacentTrianglesCount(vertices VertexSelector) ([]int64, error) {
	return g.adjacentTrianglesCount(vertices, nil)
}

// TrianglesCount returns the number of fully connected vertex triples. Edge
// directions, multiplicities, and loops are ignored.
//
//igraph:bind igraph_count_triangles
func (g *Graph) TrianglesCount() (int64, error) {
	return g.trianglesCount(nil)
}

// TrianglesList returns every fully connected vertex triple exactly once.
// Edge directions, multiplicities, and loops are ignored. Vertex order within
// a triple and outer result order are not compatibility promises. The non-nil
// result is Go-owned and remains valid after graph closure.
//
//igraph:bind igraph_list_triangles
func (g *Graph) TrianglesList() ([][3]int, error) {
	return g.trianglesList(nil)
}

type motifAdapters struct {
	initializeReal func() (*realVector, error)
	closeReal      func(*realVector)
	convertReal    func(*realVector) ([]float64, error)
	initializeInt  func() (*intVector, error)
	closeInt       func(*intVector)
	convertInt     func(*intVector) ([]int, error)
	dyadCall       func(*Graph) ([3]float64, int)
	triadCall      func(*Graph, *realVector) int
	adjacentCall   func(*Graph, *realVector, *cVertexSelector) int
	countCall      func(*Graph) (float64, int)
	listCall       func(*Graph, *intVector) int
}

func defaultMotifAdapters() motifAdapters {
	return motifAdapters{
		initializeReal: func() (*realVector, error) { return newRealVectorSize(0) },
		closeReal:      func(vector *realVector) { vector.close() },
		convertReal:    func(vector *realVector) ([]float64, error) { return vector.slice() },
		initializeInt:  func() (*intVector, error) { return newIntVector(nil) },
		closeInt:       func(vector *intVector) { vector.close() },
		convertInt:     func(vector *intVector) ([]int, error) { return vector.slice() },
		dyadCall: func(g *Graph) ([3]float64, int) {
			var mutual, asymmetric, nullDyads C.igraph_real_t
			code := C.go_igraph_dyad_census(&g.graph, &mutual, &asymmetric, &nullDyads)
			return [3]float64{float64(mutual), float64(asymmetric), float64(nullDyads)}, int(code)
		},
		triadCall: func(g *Graph, result *realVector) int {
			return int(C.go_igraph_triad_census(&g.graph, &result.value))
		},
		adjacentCall: func(g *Graph, result *realVector, vertices *cVertexSelector) int {
			return int(C.go_igraph_count_adjacent_triangles(&g.graph, &result.value, vertices.value))
		},
		countCall: func(g *Graph) (float64, int) {
			var result C.igraph_real_t
			code := C.go_igraph_count_triangles(&g.graph, &result)
			return float64(result), int(code)
		},
		listCall: func(g *Graph, result *intVector) int {
			return int(C.go_igraph_list_triangles(&g.graph, &result.value))
		},
	}
}

func resolvedMotifAdapters(adapters *motifAdapters) motifAdapters {
	if adapters == nil {
		return defaultMotifAdapters()
	}
	return *adapters
}

func (g *Graph) dyadCensus(adapters *motifAdapters) (DyadCensusResult, error) {
	if g == nil {
		return DyadCensusResult{}, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return DyadCensusResult{}, ErrClosed
	}
	resolved := resolvedMotifAdapters(adapters)
	values, code := resolved.dyadCall(g)
	if code != int(C.IGRAPH_SUCCESS) {
		return DyadCensusResult{}, igraphError("calculate dyad census", code)
	}
	converted, err := checkedMotifCounts(values[:], 3, "dyad census")
	if err != nil {
		return DyadCensusResult{}, err
	}
	return DyadCensusResult{Mutual: converted[0], Asymmetric: converted[1], Null: converted[2]}, nil
}

func (g *Graph) triadCensus(adapters *motifAdapters) ([]int64, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	resolved := resolvedMotifAdapters(adapters)
	result, err := resolved.initializeReal()
	if err != nil {
		return nil, err
	}
	defer resolved.closeReal(result)
	if code := resolved.triadCall(g, result); code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("calculate triad census", code)
	}
	values, err := resolved.convertReal(result)
	if err != nil {
		return nil, err
	}
	return checkedMotifCounts(values, 16, "triad census")
}

func (g *Graph) adjacentTrianglesCount(vertices VertexSelector, adapters *motifAdapters) ([]int64, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	vertexCount := int(C.igraph_vcount(&g.graph))
	if err := validateVertexSelector(vertices, vertexCount); err != nil {
		return nil, err
	}
	vertexIDs, err := materializeVertexIDs(&g.graph, vertices)
	if err != nil {
		return nil, fmt.Errorf("igraph: materialize adjacent-triangle selector: %w", err)
	}
	selected, err := VertexIDs(vertexIDs...)
	if err != nil {
		return nil, err
	}
	cVertices, err := newCVertexSelector(selected)
	if err != nil {
		return nil, err
	}
	defer cVertices.close()
	resolved := resolvedMotifAdapters(adapters)
	result, err := resolved.initializeReal()
	if err != nil {
		return nil, err
	}
	defer resolved.closeReal(result)
	if code := resolved.adjacentCall(g, result, cVertices); code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("count adjacent triangles", code)
	}
	values, err := resolved.convertReal(result)
	if err != nil {
		return nil, err
	}
	return checkedMotifCounts(values, len(vertexIDs), "adjacent triangle")
}

func (g *Graph) trianglesCount(adapters *motifAdapters) (int64, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return 0, ErrClosed
	}
	resolved := resolvedMotifAdapters(adapters)
	value, code := resolved.countCall(g)
	if code != int(C.IGRAPH_SUCCESS) {
		return 0, igraphError("count triangles", code)
	}
	return checkedMotifCount(value, "triangle count")
}

func (g *Graph) trianglesList(adapters *motifAdapters) ([][3]int, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	resolved := resolvedMotifAdapters(adapters)
	result, err := resolved.initializeInt()
	if err != nil {
		return nil, err
	}
	defer resolved.closeInt(result)
	if code := resolved.listCall(g, result); code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("list triangles", code)
	}
	values, err := resolved.convertInt(result)
	if err != nil {
		return nil, err
	}
	if len(values)%3 != 0 {
		return nil, fmt.Errorf("igraph: triangle list length %d is not divisible by three", len(values))
	}
	triangles := make([][3]int, len(values)/3)
	for index := range triangles {
		copy(triangles[index][:], values[index*3:index*3+3])
	}
	return triangles, nil
}

func checkedMotifCounts(values []float64, expected int, description string) ([]int64, error) {
	if len(values) != expected {
		return nil, fmt.Errorf("igraph: %s result length %d does not match expected length %d", description, len(values), expected)
	}
	result := make([]int64, len(values))
	for index, value := range values {
		converted, err := checkedMotifCount(value, fmt.Sprintf("%s count at index %d", description, index))
		if err != nil {
			return nil, err
		}
		result[index] = converted
	}
	return result, nil
}

func checkedMotifCount(value float64, description string) (int64, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || math.Trunc(value) != value {
		return 0, fmt.Errorf("igraph: %s is not a finite non-negative integer: %g", description, value)
	}
	if value > maximumExactMotifCount {
		return 0, fmt.Errorf("igraph: %s exceeds the exact integer range of C double: %g", description, value)
	}
	return int64(value), nil
}

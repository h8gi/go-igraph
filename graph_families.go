package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "graph_families_cgo.h"
import "C"

import "fmt"

// WheelMode controls the direction of rim and spoke edges in a wheel.
type WheelMode uint8

const (
	WheelOut WheelMode = iota
	WheelIn
	WheelUndirected
	WheelMutual
)

func (mode WheelMode) cValue() (C.igraph_wheel_mode_t, error) {
	switch mode {
	case WheelOut:
		return C.IGRAPH_WHEEL_OUT, nil
	case WheelIn:
		return C.IGRAPH_WHEEL_IN, nil
	case WheelUndirected:
		return C.IGRAPH_WHEEL_UNDIRECTED, nil
	case WheelMutual:
		return C.IGRAPH_WHEEL_MUTUAL, nil
	default:
		return 0, fmt.Errorf("igraph: invalid wheel mode: %d", mode)
	}
}

// MultipartiteGraphResult contains a graph and its vertex-ID-aligned,
// zero-based part indexes. Both values are independently Go-owned; Graph must
// be closed.
type MultipartiteGraphResult struct {
	Graph *Graph
	Parts []int
}

// NewCirculant constructs G(vertexCount, shifts). For each shift s and vertex
// v it connects v to (v+s) modulo vertexCount. Shifts are normalized modulo
// the vertex count; duplicates and zero shifts are ignored. Undirected inverse
// shifts are also deduplicated. The input is borrowed only for the call.
//
//igraph:bind igraph_circulant
func NewCirculant(vertexCount int, shifts []int, directed bool) (*Graph, error) {
	if err := validateConstructorSize("circulant vertex count", vertexCount); err != nil {
		return nil, err
	}
	if vertexCount > 0 && len(shifts) > int(^uint(0)>>1)/vertexCount/2 {
		return nil, fmt.Errorf("igraph: circulant edge capacity overflows int")
	}
	v, err := newIntVector(shifts)
	if err != nil {
		return nil, err
	}
	defer v.close()
	var graph C.igraph_t
	code := C.go_igraph_circulant(&graph, C.igraph_int_t(vertexCount), &v.value, booltoint(directed))
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("construct circulant graph", int(code))
	}
	return adoptInitializedGraph(&graph), nil
}

// NewWheel constructs a wheel with center as its hub. Rim vertices retain
// vertex-ID order with center omitted. WheelOut points spokes from the center,
// WheelIn points them inward, WheelMutual creates reverse pairs, and
// WheelUndirected creates an undirected graph.
//
//igraph:bind igraph_wheel
func NewWheel(vertexCount, center int, mode WheelMode) (*Graph, error) {
	if vertexCount <= 0 {
		return nil, fmt.Errorf("igraph: wheel vertex count must be positive: %d", vertexCount)
	}
	if err := validateConstructorSize("wheel vertex count", vertexCount); err != nil {
		return nil, err
	}
	if center < 0 || center >= vertexCount {
		return nil, fmt.Errorf("igraph: wheel center %d out of range [0, %d)", center, vertexCount)
	}
	cMode, err := mode.cValue()
	if err != nil {
		return nil, err
	}
	var graph C.igraph_t
	if vertexCount == 1 {
		code := C.igraph_empty(&graph, 1, booltoint(mode != WheelUndirected))
		if code != C.IGRAPH_SUCCESS {
			return nil, igraphError("construct singleton wheel graph", int(code))
		}
		return adoptInitializedGraph(&graph), nil
	}
	code := C.go_igraph_wheel(&graph, C.igraph_int_t(vertexCount), cMode, C.igraph_int_t(center))
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("construct wheel graph", int(code))
	}
	return adoptInitializedGraph(&graph), nil
}

// NewGeneralizedPetersen constructs G(n,k). Vertices 0..n-1 form the outer
// cycle and n..2n-1 form the inner circulant; corresponding vertices are
// joined by spokes. n must be at least three and 0 < 2k < n.
//
//igraph:bind igraph_generalized_petersen
func NewGeneralizedPetersen(n, k int) (*Graph, error) {
	if n < 3 {
		return nil, fmt.Errorf("igraph: Petersen size must be at least three: %d", n)
	}
	if err := validateConstructorSize("Petersen size", n); err != nil {
		return nil, err
	}
	maximum := int(^uint(0) >> 1)
	if n > maximum/6 {
		return nil, fmt.Errorf("igraph: Petersen graph size overflows int")
	}
	if k <= 0 || k >= n || k > (n-1)/2 {
		return nil, fmt.Errorf("igraph: Petersen shift must satisfy 0 < 2k < n: %d", k)
	}
	var graph C.igraph_t
	code := C.go_igraph_generalized_petersen(&graph, C.igraph_int_t(n), C.igraph_int_t(k))
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("construct generalized Petersen graph", int(code))
	}
	return adoptInitializedGraph(&graph), nil
}

func validatePartSizes(sizes []int) (int, error) {
	total := 0
	for index, size := range sizes {
		if size < 0 {
			return 0, fmt.Errorf("igraph: part size at index %d must be non-negative: %d", index, size)
		}
		if total > int(^uint(0)>>1)-size {
			return 0, fmt.Errorf("igraph: multipartite vertex count overflows int")
		}
		total += size
	}
	if err := validateConstructorSize("multipartite vertex count", total); err != nil {
		return 0, err
	}
	return total, nil
}

// NewFullMultipartite constructs a complete multipartite graph. Vertices are
// grouped by part index and retain order within each part. For directed graphs,
// DirectionOut points low-part to high-part, DirectionIn reverses them, and
// DirectionAll creates both. PartSizes is borrowed only for the call.
//
//igraph:bind igraph_full_multipartite
func NewFullMultipartite(partSizes []int, directed bool, direction DirectionMode) (MultipartiteGraphResult, error) {
	if _, err := validatePartSizes(partSizes); err != nil {
		return MultipartiteGraphResult{}, err
	}
	cDirection, err := direction.cValue()
	if err != nil {
		return MultipartiteGraphResult{}, err
	}
	sizes, err := newIntVector(partSizes)
	if err != nil {
		return MultipartiteGraphResult{}, err
	}
	defer sizes.close()
	types, err := newIntVector(nil)
	if err != nil {
		return MultipartiteGraphResult{}, err
	}
	defer types.close()
	var graph C.igraph_t
	code := C.go_igraph_full_multipartite(&graph, &types.value, &sizes.value, booltoint(directed), cDirection)
	if code != C.IGRAPH_SUCCESS {
		return MultipartiteGraphResult{}, igraphError("construct complete multipartite graph", int(code))
	}
	parts, err := types.slice()
	if err != nil {
		C.igraph_destroy(&graph)
		return MultipartiteGraphResult{}, err
	}
	return MultipartiteGraphResult{Graph: adoptInitializedGraph(&graph), Parts: parts}, nil
}

// NewTuran constructs the undirected Turán graph on vertexCount vertices with
// positive partCount. Parts are as equal as possible and lower-index parts
// receive the extra vertices. If partCount exceeds vertexCount, the non-empty
// parts form a complete graph.
//
//igraph:bind igraph_turan
func NewTuran(vertexCount, partCount int) (MultipartiteGraphResult, error) {
	if err := validateConstructorSize("Turán vertex count", vertexCount); err != nil {
		return MultipartiteGraphResult{}, err
	}
	if partCount <= 0 {
		return MultipartiteGraphResult{}, fmt.Errorf("igraph: Turán part count must be positive: %d", partCount)
	}
	if err := validateConstructorSize("Turán part count", partCount); err != nil {
		return MultipartiteGraphResult{}, err
	}
	types, err := newIntVector(nil)
	if err != nil {
		return MultipartiteGraphResult{}, err
	}
	defer types.close()
	var graph C.igraph_t
	code := C.go_igraph_turan(&graph, &types.value, C.igraph_int_t(vertexCount), C.igraph_int_t(partCount))
	if code != C.IGRAPH_SUCCESS {
		return MultipartiteGraphResult{}, igraphError("construct Turán graph", int(code))
	}
	parts, err := types.slice()
	if err != nil {
		C.igraph_destroy(&graph)
		return MultipartiteGraphResult{}, err
	}
	return MultipartiteGraphResult{Graph: adoptInitializedGraph(&graph), Parts: parts}, nil
}

// NewFullCitation constructs edges i→j for every j<i. When directed is false
// it produces an undirected complete graph. Vertex IDs retain chronological
// order from oldest (0) to newest (vertexCount-1).
//
//igraph:bind igraph_full_citation
func NewFullCitation(vertexCount int, directed bool) (*Graph, error) {
	if err := validateConstructorSize("citation vertex count", vertexCount); err != nil {
		return nil, err
	}
	if vertexCount > 1 && vertexCount > int(^uint(0)>>1)/(vertexCount-1) {
		return nil, fmt.Errorf("igraph: citation edge capacity overflows int")
	}
	var graph C.igraph_t
	code := C.go_igraph_full_citation(&graph, C.igraph_int_t(vertexCount), booltoint(directed))
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("construct full citation graph", int(code))
	}
	return adoptInitializedGraph(&graph), nil
}

// LineGraph constructs L(g). Original edge i becomes vertex i. For an
// undirected graph, distinct edges are adjacent for each shared endpoint. For
// a directed graph, i→j exists when edge i's target equals edge j's source.
// The returned graph is independently owned and must be closed.
//
//igraph:bind igraph_linegraph
func (g *Graph) LineGraph() (*Graph, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return nil, ErrClosed
	}
	var graph C.igraph_t
	code := C.go_igraph_linegraph(&g.graph, &graph)
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("construct line graph", int(code))
	}
	return adoptInitializedGraph(&graph), nil
}

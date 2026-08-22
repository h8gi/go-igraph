package igraph

/*
#cgo pkg-config: igraph
#include <igraph.h>
#include "scan_cgo.h"
*/
import "C"

import "fmt"

// LocalScanOptions controls a same-graph local scan. Radius must be
// non-negative. Radius zero returns degree, or strength with weights. For
// larger radii the result is the edge count or weight sum in each induced
// neighborhood. Weights are borrowed only for the synchronous call and must
// contain one finite value per edge.
type LocalScanOptions struct {
	Radius    int
	Direction DirectionMode
	Weights   []float64
}

// SubsetLocalScanOptions controls induced-edge scans for caller-supplied
// vertex subsets. Weights are borrowed only for the synchronous call and must
// contain one finite value per edge.
type SubsetLocalScanOptions struct {
	Weights []float64
}

type localScanHooks struct {
	newResult func([]float64) (*realVector, error)
	newList   func() (*intVectorList, error)
	newVector func([]int) (*intVector, error)
	append    func(*intVectorList, *intVector) error
	run       func() error
}

// LocalScan returns one Go-owned value per vertex in vertex-ID order.
// Options and weights are borrowed only for the call.
//
//igraph:bind igraph_local_scan_0
//igraph:bind igraph_local_scan_1_ecount
//igraph:bind igraph_local_scan_k_ecount
func (g *Graph) LocalScan(options LocalScanOptions) ([]float64, error) {
	return g.localScan(options, localScanHooks{})
}

func (g *Graph) localScan(options LocalScanOptions, hooks localScanHooks) ([]float64, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	if options.Radius < 0 {
		return nil, fmt.Errorf("igraph: local scan radius must be non-negative: %d", options.Radius)
	}
	radius, err := intToIgraphInt(options.Radius, "local scan radius")
	if err != nil {
		return nil, err
	}
	mode, err := options.Direction.cValue()
	if err != nil {
		return nil, err
	}
	weights, err := newOptionalEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return nil, err
	}
	if weights != nil {
		defer weights.close()
	}
	newResult := hooks.newResult
	if newResult == nil {
		newResult = newRealVector
	}
	result, err := newResult(nil)
	if err != nil {
		return nil, err
	}
	defer result.close()
	run := func() error {
		var code C.igraph_error_t
		switch options.Radius {
		case 0:
			code = C.go_igraph_local_scan_0(&g.graph, &result.value, edgeWeightPointer(weights), mode)
		case 1:
			code = C.go_igraph_local_scan_1_ecount(&g.graph, &result.value, edgeWeightPointer(weights), mode)
		default:
			code = C.go_igraph_local_scan_k_ecount(&g.graph, radius, &result.value, edgeWeightPointer(weights), mode)
		}
		if code != C.IGRAPH_SUCCESS {
			return igraphError("calculate local scan", int(code))
		}
		return nil
	}
	if hooks.run != nil {
		run = hooks.run
	}
	if err := run(); err != nil {
		return nil, err
	}
	return result.slice()
}

// LocalScanSubsets returns the edge count or weight sum induced by each subset
// in exact input order. The outer and nested slices are borrowed only for the
// call. Empty and repeated subsets are valid; duplicate IDs within one subset
// are rejected. The returned non-nil slice is independently Go-owned.
//
//igraph:bind igraph_local_scan_subset_ecount
func (g *Graph) LocalScanSubsets(subsets [][]int, options SubsetLocalScanOptions) ([]float64, error) {
	return g.localScanSubsets(subsets, options, localScanHooks{})
}

func (g *Graph) localScanSubsets(subsets [][]int, options SubsetLocalScanOptions, hooks localScanHooks) ([]float64, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}
	vertexCount := int(C.igraph_vcount(&g.graph))
	if _, err := intToIgraphInt(len(subsets), "local scan subset count"); err != nil {
		return nil, err
	}
	for subsetIndex, subset := range subsets {
		seen := make(map[int]struct{}, len(subset))
		for valueIndex, vertex := range subset {
			if vertex < 0 || vertex >= vertexCount {
				return nil, fmt.Errorf("igraph: local scan subset %d vertex at index %d out of range: %d", subsetIndex, valueIndex, vertex)
			}
			if _, exists := seen[vertex]; exists {
				return nil, fmt.Errorf("igraph: local scan subset %d contains duplicate vertex %d", subsetIndex, vertex)
			}
			seen[vertex] = struct{}{}
		}
	}
	weights, err := newOptionalEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return nil, err
	}
	if weights != nil {
		defer weights.close()
	}
	newList := hooks.newList
	if newList == nil {
		newList = newIntVectorList
	}
	list, err := newList()
	if err != nil {
		return nil, err
	}
	defer list.close()
	newVector := hooks.newVector
	if newVector == nil {
		newVector = newIntVector
	}
	appendVector := hooks.append
	if appendVector == nil {
		appendVector = func(list *intVectorList, vector *intVector) error {
			if code := C.go_igraph_scan_list_append_copy(&list.value, &vector.value); code != C.IGRAPH_SUCCESS {
				return igraphError("copy local scan subset", int(code))
			}
			return nil
		}
	}
	for index, subset := range subsets {
		vector, err := newVector(subset)
		if err != nil {
			return nil, fmt.Errorf("igraph: initialize local scan subset %d: %w", index, err)
		}
		err = appendVector(list, vector)
		vector.close()
		if err != nil {
			return nil, fmt.Errorf("igraph: append local scan subset %d: %w", index, err)
		}
	}
	newResult := hooks.newResult
	if newResult == nil {
		newResult = newRealVector
	}
	result, err := newResult(nil)
	if err != nil {
		return nil, err
	}
	defer result.close()
	run := func() error {
		if code := C.go_igraph_local_scan_subset_ecount(&g.graph, &result.value, edgeWeightPointer(weights), &list.value); code != C.IGRAPH_SUCCESS {
			return igraphError("calculate subset local scan", int(code))
		}
		return nil
	}
	if hooks.run != nil {
		run = hooks.run
	}
	if err := run(); err != nil {
		return nil, err
	}
	return result.slice()
}

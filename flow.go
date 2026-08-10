package igraph

/*
#include <igraph.h>
#include "flow_cgo.h"
*/
import "C"
import (
	"fmt"
	"math"
)

// MaxFlowResult represents the result of a maximum flow computation.
type MaxFlowResult struct {
	// Value is the total maximum flow value.
	Value float64
	// Flow is a slice containing the flow on each edge.
	Flow []float64
	// Cut contains the IDs of edges forming the minimum s-t cut.
	Cut []int
	// Partition contains the vertex IDs in the source component of the cut.
	Partition []int
	// Partition2 contains the vertex IDs in the target component of the cut.
	Partition2 []int
}

// MinCutResult represents the result of a global minimum cut computation.
type MinCutResult struct {
	// Value is the total capacity of the minimum cut.
	Value float64
	// Cut contains the IDs of edges forming the minimum cut.
	Cut []int
	// Partition contains the vertex IDs on one side of the cut.
	Partition []int
	// Partition2 contains the vertex IDs on the other side of the cut.
	Partition2 []int
}

// STMinCutResult represents the result of a source-target minimum cut computation.
type STMinCutResult struct {
	// Value is the total capacity of the minimum cut.
	Value float64
	// Cut contains the IDs of edges forming the minimum cut.
	Cut []int
	// Partition contains the vertex IDs in the source component.
	Partition []int
	// Partition2 contains the vertex IDs in the target component.
	Partition2 []int
}

func validateCapacities(g *Graph, capacities []float64) (*realVector, error) {
	if capacities == nil {
		return nil, nil
	}
	numEdges := int(C.igraph_ecount(&g.graph))
	if len(capacities) != numEdges {
		return nil, fmt.Errorf("igraph: capacities slice length (%d) must match number of edges (%d)", len(capacities), numEdges)
	}
	for i, c := range capacities {
		if math.IsNaN(c) || c < 0 {
			return nil, fmt.Errorf("igraph: capacity value at index %d must be non-negative: %v", i, c)
		}
	}
	return newRealVector(capacities)
}

func validateVertex(g *Graph, v int, name string) (C.igraph_int_t, error) {
	numVertices := int(C.igraph_vcount(&g.graph))
	if v < 0 || v >= numVertices {
		return 0, fmt.Errorf("igraph: %s vertex ID out of range [0, %d): %d", name, numVertices, v)
	}
	return intToIgraphInt(v, name)
}

func validateSourceTarget(g *Graph, source, target int) (C.igraph_int_t, C.igraph_int_t, error) {
	src, err := validateVertex(g, source, "source")
	if err != nil {
		return 0, 0, err
	}
	tgt, err := validateVertex(g, target, "target")
	if err != nil {
		return 0, 0, err
	}
	if source == target {
		return 0, 0, fmt.Errorf("igraph: source (%d) and target (%d) must be distinct", source, target)
	}
	return src, tgt, nil
}

// MaxFlow calculates the maximum flow between a source and a target vertex using optional edge capacities.
// If capacities is nil, unit capacity (1.0) is used for all edges.
//
//igraph:bind igraph_maxflow
func (g *Graph) MaxFlow(source, target int, capacities []float64) (*MaxFlowResult, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return nil, ErrClosed
	}

	src, tgt, err := validateSourceTarget(g, source, target)
	if err != nil {
		return nil, err
	}

	capVec, err := validateCapacities(g, capacities)
	if err != nil {
		return nil, err
	}
	if capVec != nil {
		defer capVec.close()
	}

	flowVec, _ := newRealVectorSize(0)
	defer flowVec.close()

	cutVec, _ := newIntVector(nil)
	defer cutVec.close()

	partVec, _ := newIntVector(nil)
	defer partVec.close()

	part2Vec, _ := newIntVector(nil)
	defer part2Vec.close()

	var value C.igraph_real_t
	var capPtr *C.igraph_vector_t
	if capVec != nil {
		capPtr = &capVec.value
	}

	code := C.go_igraph_maxflow(
		&g.graph,
		&value,
		&flowVec.value,
		&cutVec.value,
		&partVec.value,
		&part2Vec.value,
		src,
		tgt,
		capPtr,
		nil,
	)
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("igraph_maxflow", int(code))
	}

	flowSlice, _ := flowVec.slice()
	cutSlice, _ := cutVec.slice()
	partSlice, _ := partVec.slice()
	part2Slice, _ := part2Vec.slice()

	return &MaxFlowResult{
		Value:      float64(value),
		Flow:       flowSlice,
		Cut:        cutSlice,
		Partition:  partSlice,
		Partition2: part2Slice,
	}, nil
}

// MaxFlowValue calculates only the maximum flow value between a source and a target vertex.
//
//igraph:bind igraph_maxflow_value
func (g *Graph) MaxFlowValue(source, target int, capacities []float64) (float64, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return 0, ErrClosed
	}

	src, tgt, err := validateSourceTarget(g, source, target)
	if err != nil {
		return 0, err
	}

	capVec, err := validateCapacities(g, capacities)
	if err != nil {
		return 0, err
	}
	if capVec != nil {
		defer capVec.close()
	}

	var value C.igraph_real_t
	var capPtr *C.igraph_vector_t
	if capVec != nil {
		capPtr = &capVec.value
	}

	code := C.go_igraph_maxflow_value(
		&g.graph,
		&value,
		src,
		tgt,
		capPtr,
		nil,
	)
	if code != C.IGRAPH_SUCCESS {
		return 0, igraphError("igraph_maxflow_value", int(code))
	}

	return float64(value), nil
}

// STMinCut computes the minimum s-t cut between source and target.
//
//igraph:bind igraph_st_mincut
func (g *Graph) STMinCut(source, target int, capacities []float64) (*STMinCutResult, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return nil, ErrClosed
	}

	src, tgt, err := validateSourceTarget(g, source, target)
	if err != nil {
		return nil, err
	}

	capVec, err := validateCapacities(g, capacities)
	if err != nil {
		return nil, err
	}
	if capVec != nil {
		defer capVec.close()
	}

	cutVec, _ := newIntVector(nil)
	defer cutVec.close()

	partVec, _ := newIntVector(nil)
	defer partVec.close()

	part2Vec, _ := newIntVector(nil)
	defer part2Vec.close()

	var value C.igraph_real_t
	var capPtr *C.igraph_vector_t
	if capVec != nil {
		capPtr = &capVec.value
	}

	code := C.go_igraph_st_mincut(
		&g.graph,
		&value,
		&cutVec.value,
		&partVec.value,
		&part2Vec.value,
		src,
		tgt,
		capPtr,
	)
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("igraph_st_mincut", int(code))
	}

	cutSlice, _ := cutVec.slice()
	partSlice, _ := partVec.slice()
	part2Slice, _ := part2Vec.slice()

	return &STMinCutResult{
		Value:      float64(value),
		Cut:        cutSlice,
		Partition:  partSlice,
		Partition2: part2Slice,
	}, nil
}

// STMinCutValue calculates only the value of the minimum s-t cut.
//
//igraph:bind igraph_st_mincut_value
func (g *Graph) STMinCutValue(source, target int, capacities []float64) (float64, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return 0, ErrClosed
	}

	src, tgt, err := validateSourceTarget(g, source, target)
	if err != nil {
		return 0, err
	}

	capVec, err := validateCapacities(g, capacities)
	if err != nil {
		return 0, err
	}
	if capVec != nil {
		defer capVec.close()
	}

	var value C.igraph_real_t
	var capPtr *C.igraph_vector_t
	if capVec != nil {
		capPtr = &capVec.value
	}

	code := C.go_igraph_st_mincut_value(
		&g.graph,
		&value,
		src,
		tgt,
		capPtr,
	)
	if code != C.IGRAPH_SUCCESS {
		return 0, igraphError("igraph_st_mincut_value", int(code))
	}

	return float64(value), nil
}

// MinCut computes the global minimum cut of the graph.
//
//igraph:bind igraph_mincut
func (g *Graph) MinCut(capacities []float64) (*MinCutResult, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return nil, ErrClosed
	}

	capVec, err := validateCapacities(g, capacities)
	if err != nil {
		return nil, err
	}
	if capVec != nil {
		defer capVec.close()
	}

	partVec, _ := newIntVector(nil)
	defer partVec.close()

	part2Vec, _ := newIntVector(nil)
	defer part2Vec.close()

	cutVec, _ := newIntVector(nil)
	defer cutVec.close()

	var value C.igraph_real_t
	var capPtr *C.igraph_vector_t
	if capVec != nil {
		capPtr = &capVec.value
	}

	code := C.go_igraph_mincut(
		&g.graph,
		&value,
		&partVec.value,
		&part2Vec.value,
		&cutVec.value,
		capPtr,
	)
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("igraph_mincut", int(code))
	}

	partSlice, _ := partVec.slice()
	part2Slice, _ := part2Vec.slice()
	cutSlice, _ := cutVec.slice()

	return &MinCutResult{
		Value:      float64(value),
		Cut:        cutSlice,
		Partition:  partSlice,
		Partition2: part2Slice,
	}, nil
}

// MinCutValue calculates only the value of the global minimum cut.
//
//igraph:bind igraph_mincut_value
func (g *Graph) MinCutValue(capacities []float64) (float64, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return 0, ErrClosed
	}

	capVec, err := validateCapacities(g, capacities)
	if err != nil {
		return 0, err
	}
	if capVec != nil {
		defer capVec.close()
	}

	var value C.igraph_real_t
	var capPtr *C.igraph_vector_t
	if capVec != nil {
		capPtr = &capVec.value
	}

	code := C.go_igraph_mincut_value(
		&g.graph,
		&value,
		capPtr,
	)
	if code != C.IGRAPH_SUCCESS {
		return 0, igraphError("igraph_mincut_value", int(code))
	}

	return float64(value), nil
}

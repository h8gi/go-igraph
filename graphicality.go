package igraph

/*
#include <igraph.h>
#include "graphicality_cgo.h"
*/
import "C"

import "fmt"

// EdgeType specifies allowed edge types for degree sequence graphicality checks.
type EdgeType uint8

const (
	// EdgeTypeSimple allows simple edges only (no self-loops, no multiple edges).
	EdgeTypeSimple EdgeType = iota
	// EdgeTypeLoops allows self-loops but no multiple edges.
	EdgeTypeLoops
	// EdgeTypeMulti allows multiple edges between distinct vertices but no self-loops.
	EdgeTypeMulti
	// EdgeTypeLoopsAndMulti allows both self-loops and multiple edges.
	EdgeTypeLoopsAndMulti
)

func (et EdgeType) cValue() (C.igraph_edge_type_sw_t, error) {
	switch et {
	case EdgeTypeSimple:
		return C.IGRAPH_SIMPLE_SW, nil
	case EdgeTypeLoops:
		return C.IGRAPH_LOOPS_SW, nil
	case EdgeTypeMulti:
		return C.IGRAPH_MULTI_SW, nil
	case EdgeTypeLoopsAndMulti:
		return C.IGRAPH_LOOPS_SW | C.IGRAPH_MULTI_SW, nil
	default:
		return 0, fmt.Errorf("igraph: invalid edge type: %d", et)
	}
}

// IsGraphical checks whether the given degree sequence(s) can be realized by a graph
// under the specified allowed edge types.
//
// A nil or empty inDeg checks the undirected form; otherwise outDeg holds
// out-degrees and inDeg holds in-degrees of a directed sequence pair.
//
// Input slices are borrowed for the duration of the call.
//
//igraph:bind igraph_is_graphical
func IsGraphical(outDeg []int, inDeg []int, edgeTypes EdgeType) (bool, error) {
	cEdgeTypes, err := edgeTypes.cValue()
	if err != nil {
		return false, err
	}

	inVec, err := newOptionalInDegrees(outDeg, inDeg)
	if err != nil {
		return false, err
	}
	if inVec != nil {
		defer inVec.close()
	}

	outVec, err := newIntVector(outDeg)
	if err != nil {
		return false, err
	}
	defer outVec.close()

	var inVecPtr *C.igraph_vector_int_t
	if inVec != nil {
		inVecPtr = &inVec.value
	}

	var res C.igraph_bool_t
	if code := C.go_igraph_is_graphical(&outVec.value, inVecPtr, cEdgeTypes, &res); code != C.IGRAPH_SUCCESS {
		return false, igraphError("igraph_is_graphical", int(code))
	}
	return res != booltoint(false), nil
}

// IsBigraphical checks whether two degree sequences can be realized by a bipartite graph
// under the specified allowed edge types.
//
// Bipartite graphs cannot contain self-loops, so the self-loop component of
// edgeTypes is ignored: EdgeTypeLoops behaves like EdgeTypeSimple and
// EdgeTypeLoopsAndMulti behaves like EdgeTypeMulti.
//
// Input slices are borrowed for the duration of the call.
//
//igraph:bind igraph_is_bigraphical
func IsBigraphical(deg1 []int, deg2 []int, edgeTypes EdgeType) (bool, error) {
	cEdgeTypes, err := edgeTypes.cValue()
	if err != nil {
		return false, err
	}
	vec1, err := newIntVector(deg1)
	if err != nil {
		return false, err
	}
	defer vec1.close()

	vec2, err := newIntVector(deg2)
	if err != nil {
		return false, err
	}
	defer vec2.close()

	var res C.igraph_bool_t
	if code := C.go_igraph_is_bigraphical(&vec1.value, &vec2.value, cEdgeTypes, &res); code != C.IGRAPH_SUCCESS {
		return false, igraphError("igraph_is_bigraphical", int(code))
	}
	return res != booltoint(false), nil
}

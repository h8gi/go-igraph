package igraph

/*
#include <igraph.h>
#include "cut_enumeration_cgo.h"
*/
import "C"
import "fmt"

// STCut represents a source-target cut consisting of cut edge IDs and source component vertex IDs.
type STCut struct {
	// Cut contains the edge IDs included in the cut set.
	Cut []int
	// Partition contains the vertex IDs in the source component of the cut.
	Partition []int
}

// AllSTCuts finds all source-target cuts between a source and a target vertex.
// Source and target vertex IDs are validated.
// The returned slice of STCut structs is Go-owned and remains valid after the graph is closed.
//
//igraph:bind igraph_all_st_cuts
func (g *Graph) AllSTCuts(source, target int) ([]STCut, error) {
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

	cutsList, err := newIntVectorList()
	if err != nil {
		return nil, err
	}
	defer cutsList.close()

	partsList, err := newIntVectorList()
	if err != nil {
		return nil, err
	}
	defer partsList.close()

	code := C.go_igraph_all_st_cuts(&g.graph, &cutsList.value, &partsList.value, src, tgt)
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("igraph_all_st_cuts", int(code))
	}

	cutSlices, err := cutsList.slices()
	if err != nil {
		return nil, err
	}
	partSlices, err := partsList.slices()
	if err != nil {
		return nil, err
	}

	if len(cutSlices) != len(partSlices) {
		return nil, fmt.Errorf("igraph: all_st_cuts returned mismatched cuts (%d) and partitions (%d)", len(cutSlices), len(partSlices))
	}

	results := make([]STCut, len(cutSlices))
	for i := range results {
		results[i] = STCut{
			Cut:       cutSlices[i],
			Partition: partSlices[i],
		}
	}

	return results, nil
}

// AllSTMincuts finds all minimum source-target cuts between a source and a target vertex.
// Source and target vertex IDs are validated.
// The capacities slice is borrowed only for the call and copied into C storage.
// The returned mincut value and slice of STCut structs are Go-owned and remain valid after graph closure.
//
//igraph:bind igraph_all_st_mincuts
func (g *Graph) AllSTMincuts(source, target int, capacities []float64) (float64, []STCut, error) {
	if g == nil {
		return 0, nil, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return 0, nil, ErrClosed
	}

	src, tgt, err := validateSourceTarget(g, source, target)
	if err != nil {
		return 0, nil, err
	}

	capVec, err := validateCapacities(g, capacities)
	if err != nil {
		return 0, nil, err
	}
	if capVec != nil {
		defer capVec.close()
	}

	cutsList, err := newIntVectorList()
	if err != nil {
		return 0, nil, err
	}
	defer cutsList.close()

	partsList, err := newIntVectorList()
	if err != nil {
		return 0, nil, err
	}
	defer partsList.close()

	var value C.igraph_real_t
	var capPtr *C.igraph_vector_t
	if capVec != nil {
		capPtr = &capVec.value
	}

	code := C.go_igraph_all_st_mincuts(&g.graph, &value, &cutsList.value, &partsList.value, src, tgt, capPtr)
	if code != C.IGRAPH_SUCCESS {
		return 0, nil, igraphError("igraph_all_st_mincuts", int(code))
	}

	cutSlices, err := cutsList.slices()
	if err != nil {
		return 0, nil, err
	}
	partSlices, err := partsList.slices()
	if err != nil {
		return 0, nil, err
	}

	if len(cutSlices) != len(partSlices) {
		return 0, nil, fmt.Errorf("igraph: all_st_mincuts returned mismatched cuts (%d) and partitions (%d)", len(cutSlices), len(partSlices))
	}

	results := make([]STCut, len(cutSlices))
	for i := range results {
		results[i] = STCut{
			Cut:       cutSlices[i],
			Partition: partSlices[i],
		}
	}

	return float64(value), results, nil
}

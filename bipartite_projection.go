package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
import "C"

import "fmt"

// BipartiteMode identifies one side of an explicit bipartite partition.
type BipartiteMode bool

const (
	BipartiteModeFalse BipartiteMode = false
	BipartiteModeTrue  BipartiteMode = true
)

// BipartiteProjectionSize describes one projected graph without constructing
// it. Counts are Go-owned scalar values.
type BipartiteProjectionSize struct {
	Vertices int
	Edges    int
}

// BipartiteProjectionSizes contains size estimates for both named modes.
type BipartiteProjectionSizes struct {
	False BipartiteProjectionSize
	True  BipartiteProjectionSize
}

// BipartiteProjectionResult contains one independently owned projection.
// SourceVertexIDs[result vertex ID] is the corresponding source graph vertex
// ID. Multiplicities[result edge ID] is the number of distinct opposite-mode
// common neighbours producing that edge. Both slices are non-nil and Go-owned
// and remain valid after either graph is closed. Graph must be closed by the
// caller.
type BipartiteProjectionResult struct {
	Graph           *Graph
	SourceVertexIDs []int
	Multiplicities  []int
}

// BipartiteProjectionsResult contains independently owned projections of both
// explicitly named modes. Either graph may be closed without affecting the
// other graph or any returned slice.
type BipartiteProjectionsResult struct {
	False BipartiteProjectionResult
	True  BipartiteProjectionResult
}

// BipartiteProjectionSizes reports the projected vertex and edge counts for
// both modes without constructing either projection. Edge directions are
// ignored. Partition is borrowed only for the synchronous call.
//
//igraph:bind igraph_bipartite_projection_size
func (g *Graph) BipartiteProjectionSizes(partition BipartitePartition) (BipartiteProjectionSizes, error) {
	if g == nil {
		return BipartiteProjectionSizes{}, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return BipartiteProjectionSizes{}, ErrClosed
	}
	if err := validateProjectionPartitionLocked(g, partition); err != nil {
		return BipartiteProjectionSizes{}, err
	}
	types, err := newBoolVector([]bool(partition))
	if err != nil {
		return BipartiteProjectionSizes{}, err
	}
	defer types.close()
	var falseVertices, falseEdges, trueVertices, trueEdges C.igraph_int_t
	code := C.igraph_bipartite_projection_size(
		&g.graph, &types.value,
		&falseVertices, &falseEdges, &trueVertices, &trueEdges,
	)
	if code != C.IGRAPH_SUCCESS {
		return BipartiteProjectionSizes{}, igraphError("calculate bipartite projection sizes", int(code))
	}
	fv, err := igraphIntToInt(falseVertices, "false-mode projection vertex count")
	if err != nil {
		return BipartiteProjectionSizes{}, err
	}
	fe, err := igraphIntToInt(falseEdges, "false-mode projection edge count")
	if err != nil {
		return BipartiteProjectionSizes{}, err
	}
	tv, err := igraphIntToInt(trueVertices, "true-mode projection vertex count")
	if err != nil {
		return BipartiteProjectionSizes{}, err
	}
	te, err := igraphIntToInt(trueEdges, "true-mode projection edge count")
	if err != nil {
		return BipartiteProjectionSizes{}, err
	}
	return BipartiteProjectionSizes{
		False: BipartiteProjectionSize{Vertices: fv, Edges: fe},
		True:  BipartiteProjectionSize{Vertices: tv, Edges: te},
	}, nil
}

// BipartiteProjection constructs only the requested mode's projection. Edge
// directions are ignored. Partition is borrowed only for the synchronous call;
// the returned graph and slices are independently Go-owned.
func (g *Graph) BipartiteProjection(partition BipartitePartition, mode BipartiteMode) (BipartiteProjectionResult, error) {
	return g.bipartiteProjection(partition, mode, nil)
}

// BipartiteProjections constructs both named projections in one call. Edge
// directions are ignored. Partition is borrowed only for the synchronous call;
// both graphs and all slices are independently Go-owned.
//
//igraph:bind igraph_bipartite_projection
func (g *Graph) BipartiteProjections(partition BipartitePartition) (BipartiteProjectionsResult, error) {
	return g.bipartiteProjections(partition, nil)
}

type projectionCallResult struct {
	first, second C.igraph_t
	code          int
}

type projectionAdapters struct {
	newBool    func([]bool) (*boolVector, error)
	newInt     func([]int) (*intVector, error)
	convertInt func(*intVector) ([]int, error)
	call       func(*Graph, *boolVector, *intVector, *intVector, bool, bool, int) projectionCallResult
	closeGraph func(*Graph) error
}

func defaultProjectionAdapters() projectionAdapters {
	return projectionAdapters{
		newBool: newBoolVector, newInt: newIntVector, convertInt: (*intVector).slice,
		call: func(g *Graph, types *boolVector, firstMultiplicity, secondMultiplicity *intVector, wantFirst, wantSecond bool, probe int) projectionCallResult {
			var result projectionCallResult
			var first, second *C.igraph_t
			var firstMult, secondMult *C.igraph_vector_int_t
			if wantFirst {
				first = &result.first
				firstMult = &firstMultiplicity.value
			}
			if wantSecond {
				second = &result.second
				secondMult = &secondMultiplicity.value
			}
			result.code = int(C.igraph_bipartite_projection(&g.graph, &types.value, first, second, firstMult, secondMult, C.igraph_int_t(probe)))
			return result
		},
		closeGraph: (*Graph).Close,
	}
}

func resolvedProjectionAdapters(adapters *projectionAdapters) projectionAdapters {
	if adapters == nil {
		return defaultProjectionAdapters()
	}
	return *adapters
}

func validateProjectionPartitionLocked(g *Graph, partition BipartitePartition) error {
	if err := validateBipartitePartitionLength(partition, int(C.igraph_vcount(&g.graph))); err != nil {
		return err
	}
	return validatePartitionEdgesLocked(g, partition)
}

func projectionSourceIDs(partition BipartitePartition, mode BipartiteMode) []int {
	result := make([]int, 0)
	for id, value := range partition {
		if value == bool(mode) {
			result = append(result, id)
		}
	}
	return result
}

func (g *Graph) bipartiteProjection(partition BipartitePartition, mode BipartiteMode, adapters *projectionAdapters) (BipartiteProjectionResult, error) {
	if mode != BipartiteModeFalse && mode != BipartiteModeTrue {
		return BipartiteProjectionResult{}, fmt.Errorf("igraph: invalid bipartite mode")
	}
	if g == nil {
		return BipartiteProjectionResult{}, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return BipartiteProjectionResult{}, ErrClosed
	}
	if err := validateProjectionPartitionLocked(g, partition); err != nil {
		return BipartiteProjectionResult{}, err
	}
	resolved := resolvedProjectionAdapters(adapters)
	types, err := resolved.newBool([]bool(partition))
	if err != nil {
		return BipartiteProjectionResult{}, err
	}
	defer types.close()
	multiplicity, err := resolved.newInt(nil)
	if err != nil {
		return BipartiteProjectionResult{}, err
	}
	defer multiplicity.close()
	probe := -1
	for id, value := range partition {
		if value == bool(mode) {
			probe = id
			break
		}
	}
	// An empty requested mode is unambiguous without a probe: request the slot
	// determined by the partition's false/true ordering.
	wantFirst := mode == BipartiteModeFalse
	wantSecond := !wantFirst
	if probe >= 0 {
		wantFirst, wantSecond = true, false
	}
	call := resolved.call(g, types, multiplicity, multiplicity, wantFirst, wantSecond, probe)
	if call.code != int(C.IGRAPH_SUCCESS) {
		return BipartiteProjectionResult{}, igraphError("construct bipartite projection", call.code)
	}
	graphValue := &call.first
	if wantSecond {
		graphValue = &call.second
	}
	graph := adoptInitializedGraph(graphValue)
	multiplicities, err := resolved.convertInt(multiplicity)
	if err != nil {
		_ = resolved.closeGraph(graph)
		return BipartiteProjectionResult{}, err
	}
	return BipartiteProjectionResult{Graph: graph, SourceVertexIDs: projectionSourceIDs(partition, mode), Multiplicities: multiplicities}, nil
}

func (g *Graph) bipartiteProjections(partition BipartitePartition, adapters *projectionAdapters) (BipartiteProjectionsResult, error) {
	if g == nil {
		return BipartiteProjectionsResult{}, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return BipartiteProjectionsResult{}, ErrClosed
	}
	if err := validateProjectionPartitionLocked(g, partition); err != nil {
		return BipartiteProjectionsResult{}, err
	}
	resolved := resolvedProjectionAdapters(adapters)
	types, err := resolved.newBool([]bool(partition))
	if err != nil {
		return BipartiteProjectionsResult{}, err
	}
	defer types.close()
	falseMultiplicity, err := resolved.newInt(nil)
	if err != nil {
		return BipartiteProjectionsResult{}, err
	}
	defer falseMultiplicity.close()
	trueMultiplicity, err := resolved.newInt(nil)
	if err != nil {
		return BipartiteProjectionsResult{}, err
	}
	defer trueMultiplicity.close()
	call := resolved.call(g, types, falseMultiplicity, trueMultiplicity, true, true, -1)
	if call.code != int(C.IGRAPH_SUCCESS) {
		return BipartiteProjectionsResult{}, igraphError("construct bipartite projections", call.code)
	}
	falseGraph := adoptInitializedGraph(&call.first)
	trueGraph := adoptInitializedGraph(&call.second)
	fm, err := resolved.convertInt(falseMultiplicity)
	if err != nil {
		_ = resolved.closeGraph(falseGraph)
		_ = resolved.closeGraph(trueGraph)
		return BipartiteProjectionsResult{}, err
	}
	tm, err := resolved.convertInt(trueMultiplicity)
	if err != nil {
		_ = resolved.closeGraph(falseGraph)
		_ = resolved.closeGraph(trueGraph)
		return BipartiteProjectionsResult{}, err
	}
	return BipartiteProjectionsResult{
		False: BipartiteProjectionResult{Graph: falseGraph, SourceVertexIDs: projectionSourceIDs(partition, BipartiteModeFalse), Multiplicities: fm},
		True:  BipartiteProjectionResult{Graph: trueGraph, SourceVertexIDs: projectionSourceIDs(partition, BipartiteModeTrue), Multiplicities: tm},
	}, nil
}

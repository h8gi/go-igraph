package igraph

/*
#include <igraph.h>
*/
import "C"

import (
	"fmt"
	"math"
)

// BipartiteRandomOptions controls the orientation and reproducibility of a
// random bipartite graph. When Directed is true, DirectionOut orients edges
// from the false mode to the true mode, DirectionIn reverses them, and
// DirectionAll samples the two orientations independently. Direction is
// validated but otherwise ignored for undirected graphs.
type BipartiteRandomOptions struct {
	Directed  bool
	Direction DirectionMode
	// Seed optionally seeds the package random number generator.
	Seed *uint64
}

// NewBipartiteGNP samples a simple bipartite graph in which every possible
// edge is present independently with probability p. Self-loops and parallel
// edges are never generated.
//
// The false-mode vertices precede the true-mode vertices. The returned graph
// is independently owned and must be closed by the caller; its non-nil
// partition is Go-owned. Options are borrowed only for the synchronous call.
//
//igraph:bind igraph_bipartite_game_gnp
func NewBipartiteGNP(falseModeSize, trueModeSize int, p float64, options BipartiteRandomOptions) (BipartiteGraphResult, error) {
	if math.IsNaN(p) || p < 0 || p > 1 {
		return BipartiteGraphResult{}, fmt.Errorf("igraph: bipartite edge probability must be in [0, 1]: %v", p)
	}
	return randomBipartiteGraph(falseModeSize, trueModeSize, options, "sample bipartite G(n,p) graph", nil, func(graph *C.igraph_t, types *C.igraph_vector_bool_t, direction C.igraph_neimode_t) C.igraph_error_t {
		return C.igraph_bipartite_game_gnp(graph, types, C.igraph_int_t(falseModeSize), C.igraph_int_t(trueModeSize), C.igraph_real_t(p), booltoint(options.Directed), direction, C.IGRAPH_SIMPLE_SW, booltoint(false))
	})
}

// NewBipartiteGNM uniformly samples a simple bipartite graph with exactly
// edgeCount edges. Self-loops and parallel edges are never generated.
//
// The false-mode vertices precede the true-mode vertices. The returned graph
// is independently owned and must be closed by the caller; its non-nil
// partition is Go-owned. Options are borrowed only for the synchronous call.
//
//igraph:bind igraph_bipartite_game_gnm
func NewBipartiteGNM(falseModeSize, trueModeSize, edgeCount int, options BipartiteRandomOptions) (BipartiteGraphResult, error) {
	if err := validateConstructorSize("bipartite edge count", edgeCount); err != nil {
		return BipartiteGraphResult{}, err
	}
	return randomBipartiteGraph(falseModeSize, trueModeSize, options, "sample bipartite G(n,m) graph", func() error {
		maximum, err := bipartiteSimpleEdgeLimit(falseModeSize, trueModeSize, options)
		if err != nil {
			return err
		}
		if edgeCount > maximum {
			return fmt.Errorf("igraph: bipartite edge count %d exceeds maximum %d", edgeCount, maximum)
		}
		return nil
	}, func(graph *C.igraph_t, types *C.igraph_vector_bool_t, direction C.igraph_neimode_t) C.igraph_error_t {
		return C.igraph_bipartite_game_gnm(graph, types, C.igraph_int_t(falseModeSize), C.igraph_int_t(trueModeSize), C.igraph_int_t(edgeCount), booltoint(options.Directed), direction, C.IGRAPH_SIMPLE_SW, booltoint(false))
	})
}

// NewBipartiteIEA generates a bipartite multigraph by independently assigning
// each of edgeCount labeled draws to a uniformly selected endpoint pair.
// Parallel edges may be generated; self-loops cannot be generated.
//
// IEA does not sample multigraphs uniformly. The false-mode vertices precede
// the true-mode vertices. The returned graph is independently owned and must
// be closed by the caller; its non-nil partition is Go-owned. Options are
// borrowed only for the synchronous call. This binds an experimental API in
// pinned igraph 1.0.1.
//
//igraph:bind igraph_bipartite_iea_game
func NewBipartiteIEA(falseModeSize, trueModeSize, edgeCount int, options BipartiteRandomOptions) (BipartiteGraphResult, error) {
	if err := validateConstructorSize("bipartite edge count", edgeCount); err != nil {
		return BipartiteGraphResult{}, err
	}
	return randomBipartiteGraph(falseModeSize, trueModeSize, options, "sample bipartite IEA graph", func() error {
		if edgeCount > 0 && (falseModeSize == 0 || trueModeSize == 0) {
			return fmt.Errorf("igraph: positive bipartite edge count requires both modes to be non-empty")
		}
		return nil
	}, func(graph *C.igraph_t, types *C.igraph_vector_bool_t, direction C.igraph_neimode_t) C.igraph_error_t {
		return C.igraph_bipartite_iea_game(graph, types, C.igraph_int_t(falseModeSize), C.igraph_int_t(trueModeSize), C.igraph_int_t(edgeCount), booltoint(options.Directed), direction)
	})
}

type randomBipartiteAdapters struct {
	newBool     func([]bool) (*boolVector, error)
	convertBool func(*boolVector) ([]bool, error)
	closeBool   func(*boolVector)
	call        func(*boolVector, DirectionMode) bipartiteGraphCallResult
}

func randomBipartiteGraph(n1, n2 int, options BipartiteRandomOptions, operation string, validate func() error, generate func(*C.igraph_t, *C.igraph_vector_bool_t, C.igraph_neimode_t) C.igraph_error_t) (BipartiteGraphResult, error) {
	return randomBipartiteGraphWithAdapters(n1, n2, options, operation, validate, &randomBipartiteAdapters{
		newBool: newBoolVector, convertBool: (*boolVector).slice, closeBool: (*boolVector).close,
		call: func(types *boolVector, direction DirectionMode) bipartiteGraphCallResult {
			var graph C.igraph_t
			cDirection, _ := direction.cValue()
			code := generate(&graph, &types.value, cDirection)
			return bipartiteGraphCallResult{graph: graph, code: int(code)}
		},
	})
}

func randomBipartiteGraphWithAdapters(n1, n2 int, options BipartiteRandomOptions, operation string, validate func() error, adapters *randomBipartiteAdapters) (BipartiteGraphResult, error) {
	if err := validateBipartiteModeSizes(n1, n2); err != nil {
		return BipartiteGraphResult{}, err
	}
	if _, err := options.Direction.cValue(); err != nil {
		return BipartiteGraphResult{}, err
	}
	if validate != nil {
		if err := validate(); err != nil {
			return BipartiteGraphResult{}, err
		}
	}
	types, err := adapters.newBool(nil)
	if err != nil {
		return BipartiteGraphResult{}, err
	}
	defer adapters.closeBool(types)
	var call bipartiteGraphCallResult
	err = withRNG(options.Seed, func() error {
		call = adapters.call(types, options.Direction)
		if call.code != int(C.IGRAPH_SUCCESS) {
			return igraphError(operation, call.code)
		}
		return nil
	})
	if err != nil {
		return BipartiteGraphResult{}, err
	}
	partition, err := adapters.convertBool(types)
	if err != nil {
		C.igraph_destroy(&call.graph)
		return BipartiteGraphResult{}, err
	}
	return BipartiteGraphResult{Graph: adoptInitializedGraph(&call.graph), Partition: BipartitePartition(partition)}, nil
}

func validateBipartiteModeSizes(n1, n2 int) error {
	if err := validateConstructorSize("false-mode vertex count", n1); err != nil {
		return err
	}
	if err := validateConstructorSize("true-mode vertex count", n2); err != nil {
		return err
	}
	maximum := int(^uint(0) >> 1)
	if n1 > maximum-n2 {
		return fmt.Errorf("igraph: total bipartite vertex count is too large")
	}
	return validateConstructorSize("total bipartite vertex count", n1+n2)
}

func bipartiteSimpleEdgeLimit(n1, n2 int, options BipartiteRandomOptions) (int, error) {
	maximum := int(^uint(0) >> 1)
	factor := 1
	if options.Directed && options.Direction == DirectionAll {
		factor = 2
	}
	if n1 != 0 && n2 > maximum/n1 {
		return 0, fmt.Errorf("igraph: bipartite edge capacity is too large")
	}
	result := n1 * n2
	if result > maximum/factor {
		return 0, fmt.Errorf("igraph: bipartite edge capacity is too large")
	}
	return result * factor, nil
}

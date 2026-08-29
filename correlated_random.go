package igraph

/*
#include <igraph.h>
#include "random_games_cgo.h"
*/
import "C"

import (
	"fmt"
	"math"
)

// CorrelatedGraphOptions controls reproducibility and optional relabeling. If
// Permutation is non-empty, element i is the source-graph vertex that becomes
// vertex i in the generated graph. The slice is borrowed synchronously.
type CorrelatedGraphOptions struct {
	Seed        *uint64
	Permutation []int
}

// CorrelatedGraphPair contains two independently Go-owned graphs. Each graph
// must be closed, in either order.
type CorrelatedGraphPair struct {
	First  *Graph
	Second *Graph
}

func validateCorrelation(correlation, edgeProbability float64) error {
	if math.IsNaN(correlation) || math.IsInf(correlation, 0) || correlation < 0 || correlation > 1 {
		return fmt.Errorf("igraph: adjacency correlation must be finite and in [0, 1]: %g", correlation)
	}
	if math.IsNaN(edgeProbability) || math.IsInf(edgeProbability, 0) || edgeProbability <= 0 || edgeProbability >= 1 {
		return fmt.Errorf("igraph: edge probability must be finite and in (0, 1): %g", edgeProbability)
	}
	return nil
}

func correlatedPermutation(vertexCount int, values []int) (*intVector, *C.igraph_vector_int_t, error) {
	if len(values) == 0 {
		return nil, nil, nil
	}
	if len(values) != vertexCount {
		return nil, nil, fmt.Errorf("igraph: permutation length %d does not match vertex count %d", len(values), vertexCount)
	}
	seen := make([]bool, vertexCount)
	for i, v := range values {
		if v < 0 || v >= vertexCount {
			return nil, nil, fmt.Errorf("igraph: permutation value %d at index %d is outside [0, %d)", v, i, vertexCount)
		}
		if seen[v] {
			return nil, nil, fmt.Errorf("igraph: permutation repeats vertex %d", v)
		}
		seen[v] = true
	}
	vector, err := newIntVector(values)
	if err != nil {
		return nil, nil, err
	}
	return vector, &vector.value, nil
}

// CorrelatedGame returns a simple graph sampled by retaining/deleting source
// adjacencies and adding missing adjacencies so their Pearson correlation has
// the requested target under the supplied marginal edgeProbability. Diagonal
// entries are excluded. Source must be simple; its direction is preserved.
// The source is read-locked through sampling. The result has no copied graph
// attributes, is independently Go-owned, and remains usable after source closes.
//
//igraph:bind igraph_correlated_game
func (source *Graph) CorrelatedGame(correlation, edgeProbability float64, options CorrelatedGraphOptions) (*Graph, error) {
	if source == nil {
		return nil, ErrClosed
	}
	source.mu.RLock()
	defer source.mu.RUnlock()
	if source.closed {
		return nil, ErrClosed
	}
	if err := validateCorrelation(correlation, edgeProbability); err != nil {
		return nil, err
	}
	vertexCount := int(C.igraph_vcount(&source.graph))
	permutation, pointer, err := correlatedPermutation(vertexCount, options.Permutation)
	if err != nil {
		return nil, err
	}
	if permutation != nil {
		defer permutation.close()
	}
	var result C.igraph_t
	err = withRNG(options.Seed, func() error {
		code := C.go_igraph_correlated_game(&result, &source.graph, C.igraph_real_t(correlation), C.igraph_real_t(edgeProbability), pointer)
		if code != C.IGRAPH_SUCCESS {
			return igraphError("igraph_correlated_game", int(code))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return adoptInitializedGraph(&result), nil
}

// CorrelatedPairGame samples two simple Erdős-Rényi graphs with the same
// marginal edgeProbability and target adjacency correlation. If Permutation is
// present, its element i names the First vertex that becomes vertex i in Second.
// Both results are independently owned and have the requested direction.
//
//igraph:bind igraph_correlated_pair_game
func CorrelatedPairGame(vertexCount int, correlation, edgeProbability float64, directed bool, options CorrelatedGraphOptions) (CorrelatedGraphPair, error) {
	if err := validateConstructorSize("vertex count", vertexCount); err != nil {
		return CorrelatedGraphPair{}, err
	}
	if err := validateCorrelation(correlation, edgeProbability); err != nil {
		return CorrelatedGraphPair{}, err
	}
	permutation, pointer, err := correlatedPermutation(vertexCount, options.Permutation)
	if err != nil {
		return CorrelatedGraphPair{}, err
	}
	if permutation != nil {
		defer permutation.close()
	}
	var first, second C.igraph_t
	var firstInitialized, secondInitialized C.igraph_bool_t
	err = withRNG(options.Seed, func() error {
		code := C.go_igraph_correlated_pair_game(&first, &second, C.igraph_int_t(vertexCount), C.igraph_real_t(correlation), C.igraph_real_t(edgeProbability), booltoint(directed), pointer, &firstInitialized, &secondInitialized)
		if code != C.IGRAPH_SUCCESS {
			return igraphError("igraph_correlated_pair_game", int(code))
		}
		return nil
	})
	if err != nil {
		if bool(secondInitialized) {
			C.igraph_destroy(&second)
		}
		if bool(firstInitialized) {
			C.igraph_destroy(&first)
		}
		return CorrelatedGraphPair{}, err
	}
	return CorrelatedGraphPair{First: adoptInitializedGraph(&first), Second: adoptInitializedGraph(&second)}, nil
}

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

// ErdosRenyiOptions contains options for the Erdős-Rényi random graph models.
type ErdosRenyiOptions struct {
	// Seed optionally seeds the package random number generator.
	Seed *uint64
}

// KRegularOptions contains options for the k-regular random graph model.
type KRegularOptions struct {
	// Seed optionally seeds the package random number generator.
	Seed *uint64
}

// TreeGameOptions contains options for uniform random tree sampling.
type TreeGameOptions struct {
	// Seed optionally seeds the package random number generator.
	Seed *uint64
}

// TreeGameMethod selects the sampling algorithm used by RandomTreeGame.
type TreeGameMethod uint8

const (
	// TreeGamePrufer samples uniformly through random Prüfer sequences.
	// It supports undirected trees only.
	TreeGamePrufer TreeGameMethod = iota
	// TreeGameLERW samples uniformly through loop-erased random walks and
	// supports directed and undirected trees.
	TreeGameLERW
)

func (method TreeGameMethod) cValue() (C.igraph_random_tree_t, error) {
	switch method {
	case TreeGamePrufer:
		return C.IGRAPH_RANDOM_TREE_PRUFER, nil
	case TreeGameLERW:
		return C.IGRAPH_RANDOM_TREE_LERW, nil
	default:
		return 0, fmt.Errorf("igraph: invalid tree game method: %d", method)
	}
}

func allowedEdgeTypes(loops bool) C.igraph_edge_type_sw_t {
	if loops {
		return C.IGRAPH_LOOPS_SW
	}
	return C.IGRAPH_SIMPLE_SW
}

// generateGraph runs generate under the package RNG lock and adopts the
// initialized C graph into an independently owned public Graph.
func generateGraph(operation string, seed *uint64, generate func(graph *C.igraph_t) C.igraph_error_t) (*Graph, error) {
	var graph C.igraph_t
	errRNG := withRNG(seed, func() error {
		if code := generate(&graph); code != C.IGRAPH_SUCCESS {
			return igraphError(operation, int(code))
		}
		return nil
	})
	if errRNG != nil {
		return nil, errRNG
	}
	return adoptInitializedGraph(&graph), nil
}

// ErdosRenyiGNM samples a uniform random graph with exactly n vertices and m
// edges from the G(n,m) Erdős-Rényi model. When loops is true, self-loops may
// be sampled; parallel edges are never produced. The requested edge count must
// fit the chosen model, otherwise the upstream sampler reports an error.
//
// The returned graph is independently Go-owned and must be closed by the
// caller. Options are read only for the duration of the call. A non-nil
// options.Seed makes the sample reproducible under the package RNG contract.
//
//igraph:bind igraph_erdos_renyi_game_gnm
func ErdosRenyiGNM(n int, m int, directed bool, loops bool, options ErdosRenyiOptions) (*Graph, error) {
	if err := validateConstructorSize("vertex count", n); err != nil {
		return nil, err
	}
	if err := validateConstructorSize("edge count", m); err != nil {
		return nil, err
	}
	return generateGraph("igraph_erdos_renyi_game_gnm", options.Seed, func(graph *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_erdos_renyi_game_gnm(
			graph,
			C.igraph_int_t(n),
			C.igraph_int_t(m),
			booltoint(directed),
			allowedEdgeTypes(loops),
			booltoint(false),
		)
	})
}

// ErdosRenyiGNP samples a random graph with n vertices where every possible
// edge is included independently with probability p, following the G(n,p)
// Erdős-Rényi model. When loops is true, self-loops are sampled with the same
// probability; parallel edges are never produced.
//
// The returned graph is independently Go-owned and must be closed by the
// caller. Options are read only for the duration of the call. A non-nil
// options.Seed makes the sample reproducible under the package RNG contract.
//
//igraph:bind igraph_erdos_renyi_game_gnp
func ErdosRenyiGNP(n int, p float64, directed bool, loops bool, options ErdosRenyiOptions) (*Graph, error) {
	if err := validateConstructorSize("vertex count", n); err != nil {
		return nil, err
	}
	if math.IsNaN(p) || p < 0 || p > 1 {
		return nil, fmt.Errorf("igraph: edge probability must be in [0, 1]: %v", p)
	}
	return generateGraph("igraph_erdos_renyi_game_gnp", options.Seed, func(graph *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_erdos_renyi_game_gnp(
			graph,
			C.igraph_int_t(n),
			C.igraph_real_t(p),
			booltoint(directed),
			allowedEdgeTypes(loops),
			booltoint(false),
		)
	})
}

// KRegularGame samples a random graph with n vertices where every vertex has
// degree k. On a directed graph both the in- and out-degree of every vertex
// equal k. When multiple is false the sample is a simple graph and k must be
// smaller than n; when multiple is true parallel edges may be produced. An
// undirected sample requires n*k to be even.
//
// The returned graph is independently Go-owned and must be closed by the
// caller. Options are read only for the duration of the call. A non-nil
// options.Seed makes the sample reproducible under the package RNG contract.
//
//igraph:bind igraph_k_regular_game
func KRegularGame(n int, k int, directed bool, multiple bool, options KRegularOptions) (*Graph, error) {
	if err := validateConstructorSize("vertex count", n); err != nil {
		return nil, err
	}
	if err := validateConstructorSize("vertex degree", k); err != nil {
		return nil, err
	}
	if !multiple && n > 0 && k >= n {
		return nil, fmt.Errorf("igraph: simple k-regular degree %d must be smaller than vertex count %d", k, n)
	}
	if !directed && n%2 != 0 && k%2 != 0 {
		return nil, fmt.Errorf("igraph: undirected k-regular graph requires an even n*k: n=%d, k=%d", n, k)
	}
	return generateGraph("igraph_k_regular_game", options.Seed, func(graph *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_k_regular_game(
			graph,
			C.igraph_int_t(n),
			C.igraph_int_t(k),
			booltoint(directed),
			booltoint(multiple),
		)
	})
}

// RandomTreeGame samples a labelled tree on n vertices uniformly at random.
// Graphs with fewer than two vertices contain no edges. TreeGamePrufer
// supports undirected trees only; TreeGameLERW samples directed trees as
// out-trees rooted at vertex 0.
//
// The returned graph is independently Go-owned and must be closed by the
// caller. Options are read only for the duration of the call. A non-nil
// options.Seed makes the sample reproducible under the package RNG contract.
//
//igraph:bind igraph_tree_game
func RandomTreeGame(n int, directed bool, method TreeGameMethod, options TreeGameOptions) (*Graph, error) {
	if err := validateConstructorSize("vertex count", n); err != nil {
		return nil, err
	}
	cMethod, err := method.cValue()
	if err != nil {
		return nil, err
	}
	if directed && method == TreeGamePrufer {
		return nil, fmt.Errorf("igraph: the Prüfer method does not support directed trees")
	}
	return generateGraph("igraph_tree_game", options.Seed, func(graph *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_tree_game(
			graph,
			C.igraph_int_t(n),
			booltoint(directed),
			cMethod,
		)
	})
}

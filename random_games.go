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

// DegSeqOptions contains options for the degree-sequence random graph generator.
type DegSeqOptions struct {
	// Seed optionally seeds the package random number generator.
	Seed *uint64
}

// DegSeqMethod selects the sampling algorithm for DegreeSequenceGame.
type DegSeqMethod uint8

const (
	// DegSeqConfiguration samples random multigraphs using the configuration model.
	// Self-loops and multi-edges may be produced.
	DegSeqConfiguration DegSeqMethod = iota
	// DegSeqVL samples random simple connected graphs using the Viger-Latapy algorithm.
	// It supports undirected degree sequences only.
	DegSeqVL
	// DegSeqSimpleNoMultiple samples simple graphs using a fast heuristic algorithm.
	DegSeqSimpleNoMultiple
	// DegSeqSimpleNoMultipleUniform samples simple graphs uniformly at random using configuration model rejection sampling.
	DegSeqSimpleNoMultipleUniform
	// DegSeqEdgeSwitchingSimple samples simple graphs using an edge-switching MCMC chain.
	DegSeqEdgeSwitchingSimple

	// DegSeqSimple is a legacy alias for DegSeqConfiguration.
	DegSeqSimple = DegSeqConfiguration
)

func (method DegSeqMethod) cValue() (C.igraph_degseq_t, error) {
	switch method {
	case DegSeqConfiguration:
		return C.IGRAPH_DEGSEQ_CONFIGURATION, nil
	case DegSeqVL:
		return C.IGRAPH_DEGSEQ_VL, nil
	case DegSeqSimpleNoMultiple:
		return C.IGRAPH_DEGSEQ_FAST_HEUR_SIMPLE, nil
	case DegSeqSimpleNoMultipleUniform:
		return C.IGRAPH_DEGSEQ_CONFIGURATION_SIMPLE, nil
	case DegSeqEdgeSwitchingSimple:
		return C.IGRAPH_DEGSEQ_EDGE_SWITCHING_SIMPLE, nil
	default:
		return 0, fmt.Errorf("igraph: invalid degree sequence method: %d", method)
	}
}

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

// DegreeSequenceGame samples a random graph with the given degree sequence(s)
// using the specified method.
//
// For undirected graphs, pass outDeg as the degree sequence and inDeg as nil.
// For directed graphs, pass both outDeg (out-degrees) and inDeg (in-degrees).
//
// Input slices are borrowed for the duration of the call.
// The returned graph is independently Go-owned and must be closed by the
// caller. Options are read only for the duration of the call. A non-nil
// options.Seed makes the sample reproducible under the package RNG contract.
//
//igraph:bind igraph_degree_sequence_game
func DegreeSequenceGame(outDeg []int, inDeg []int, method DegSeqMethod, options DegSeqOptions) (*Graph, error) {
	cMethod, err := method.cValue()
	if err != nil {
		return nil, err
	}
	if err := validateConstructorSize("vertex count", len(outDeg)); err != nil {
		return nil, err
	}

	outSum := 0
	for i, d := range outDeg {
		if d < 0 {
			return nil, fmt.Errorf("igraph: outDeg degree at index %d must be non-negative: %d", i, d)
		}
		outSum += d
	}

	if inDeg == nil {
		if method == DegSeqVL && outSum%2 != 0 {
			return nil, fmt.Errorf("igraph: undirected degree sequence sum must be even: %d", outSum)
		}
		if outSum%2 != 0 && (method == DegSeqSimpleNoMultiple || method == DegSeqSimpleNoMultipleUniform || method == DegSeqEdgeSwitchingSimple) {
			return nil, fmt.Errorf("igraph: undirected degree sequence sum must be even: %d", outSum)
		}
	} else {
		if method == DegSeqVL {
			return nil, fmt.Errorf("igraph: Viger-Latapy method does not support directed graphs")
		}
		if len(outDeg) != len(inDeg) {
			return nil, fmt.Errorf("igraph: outDeg length (%d) and inDeg length (%d) must match", len(outDeg), len(inDeg))
		}
		inSum := 0
		for i, d := range inDeg {
			if d < 0 {
				return nil, fmt.Errorf("igraph: inDeg degree at index %d must be non-negative: %d", i, d)
			}
			inSum += d
		}
		if outSum != inSum {
			return nil, fmt.Errorf("igraph: sum of out-degrees (%d) must equal sum of in-degrees (%d)", outSum, inSum)
		}
	}

	outVec, err := newIntVector(outDeg)
	if err != nil {
		return nil, err
	}
	defer outVec.close()

	var inVecPtr *C.igraph_vector_int_t
	if inDeg != nil {
		inVec, err := newIntVector(inDeg)
		if err != nil {
			return nil, err
		}
		defer inVec.close()
		inVecPtr = &inVec.value
	}

	return generateGraph("igraph_degree_sequence_game", options.Seed, func(graph *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_degree_sequence_game(
			graph,
			&outVec.value,
			inVecPtr,
			cMethod,
		)
	})
}

// IsGraphical checks whether the given degree sequence(s) can be realized by a graph
// under the specified allowed edge types.
//
// For undirected graphs, pass the degree sequence as outDeg and set inDeg to nil.
// For directed graphs, pass both outDeg (out-degrees) and inDeg (in-degrees).
//
// Input slices are borrowed for the duration of the call.
//
//igraph:bind igraph_is_graphical
func IsGraphical(outDeg []int, inDeg []int, allowedEdgeTypes EdgeType) (bool, error) {
	cEdgeType, err := allowedEdgeTypes.cValue()
	if err != nil {
		return false, err
	}
	outVec, err := newIntVector(outDeg)
	if err != nil {
		return false, err
	}
	defer outVec.close()

	var inVecPtr *C.igraph_vector_int_t
	if inDeg != nil {
		if len(outDeg) != len(inDeg) {
			return false, fmt.Errorf("igraph: outDeg length (%d) and inDeg length (%d) must match", len(outDeg), len(inDeg))
		}
		inVec, err := newIntVector(inDeg)
		if err != nil {
			return false, err
		}
		defer inVec.close()
		inVecPtr = &inVec.value
	}

	var res C.igraph_bool_t
	if code := C.go_igraph_is_graphical(&outVec.value, inVecPtr, cEdgeType, &res); code != C.IGRAPH_SUCCESS {
		return false, igraphError("igraph_is_graphical", int(code))
	}
	return res != booltoint(false), nil
}

// IsBigraphical checks whether two degree sequences can be realized by a bipartite graph
// under the specified allowed edge types.
//
// Input slices are borrowed for the duration of the call.
//
//igraph:bind igraph_is_bigraphical
func IsBigraphical(deg1 []int, deg2 []int, allowedEdgeTypes EdgeType) (bool, error) {
	cEdgeType, err := allowedEdgeTypes.cValue()
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
	if code := C.go_igraph_is_bigraphical(&vec1.value, &vec2.value, cEdgeType, &res); code != C.IGRAPH_SUCCESS {
		return false, igraphError("igraph_is_bigraphical", int(code))
	}
	return res != booltoint(false), nil
}

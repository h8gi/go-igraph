package igraph

/*
#include <igraph.h>
#include "algorithm_cgo.h"
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

// allowedEdgeTypes maps the loops flag of the Erdős-Rényi generators onto the
// shared EdgeType constants.
func allowedEdgeTypes(loops bool) EdgeType {
	if loops {
		return EdgeTypeLoops
	}
	return EdgeTypeSimple
}

func edgeTypeFromFlags(loops, multiple bool) EdgeType {
	if loops && multiple {
		return EdgeTypeLoopsAndMulti
	} else if loops {
		return EdgeTypeLoops
	} else if multiple {
		return EdgeTypeMulti
	}
	return EdgeTypeSimple
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
	cEdgeTypes, err := allowedEdgeTypes(loops).cValue()
	if err != nil {
		return nil, err
	}
	return generateGraph("igraph_erdos_renyi_game_gnm", options.Seed, func(graph *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_erdos_renyi_game_gnm(
			graph,
			C.igraph_int_t(n),
			C.igraph_int_t(m),
			booltoint(directed),
			cEdgeTypes,
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
	cEdgeTypes, err := allowedEdgeTypes(loops).cValue()
	if err != nil {
		return nil, err
	}
	return generateGraph("igraph_erdos_renyi_game_gnp", options.Seed, func(graph *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_erdos_renyi_game_gnp(
			graph,
			C.igraph_int_t(n),
			C.igraph_real_t(p),
			booltoint(directed),
			cEdgeTypes,
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

// sumDegrees validates that every degree in the named sequence is
// non-negative and returns the sequence sum, rejecting sums that overflow int.
func sumDegrees(name string, degrees []int) (int, error) {
	sum := 0
	for i, d := range degrees {
		if d < 0 {
			return 0, fmt.Errorf("igraph: %s degree at index %d must be non-negative: %d", name, i, d)
		}
		next := sum + d
		if next < sum {
			return 0, fmt.Errorf("igraph: %s degree sum overflows int", name)
		}
		sum = next
	}
	return sum, nil
}

// newOptionalInDegrees converts a directed in-degree sequence into a C vector,
// or returns nil when inDeg is nil or empty (the undirected form). The length
// check runs before any C allocation.
func newOptionalInDegrees(outDeg, inDeg []int) (*intVector, error) {
	if len(inDeg) == 0 {
		return nil, nil
	}
	if len(outDeg) != len(inDeg) {
		return nil, fmt.Errorf("igraph: outDeg length (%d) and inDeg length (%d) must match", len(outDeg), len(inDeg))
	}
	return newIntVector(inDeg)
}

// DegreeSequenceGame samples a random graph with the given degree sequence(s)
// using the specified method.
//
// A nil or empty inDeg samples an undirected graph with outDeg as its degree
// sequence; otherwise outDeg holds out-degrees and inDeg holds in-degrees of
// a directed graph.
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

	outSum, err := sumDegrees("outDeg", outDeg)
	if err != nil {
		return nil, err
	}

	if len(inDeg) == 0 {
		// Every undirected sampling method rejects an odd degree sum.
		if outSum%2 != 0 {
			return nil, fmt.Errorf("igraph: undirected degree sequence sum must be even: %d", outSum)
		}
	} else {
		if method == DegSeqVL {
			return nil, fmt.Errorf("igraph: Viger-Latapy method does not support directed graphs")
		}
		inSum, err := sumDegrees("inDeg", inDeg)
		if err != nil {
			return nil, err
		}
		if outSum != inSum {
			return nil, fmt.Errorf("igraph: sum of out-degrees (%d) must equal sum of in-degrees (%d)", outSum, inSum)
		}
	}

	inVec, err := newOptionalInDegrees(outDeg, inDeg)
	if err != nil {
		return nil, err
	}
	if inVec != nil {
		defer inVec.close()
	}

	outVec, err := newIntVector(outDeg)
	if err != nil {
		return nil, err
	}
	defer outVec.close()

	var inVecPtr *C.igraph_vector_int_t
	if inVec != nil {
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

// BarabasiAlgorithm selects the generator algorithm used by BarabasiGame.
type BarabasiAlgorithm uint8

const (
	// BarabasiBag uses the original bag-based preferential attachment algorithm.
	BarabasiBag BarabasiAlgorithm = iota
	// BarabasiPSumTree uses the partial sum tree algorithm.
	BarabasiPSumTree
	// BarabasiPSumTreeMultiple allows multiple edges and uses the partial sum tree algorithm.
	BarabasiPSumTreeMultiple
)

func (algo BarabasiAlgorithm) cValue() (C.igraph_barabasi_algorithm_t, error) {
	switch algo {
	case BarabasiBag:
		return C.IGRAPH_BARABASI_BAG, nil
	case BarabasiPSumTree:
		return C.IGRAPH_BARABASI_PSUMTREE, nil
	case BarabasiPSumTreeMultiple:
		return C.IGRAPH_BARABASI_PSUMTREE_MULTIPLE, nil
	default:
		return 0, fmt.Errorf("igraph: invalid Barabasi algorithm: %d", algo)
	}
}

// BarabasiOptions contains options for the Barabási-Albert preferential attachment graph model.
type BarabasiOptions struct {
	// Seed optionally seeds the package random number generator.
	Seed *uint64
	// OutSeq optionally specifies out-degrees for newly added vertices.
	OutSeq []int
	// OutPref specifies whether out-degree preference is enabled.
	OutPref bool
	// Algorithm selects the generator algorithm.
	Algorithm BarabasiAlgorithm
	// StartFrom optionally specifies an initial graph to start from.
	StartFrom *Graph
}

// WattsStrogatzOptions contains options for the Watts-Strogatz small-world graph model.
type WattsStrogatzOptions struct {
	// Seed optionally seeds the package random number generator.
	Seed *uint64
}

// SBMOptions contains options for the Stochastic Block Model graph generator.
type SBMOptions struct {
	// Seed optionally seeds the package random number generator.
	Seed *uint64
}

// BarabasiGame generates a random graph according to the Barabási-Albert
// preferential attachment model.
//
// Input options slices and optional start graph are borrowed for the duration
// of the call. The returned graph is independently Go-owned and must be closed
// by the caller. Options are read only for the duration of the call. A non-nil
// options.Seed makes the sample reproducible under the package RNG contract.
//
//igraph:bind igraph_barabasi_game
func BarabasiGame(n int, m int, power float64, zeroAppeal float64, directed bool, options BarabasiOptions) (*Graph, error) {
	if err := validateConstructorSize("vertex count", n); err != nil {
		return nil, err
	}
	if err := validateConstructorSize("out-degree m", m); err != nil {
		return nil, err
	}
	if math.IsNaN(power) {
		return nil, fmt.Errorf("igraph: power must not be NaN")
	}
	if zeroAppeal < 0 || math.IsNaN(zeroAppeal) {
		return nil, fmt.Errorf("igraph: zero-appeal must be >= 0, got %g", zeroAppeal)
	}
	cAlgo, err := options.Algorithm.cValue()
	if err != nil {
		return nil, err
	}

	var outSeqPtr *C.igraph_vector_int_t
	if options.OutSeq != nil {
		if len(options.OutSeq) != n {
			return nil, fmt.Errorf("igraph: outseq length (%d) must match vertex count n (%d)", len(options.OutSeq), n)
		}
		if len(options.OutSeq) > 0 {
			for i, deg := range options.OutSeq {
				if deg < 0 {
					return nil, fmt.Errorf("igraph: outseq element at index %d must be >= 0, got %d", i, deg)
				}
			}
			outSeqVec, err := newIntVector(options.OutSeq)
			if err != nil {
				return nil, err
			}
			defer outSeqVec.close()
			outSeqPtr = &outSeqVec.value
		}
	}

	var startFromPtr *C.igraph_t
	if options.StartFrom != nil {
		options.StartFrom.mu.RLock()
		defer options.StartFrom.mu.RUnlock()
		if err := options.StartFrom.checkClosed(); err != nil {
			return nil, err
		}
		startFromPtr = &options.StartFrom.graph
	}

	return generateGraph("igraph_barabasi_game", options.Seed, func(graph *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_barabasi_game(
			graph,
			C.igraph_int_t(n),
			C.igraph_real_t(power),
			C.igraph_int_t(m),
			outSeqPtr,
			booltoint(options.OutPref),
			C.igraph_real_t(zeroAppeal),
			booltoint(directed),
			cAlgo,
			startFromPtr,
		)
	})
}

// WattsStrogatzGame generates a random small-world graph according to the
// Watts-Strogatz model.
//
// The returned graph is independently Go-owned and must be closed by the
// caller. Options are read only for the duration of the call. A non-nil
// options.Seed makes the sample reproducible under the package RNG contract.
//
//igraph:bind igraph_watts_strogatz_game
func WattsStrogatzGame(dim int, size int, nei int, p float64, loops bool, multiple bool, options WattsStrogatzOptions) (*Graph, error) {
	if dim < 1 {
		return nil, fmt.Errorf("igraph: dimension must be >= 1, got %d", dim)
	}
	if err := validateConstructorSize("dimension", dim); err != nil {
		return nil, err
	}
	if size < 1 {
		return nil, fmt.Errorf("igraph: lattice size must be >= 1, got %d", size)
	}
	if err := validateConstructorSize("lattice size", size); err != nil {
		return nil, err
	}
	if nei < 0 {
		return nil, fmt.Errorf("igraph: neighborhood distance must be >= 0, got %d", nei)
	}
	if err := validateConstructorSize("neighborhood distance", nei); err != nil {
		return nil, err
	}
	if p < 0 || p > 1 || math.IsNaN(p) {
		return nil, fmt.Errorf("igraph: probability must be between 0 and 1, got %g", p)
	}
	cEdgeTypes, err := edgeTypeFromFlags(loops, multiple).cValue()
	if err != nil {
		return nil, err
	}

	return generateGraph("igraph_watts_strogatz_game", options.Seed, func(graph *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_watts_strogatz_game(
			graph,
			C.igraph_int_t(dim),
			C.igraph_int_t(size),
			C.igraph_int_t(nei),
			C.igraph_real_t(p),
			cEdgeTypes,
		)
	})
}

// SBMGame generates a random graph according to the Stochastic Block Model.
//
// The returned graph is independently Go-owned and must be closed by the
// caller. Options are read only for the duration of the call. A non-nil
// options.Seed makes the sample reproducible under the package RNG contract.
//
//igraph:bind igraph_sbm_game
func SBMGame(n int, prefMatrix Matrix, blockSizes []int, directed bool, loops bool, options SBMOptions) (*Graph, error) {
	if err := validateConstructorSize("vertex count", n); err != nil {
		return nil, err
	}
	sum := 0
	for i, bs := range blockSizes {
		if bs < 0 {
			return nil, fmt.Errorf("igraph: block size at index %d must be >= 0, got %d", i, bs)
		}
		next := sum + bs
		if next < sum {
			return nil, fmt.Errorf("igraph: block sizes sum overflows int")
		}
		sum = next
	}
	if sum != n {
		return nil, fmt.Errorf("igraph: sum of block sizes (%d) does not match vertex count n (%d)", sum, n)
	}
	rows, cols := prefMatrix.Dims()
	k := len(blockSizes)
	if rows != k || cols != k {
		return nil, fmt.Errorf("igraph: preference matrix dimensions (%dx%d) do not match block sizes count (%d)", rows, cols, k)
	}
	for r := 0; r < k; r++ {
		for c := 0; c < k; c++ {
			val, err := prefMatrix.At(r, c)
			if err != nil {
				return nil, err
			}
			if val < 0.0 || val > 1.0 || math.IsNaN(val) {
				return nil, fmt.Errorf("igraph: preference matrix element at [%d,%d] out of bounds: %g", r, c, val)
			}
		}
	}

	cPrefMat, err := newCMatrix(prefMatrix)
	if err != nil {
		return nil, err
	}
	defer cPrefMat.close()

	cBlockSizes, err := newIntVector(blockSizes)
	if err != nil {
		return nil, err
	}
	defer cBlockSizes.close()

	cEdgeTypes, err := allowedEdgeTypes(loops).cValue()
	if err != nil {
		return nil, err
	}

	return generateGraph("igraph_sbm_game", options.Seed, func(graph *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_sbm_game(
			graph,
			&cPrefMat.value,
			&cBlockSizes.value,
			booltoint(directed),
			cEdgeTypes,
		)
	})
}

func validateOptionalEdgeWeights(g *Graph, weights []float64) (*realVector, error) {
	if weights == nil {
		return nil, nil
	}
	numEdges := int(C.igraph_ecount(&g.graph))
	if len(weights) != numEdges {
		return nil, fmt.Errorf("igraph: weights slice length (%d) must match number of edges (%d)", len(weights), numEdges)
	}
	for i, w := range weights {
		if math.IsNaN(w) || w < 0 || math.IsInf(w, 0) {
			return nil, fmt.Errorf("igraph: weight value at index %d must be non-negative finite: %v", i, w)
		}
	}
	return newRealVector(weights)
}

// RewireMode specifies the allowed edge types during graph rewiring.
type RewireMode uint8

const (
	// RewireSimple allows simple edges only (no self-loops, no multiple edges).
	RewireSimple RewireMode = iota
	// RewireLoops allows self-loops but no multiple edges.
	RewireLoops
	// RewireMulti allows multiple edges between distinct vertices but no self-loops.
	RewireMulti
	// RewireLoopsAndMulti allows both self-loops and multiple edges.
	RewireLoopsAndMulti
)

func (mode RewireMode) cValue() (C.igraph_edge_type_sw_t, error) {
	switch mode {
	case RewireSimple:
		return C.IGRAPH_SIMPLE_SW, nil
	case RewireLoops:
		return C.IGRAPH_LOOPS_SW, nil
	case RewireMulti:
		return C.IGRAPH_MULTI_SW, nil
	case RewireLoopsAndMulti:
		return C.IGRAPH_LOOPS_SW | C.IGRAPH_MULTI_SW, nil
	default:
		return 0, fmt.Errorf("igraph: invalid rewire mode: %d", mode)
	}
}

// RewireOptions contains options for graph edge rewiring operations.
type RewireOptions struct {
	// Seed optionally seeds the package random number generator.
	Seed *uint64
}

// RandomWalkStuckMode specifies the behavior when a random walk reaches a dead end.
type RandomWalkStuckMode uint8

const (
	// RandomWalkStuckError returns an error when a random walk is stuck.
	RandomWalkStuckError RandomWalkStuckMode = iota
	// RandomWalkStuckReturn returns the partial path recorded so far when a random walk is stuck.
	RandomWalkStuckReturn
)

func (mode RandomWalkStuckMode) cValue() (C.igraph_random_walk_stuck_t, error) {
	switch mode {
	case RandomWalkStuckError:
		return C.IGRAPH_RANDOM_WALK_STUCK_ERROR, nil
	case RandomWalkStuckReturn:
		return C.IGRAPH_RANDOM_WALK_STUCK_RETURN, nil
	default:
		return 0, fmt.Errorf("igraph: invalid random walk stuck mode: %d", mode)
	}
}

// RandomWalkOptions contains options for random walk sampling.
type RandomWalkOptions struct {
	// Seed optionally seeds the package random number generator.
	Seed *uint64
	// StuckMode controls behavior when a walk reaches a dead end without outgoing edges.
	StuckMode RandomWalkStuckMode
}

// SpanningTreeOptions contains options for random spanning tree sampling.
type SpanningTreeOptions struct {
	// Seed optionally seeds the package random number generator.
	Seed *uint64
	// Root optionally specifies the root vertex ID. If Root is nil or *Root < 0,
	// C/igraph chooses an arbitrary root vertex.
	Root *int
}

// Rewire randomizes the edges of the receiver graph in-place by performing n trial swaps,
// preserving the degree sequence.
//
// The receiver graph is mutated atomically on success. If parameter validation or
// execution fails, the receiver graph remains unchanged. Options are read only for
// the duration of the call. A non-nil options.Seed makes the operation reproducible
// under the package RNG contract.
//
//igraph:bind igraph_rewire
func (g *Graph) Rewire(n int, mode RewireMode, options RewireOptions) error {
	if g == nil {
		return ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return ErrClosed
	}

	if err := validateConstructorSize("rewire trial count", n); err != nil {
		return err
	}
	cMode, err := mode.cValue()
	if err != nil {
		return err
	}

	var replacement C.igraph_t
	if code := C.go_igraph_copy(&replacement, &g.graph); code != C.IGRAPH_SUCCESS {
		return igraphError("copy graph for rewire", int(code))
	}
	committed := false
	defer func() {
		if !committed {
			C.igraph_destroy(&replacement)
		}
	}()

	errRNG := withRNG(options.Seed, func() error {
		if code := C.go_igraph_rewire(&replacement, C.igraph_int_t(n), cMode); code != C.IGRAPH_SUCCESS {
			return igraphError("igraph_rewire", int(code))
		}
		return nil
	})
	if errRNG != nil {
		return errRNG
	}

	C.igraph_destroy(&g.graph)
	g.graph = replacement
	committed = true
	return nil
}

// RewireEdges returns a new graph created by rewiring each edge of the receiver
// with probability prob.
//
// The returned graph is independently Go-owned and must be closed by the caller.
// The receiver graph is not modified. Options are read only for the duration of
// the call. A non-nil options.Seed makes the operation reproducible under the package
// RNG contract.
//
//igraph:bind igraph_rewire_edges
func (g *Graph) RewireEdges(prob float64, loops bool, multiple bool, options RewireOptions) (*Graph, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return nil, ErrClosed
	}

	if prob < 0 || prob > 1 || math.IsNaN(prob) {
		return nil, fmt.Errorf("igraph: probability must be between 0 and 1, got %g", prob)
	}
	cEdgeTypes, err := edgeTypeFromFlags(loops, multiple).cValue()
	if err != nil {
		return nil, err
	}

	var clone C.igraph_t
	if code := C.go_igraph_copy(&clone, &g.graph); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("copy graph for rewire edges", int(code))
	}
	committed := false
	defer func() {
		if !committed {
			C.igraph_destroy(&clone)
		}
	}()

	errRNG := withRNG(options.Seed, func() error {
		if code := C.go_igraph_rewire_edges(&clone, C.igraph_real_t(prob), cEdgeTypes); code != C.IGRAPH_SUCCESS {
			return igraphError("igraph_rewire_edges", int(code))
		}
		return nil
	})
	if errRNG != nil {
		return nil, errRNG
	}

	committed = true
	return adoptInitializedGraph(&clone), nil
}

// RandomWalk performs a random walk of up to steps length starting at vertex start.
//
// Input weights slice is borrowed for the duration of the call.
// Returns two slices: the sequence of visited vertex IDs and traversed edge IDs.
// A non-nil options.Seed makes the walk reproducible under the package RNG contract.
//
//igraph:bind igraph_random_walk
func (g *Graph) RandomWalk(start int, steps int, mode DirectionMode, weights []float64, options RandomWalkOptions) ([]int, []int, error) {
	if g == nil {
		return nil, nil, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return nil, nil, ErrClosed
	}

	numVertices := int(C.igraph_vcount(&g.graph))
	if start < 0 || start >= numVertices {
		return nil, nil, fmt.Errorf("igraph: start vertex ID out of range [0, %d): %d", numVertices, start)
	}
	if err := validateConstructorSize("step count", steps); err != nil {
		return nil, nil, err
	}
	cMode, err := mode.cValue()
	if err != nil {
		return nil, nil, err
	}
	cStuck, err := options.StuckMode.cValue()
	if err != nil {
		return nil, nil, err
	}

	cWeights, err := validateOptionalEdgeWeights(g, weights)
	if err != nil {
		return nil, nil, err
	}
	if cWeights != nil {
		defer cWeights.close()
	}

	cVertices, err := newIntVector(nil)
	if err != nil {
		return nil, nil, err
	}
	defer cVertices.close()

	cEdges, err := newIntVector(nil)
	if err != nil {
		return nil, nil, err
	}
	defer cEdges.close()

	errRNG := withRNG(options.Seed, func() error {
		if code := C.go_igraph_random_walk(
			&g.graph,
			edgeWeightPointer(cWeights),
			&cVertices.value,
			&cEdges.value,
			C.igraph_int_t(start),
			cMode,
			C.igraph_int_t(steps),
			cStuck,
		); code != C.IGRAPH_SUCCESS {
			return igraphError("igraph_random_walk", int(code))
		}
		return nil
	})
	if errRNG != nil {
		return nil, nil, errRNG
	}

	vSlice, err := cVertices.slice()
	if err != nil {
		return nil, nil, err
	}
	eSlice, err := cEdges.slice()
	if err != nil {
		return nil, nil, err
	}
	return vSlice, eSlice, nil
}

// RandomSpanningTree samples a random spanning tree from the graph.
//
// Input weights slice is borrowed for the duration of the call.
// Returns a slice of edge IDs belonging to the sampled spanning tree.
// A non-nil options.Seed makes the sample reproducible under the package RNG contract.
//
//igraph:bind igraph_random_spanning_tree
func (g *Graph) RandomSpanningTree(weights []float64, options SpanningTreeOptions) ([]int, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return nil, ErrClosed
	}

	cWeights, err := validateOptionalEdgeWeights(g, weights)
	if err != nil {
		return nil, err
	}
	if cWeights != nil {
		defer cWeights.close()
	}

	vid := -1
	if options.Root != nil {
		if *options.Root < 0 || *options.Root >= int(C.igraph_vcount(&g.graph)) {
			return nil, fmt.Errorf("igraph: root vertex ID out of range [0, %d): %d", int(C.igraph_vcount(&g.graph)), *options.Root)
		}
		vid = *options.Root
	}

	cTreeEdges, err := newIntVector(nil)
	if err != nil {
		return nil, err
	}
	defer cTreeEdges.close()

	errRNG := withRNG(options.Seed, func() error {
		if code := C.go_igraph_random_spanning_tree(
			&g.graph,
			&cTreeEdges.value,
			C.igraph_int_t(vid),
		); code != C.IGRAPH_SUCCESS {
			return igraphError("igraph_random_spanning_tree", int(code))
		}
		return nil
	})
	if errRNG != nil {
		return nil, errRNG
	}

	return cTreeEdges.slice()
}

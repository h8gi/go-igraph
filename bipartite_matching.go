package igraph

// #cgo pkg-config: igraph
// #include <float.h>
// #include <igraph.h>
import "C"

import (
	"fmt"
	"math"
	"sort"
)

// MatchedPair identifies two matched vertices. FalseVertex belongs to the
// false mode and TrueVertex belongs to the true mode. Matching results are
// sorted by FalseVertex and contain no unmatched sentinel values.
type MatchedPair struct {
	FalseVertex int
	TrueVertex  int
}

// BipartiteMatchingOptions controls maximum bipartite matching. Nil Weights
// requests maximum-cardinality matching. Non-nil Weights must contain one
// finite value per edge and request maximum total weight. The pinned upstream
// weighted algorithm is numerically stable only for integer-valued weights.
// Epsilon is ignored for unweighted matching. Nil selects a machine-epsilon
// derived default; a supplied value must be finite and positive.
type BipartiteMatchingOptions struct {
	Weights []float64
	Epsilon *float64
}

// BipartiteMatchingResult is a Go-owned maximum matching. Size equals the
// number of pairs. Weight is Size for unweighted matching and the selected
// edge-weight sum for weighted matching. Pairs is non-nil and remains valid
// after graph closure.
type BipartiteMatchingResult struct {
	Size   int
	Weight float64
	Pairs  []MatchedPair
}

// IsBipartiteMatching reports whether pairs form a valid matching under
// partition. Edge directions are ignored. Inputs are borrowed only for the
// synchronous call. Malformed pairs return a Go error; a well-formed pair set
// that lacks a corresponding graph edge returns false.
//
//igraph:bind igraph_is_matching
func (g *Graph) IsBipartiteMatching(partition BipartitePartition, pairs []MatchedPair) (bool, error) {
	return g.validateBipartiteMatching(partition, pairs, false, nil)
}

// IsMaximalBipartiteMatching reports whether pairs form a valid matching that
// cannot be extended by another edge. Maximal does not mean maximum. Edge
// directions are ignored; inputs are borrowed only for the synchronous call.
//
//igraph:bind igraph_is_maximal_matching
func (g *Graph) IsMaximalBipartiteMatching(partition BipartitePartition, pairs []MatchedPair) (bool, error) {
	return g.validateBipartiteMatching(partition, pairs, true, nil)
}

// MaximumBipartiteMatching calculates a maximum cardinality or maximum-weight
// matching. Edge directions are ignored. Partition, weights, epsilon, and the
// graph are borrowed only for the synchronous call. The result is entirely
// Go-owned and contains no upstream unmatched sentinels.
//
//igraph:bind igraph_maximum_bipartite_matching
func (g *Graph) MaximumBipartiteMatching(partition BipartitePartition, options BipartiteMatchingOptions) (BipartiteMatchingResult, error) {
	return g.maximumBipartiteMatching(partition, options, nil)
}

type matchingAdapters struct {
	newBool    func([]bool) (*boolVector, error)
	newInt     func([]int) (*intVector, error)
	newReal    func([]float64) (*realVector, error)
	convertInt func(*intVector) ([]int, error)
	validate   func(*Graph, *boolVector, *intVector, bool) (bool, int)
	maximum    func(*Graph, *boolVector, *realVector, float64, *intVector) (int64, float64, int)
}

func defaultMatchingAdapters() matchingAdapters {
	return matchingAdapters{
		newBool: newBoolVector, newInt: newIntVector, newReal: newRealVector,
		convertInt: (*intVector).slice,
		validate: func(g *Graph, types *boolVector, matching *intVector, maximal bool) (bool, int) {
			var result C.igraph_bool_t
			var code C.igraph_error_t
			if maximal {
				code = C.igraph_is_maximal_matching(&g.graph, &types.value, &matching.value, &result)
			} else {
				code = C.igraph_is_matching(&g.graph, &types.value, &matching.value, &result)
			}
			return result != booltoint(false), int(code)
		},
		maximum: func(g *Graph, types *boolVector, weights *realVector, epsilon float64, matching *intVector) (int64, float64, int) {
			var size C.igraph_int_t
			var weight C.igraph_real_t
			code := C.igraph_maximum_bipartite_matching(&g.graph, &types.value, &size, &weight, &matching.value, realVectorPointer(weights), C.igraph_real_t(epsilon))
			return int64(size), float64(weight), int(code)
		},
	}
}

func resolvedMatchingAdapters(adapters *matchingAdapters) matchingAdapters {
	if adapters == nil {
		return defaultMatchingAdapters()
	}
	return *adapters
}

func matchingVector(partition BipartitePartition, pairs []MatchedPair) ([]int, error) {
	mates := make([]int, len(partition))
	for i := range mates {
		mates[i] = RemovedID
	}
	for index, pair := range pairs {
		if pair.FalseVertex < 0 || pair.FalseVertex >= len(partition) {
			return nil, fmt.Errorf("igraph: matching pair %d false vertex %d out of range [0, %d)", index, pair.FalseVertex, len(partition))
		}
		if pair.TrueVertex < 0 || pair.TrueVertex >= len(partition) {
			return nil, fmt.Errorf("igraph: matching pair %d true vertex %d out of range [0, %d)", index, pair.TrueVertex, len(partition))
		}
		if partition[pair.FalseVertex] || !partition[pair.TrueVertex] {
			return nil, fmt.Errorf("igraph: matching pair %d does not follow false-to-true mode order", index)
		}
		if mates[pair.FalseVertex] != RemovedID || mates[pair.TrueVertex] != RemovedID {
			return nil, fmt.Errorf("igraph: matching pair %d reuses a matched vertex", index)
		}
		mates[pair.FalseVertex], mates[pair.TrueVertex] = pair.TrueVertex, pair.FalseVertex
	}
	return mates, nil
}

func matchingPairs(partition BipartitePartition, mates []int) ([]MatchedPair, error) {
	if len(mates) != len(partition) {
		return nil, fmt.Errorf("igraph: matching result length %d does not match vertex count %d", len(mates), len(partition))
	}
	pairs := make([]MatchedPair, 0)
	for falseVertex, mate := range mates {
		if partition[falseVertex] || mate == RemovedID {
			continue
		}
		if mate < 0 || mate >= len(partition) || !partition[mate] || mates[mate] != falseVertex {
			return nil, fmt.Errorf("igraph: malformed upstream matching at vertex %d", falseVertex)
		}
		pairs = append(pairs, MatchedPair{FalseVertex: falseVertex, TrueVertex: mate})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].FalseVertex < pairs[j].FalseVertex })
	return pairs, nil
}

func (g *Graph) validateBipartiteMatching(partition BipartitePartition, pairs []MatchedPair, maximal bool, adapters *matchingAdapters) (bool, error) {
	if g == nil {
		return false, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return false, ErrClosed
	}
	if err := validateProjectionPartitionLocked(g, partition); err != nil {
		return false, err
	}
	mates, err := matchingVector(partition, pairs)
	if err != nil {
		return false, err
	}
	resolved := resolvedMatchingAdapters(adapters)
	types, err := resolved.newBool([]bool(partition))
	if err != nil {
		return false, err
	}
	defer types.close()
	matching, err := resolved.newInt(mates)
	if err != nil {
		return false, err
	}
	defer matching.close()
	result, code := resolved.validate(g, types, matching, maximal)
	if code != int(C.IGRAPH_SUCCESS) {
		return false, igraphError("validate bipartite matching", code)
	}
	return result, nil
}

func validateMatchingOptions(options BipartiteMatchingOptions, edgeCount int) (float64, error) {
	epsilon := float64(C.DBL_EPSILON) * 100
	if options.Epsilon != nil {
		if math.IsNaN(*options.Epsilon) || math.IsInf(*options.Epsilon, 0) || *options.Epsilon <= 0 {
			return 0, fmt.Errorf("igraph: matching epsilon must be finite and positive: %v", *options.Epsilon)
		}
		epsilon = *options.Epsilon
	}
	if options.Weights != nil {
		if len(options.Weights) != edgeCount {
			return 0, fmt.Errorf("igraph: matching weight count %d does not match edge count %d", len(options.Weights), edgeCount)
		}
		for index, value := range options.Weights {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return 0, fmt.Errorf("igraph: matching weight at index %d must be finite: %v", index, value)
			}
		}
	}
	return epsilon, nil
}

func (g *Graph) maximumBipartiteMatching(partition BipartitePartition, options BipartiteMatchingOptions, adapters *matchingAdapters) (BipartiteMatchingResult, error) {
	if g == nil {
		return BipartiteMatchingResult{}, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return BipartiteMatchingResult{}, ErrClosed
	}
	if err := validateProjectionPartitionLocked(g, partition); err != nil {
		return BipartiteMatchingResult{}, err
	}
	epsilon, err := validateMatchingOptions(options, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return BipartiteMatchingResult{}, err
	}
	resolved := resolvedMatchingAdapters(adapters)
	types, err := resolved.newBool([]bool(partition))
	if err != nil {
		return BipartiteMatchingResult{}, err
	}
	defer types.close()
	matching, err := resolved.newInt(nil)
	if err != nil {
		return BipartiteMatchingResult{}, err
	}
	defer matching.close()
	var weights *realVector
	if options.Weights != nil {
		weights, err = resolved.newReal(options.Weights)
		if err != nil {
			return BipartiteMatchingResult{}, err
		}
		defer weights.close()
	}
	size64, weight, code := resolved.maximum(g, types, weights, epsilon, matching)
	if code != int(C.IGRAPH_SUCCESS) {
		return BipartiteMatchingResult{}, igraphError("calculate maximum bipartite matching", code)
	}
	if size64 < 0 || uint64(size64) > uint64(^uint(0)>>1) {
		return BipartiteMatchingResult{}, fmt.Errorf("igraph: matching size is out of Go int range")
	}
	mates, err := resolved.convertInt(matching)
	if err != nil {
		return BipartiteMatchingResult{}, err
	}
	pairs, err := matchingPairs(partition, mates)
	if err != nil {
		return BipartiteMatchingResult{}, err
	}
	size := int(size64)
	if len(pairs) != size {
		return BipartiteMatchingResult{}, fmt.Errorf("igraph: matching size %d does not match pair count %d", size, len(pairs))
	}
	return BipartiteMatchingResult{Size: size, Weight: weight, Pairs: pairs}, nil
}

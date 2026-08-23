package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "degree_realization_cgo.h"
import "C"

import "fmt"

// DegreeSequenceRealizationMethod controls the deterministic order in which
// vertices with remaining degree are selected.
type DegreeSequenceRealizationMethod uint8

const (
	// RealizeSmallest selects the smallest remaining degree first and, when a
	// connected realization exists, produces one for supported undirected and
	// bipartite inputs.
	RealizeSmallest DegreeSequenceRealizationMethod = iota
	// RealizeLargest selects the largest remaining degree first.
	RealizeLargest
	// RealizeIndex selects vertices in vertex-ID order.
	RealizeIndex
)

func (method DegreeSequenceRealizationMethod) cValue() (C.igraph_realize_degseq_t, error) {
	switch method {
	case RealizeSmallest:
		return C.IGRAPH_REALIZE_DEGSEQ_SMALLEST, nil
	case RealizeLargest:
		return C.IGRAPH_REALIZE_DEGSEQ_LARGEST, nil
	case RealizeIndex:
		return C.IGRAPH_REALIZE_DEGSEQ_INDEX, nil
	default:
		return 0, fmt.Errorf("igraph: invalid degree-sequence realization method: %d", method)
	}
}

// DegreeSequenceRealizationOptions controls deterministic realization. Its
// zero value requests a simple graph using RealizeSmallest.
type DegreeSequenceRealizationOptions struct {
	EdgeTypes EdgeType
	Method    DegreeSequenceRealizationMethod
}

// RealizeDegreeSequence constructs an undirected graph when inDegrees is nil
// or empty. Otherwise, outDegrees and inDegrees specify aligned directed out-
// and in-degree sequences. Directed realization supports simple graphs only in
// pinned C/igraph 1.0.1. Inputs are borrowed only for the call. The returned
// graph is independently owned and must be closed.
//
//igraph:bind igraph_realize_degree_sequence
func RealizeDegreeSequence(outDegrees, inDegrees []int, options DegreeSequenceRealizationOptions) (*Graph, error) {
	return realizeDegreeSequence(outDegrees, inDegrees, options, nil)
}

// RealizeBipartiteDegreeSequence constructs an undirected bipartite graph.
// Vertices for falseModeDegrees come first, followed by vertices for
// trueModeDegrees. Self-loop flags in EdgeTypes are ignored because bipartite
// graphs cannot contain loops; the multi-edge flag is honored. Inputs are
// borrowed only for the call. The returned graph and partition are
// independently Go-owned.
//
//igraph:bind igraph_realize_bipartite_degree_sequence
func RealizeBipartiteDegreeSequence(falseModeDegrees, trueModeDegrees []int, options DegreeSequenceRealizationOptions) (BipartiteGraphResult, error) {
	return realizeBipartiteDegreeSequence(falseModeDegrees, trueModeDegrees, options, nil)
}

type degreeRealizationCallResult struct {
	graph C.igraph_t
	code  int
}

type degreeRealizationAdapters struct {
	newInt    func([]int) (*intVector, error)
	realize   func(*intVector, *intVector, EdgeType, DegreeSequenceRealizationMethod) degreeRealizationCallResult
	bipartite func(*intVector, *intVector, EdgeType, DegreeSequenceRealizationMethod) degreeRealizationCallResult
}

func defaultDegreeRealizationAdapters() degreeRealizationAdapters {
	return degreeRealizationAdapters{
		newInt: newIntVector,
		realize: func(outDegrees, inDegrees *intVector, edgeTypes EdgeType, method DegreeSequenceRealizationMethod) degreeRealizationCallResult {
			cEdgeTypes, _ := edgeTypes.cValue()
			cMethod, _ := method.cValue()
			var inPointer *C.igraph_vector_int_t
			if inDegrees != nil {
				inPointer = &inDegrees.value
			}
			var graph C.igraph_t
			code := C.go_igraph_realize_degree_sequence(
				&graph, &outDegrees.value, inPointer, cEdgeTypes, cMethod,
			)
			return degreeRealizationCallResult{graph: graph, code: int(code)}
		},
		bipartite: func(falseDegrees, trueDegrees *intVector, edgeTypes EdgeType, method DegreeSequenceRealizationMethod) degreeRealizationCallResult {
			cEdgeTypes, _ := edgeTypes.cValue()
			cMethod, _ := method.cValue()
			var graph C.igraph_t
			code := C.go_igraph_realize_bipartite_degree_sequence(
				&graph, &falseDegrees.value, &trueDegrees.value, cEdgeTypes, cMethod,
			)
			return degreeRealizationCallResult{graph: graph, code: int(code)}
		},
	}
}

func resolvedDegreeRealizationAdapters(adapters *degreeRealizationAdapters) degreeRealizationAdapters {
	if adapters == nil {
		return defaultDegreeRealizationAdapters()
	}
	return *adapters
}

func validateDegreeRealizationOptions(options DegreeSequenceRealizationOptions, directed, bipartite bool) (EdgeType, error) {
	if _, err := options.EdgeTypes.cValue(); err != nil {
		return 0, err
	}
	if _, err := options.Method.cValue(); err != nil {
		return 0, err
	}
	if directed && options.EdgeTypes != EdgeTypeSimple {
		return 0, fmt.Errorf("igraph: directed degree-sequence realization supports simple edges only")
	}
	if bipartite {
		if options.EdgeTypes == EdgeTypeLoops {
			return EdgeTypeSimple, nil
		}
		if options.EdgeTypes == EdgeTypeLoopsAndMulti {
			return EdgeTypeMulti, nil
		}
	}
	if !directed && !bipartite && options.EdgeTypes == EdgeTypeLoops {
		return 0, fmt.Errorf("igraph: undirected degree-sequence realization with loops but without multi-edges is not implemented upstream")
	}
	return options.EdgeTypes, nil
}

func validateDegreeSequence(name string, degrees []int) (int, error) {
	return sumDegrees(name, degrees)
}

func realizeDegreeSequence(outDegrees, inDegrees []int, options DegreeSequenceRealizationOptions, adapters *degreeRealizationAdapters) (*Graph, error) {
	directed := len(inDegrees) != 0
	edgeTypes, err := validateDegreeRealizationOptions(options, directed, false)
	if err != nil {
		return nil, err
	}
	outSum, err := validateDegreeSequence("out", outDegrees)
	if err != nil {
		return nil, err
	}
	if directed {
		if len(outDegrees) != len(inDegrees) {
			return nil, fmt.Errorf("igraph: out-degree length (%d) and in-degree length (%d) must match", len(outDegrees), len(inDegrees))
		}
		inSum, err := validateDegreeSequence("in", inDegrees)
		if err != nil {
			return nil, err
		}
		if outSum != inSum {
			return nil, fmt.Errorf("igraph: out-degree sum %d and in-degree sum %d must match", outSum, inSum)
		}
	} else if outSum%2 != 0 {
		return nil, fmt.Errorf("igraph: undirected degree sum must be even: %d", outSum)
	}
	resolved := resolvedDegreeRealizationAdapters(adapters)
	outVector, err := resolved.newInt(outDegrees)
	if err != nil {
		return nil, err
	}
	defer outVector.close()
	var inVector *intVector
	if directed {
		inVector, err = resolved.newInt(inDegrees)
		if err != nil {
			return nil, err
		}
		defer inVector.close()
	}
	call := resolved.realize(outVector, inVector, edgeTypes, options.Method)
	if call.code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("realize degree sequence", call.code)
	}
	return adoptInitializedGraph(&call.graph), nil
}

func realizeBipartiteDegreeSequence(falseModeDegrees, trueModeDegrees []int, options DegreeSequenceRealizationOptions, adapters *degreeRealizationAdapters) (BipartiteGraphResult, error) {
	edgeTypes, err := validateDegreeRealizationOptions(options, false, true)
	if err != nil {
		return BipartiteGraphResult{}, err
	}
	falseSum, err := validateDegreeSequence("false-mode", falseModeDegrees)
	if err != nil {
		return BipartiteGraphResult{}, err
	}
	trueSum, err := validateDegreeSequence("true-mode", trueModeDegrees)
	if err != nil {
		return BipartiteGraphResult{}, err
	}
	if falseSum != trueSum {
		return BipartiteGraphResult{}, fmt.Errorf("igraph: false-mode degree sum %d and true-mode degree sum %d must match", falseSum, trueSum)
	}
	resolved := resolvedDegreeRealizationAdapters(adapters)
	falseVector, err := resolved.newInt(falseModeDegrees)
	if err != nil {
		return BipartiteGraphResult{}, err
	}
	defer falseVector.close()
	trueVector, err := resolved.newInt(trueModeDegrees)
	if err != nil {
		return BipartiteGraphResult{}, err
	}
	defer trueVector.close()
	call := resolved.bipartite(falseVector, trueVector, edgeTypes, options.Method)
	if call.code != int(C.IGRAPH_SUCCESS) {
		return BipartiteGraphResult{}, igraphError("realize bipartite degree sequence", call.code)
	}
	partition := make(BipartitePartition, len(falseModeDegrees)+len(trueModeDegrees))
	for index := len(falseModeDegrees); index < len(partition); index++ {
		partition[index] = true
	}
	return BipartiteGraphResult{Graph: adoptInitializedGraph(&call.graph), Partition: partition}, nil
}

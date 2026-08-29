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

// ChungLuVariant selects the connection-probability formulation used by
// ChungLuGame.
type ChungLuVariant uint8

const (
	// ChungLuOriginal uses min(q, 1), where q is the product of the endpoint
	// weights divided by their common sum. ChungLuGame rejects inputs for which
	// an eligible pair would require clamping q above one.
	ChungLuOriginal ChungLuVariant = iota
	// ChungLuMaximumEntropy uses q/(1+q), the maximum-entropy formulation.
	ChungLuMaximumEntropy
	// ChungLuNorrosReittu uses 1-exp(-q), the simple-graph projection of the
	// Norros-Reittu model.
	ChungLuNorrosReittu
)

func (variant ChungLuVariant) cValue() (C.igraph_chung_lu_t, error) {
	switch variant {
	case ChungLuOriginal:
		return C.IGRAPH_CHUNG_LU_ORIGINAL, nil
	case ChungLuMaximumEntropy:
		return C.IGRAPH_CHUNG_LU_MAXENT, nil
	case ChungLuNorrosReittu:
		return C.IGRAPH_CHUNG_LU_NR, nil
	default:
		return 0, fmt.Errorf("igraph: invalid Chung-Lu variant: %d", variant)
	}
}

// ChungLuOptions controls Chung-Lu graph generation. Its zero value selects
// the original model without self-loops and uses the package RNG's current
// state.
type ChungLuOptions struct {
	// Seed optionally seeds the package random number generator.
	Seed *uint64
	// Loops allows self-loops.
	Loops bool
	// Variant selects the connection-probability formulation.
	Variant ChungLuVariant
}

// StaticFitnessOptions controls fixed-edge-count fitness graph generation.
// Its zero value requests a simple graph and uses the package RNG's current
// state.
type StaticFitnessOptions struct {
	// Seed optionally seeds the package random number generator.
	Seed *uint64
	// EdgeTypes controls whether loops and parallel edges are allowed.
	EdgeTypes EdgeType
}

// StaticPowerLawOptions controls static power-law graph generation. Its zero
// value requests an undirected simple graph without finite-size correction.
type StaticPowerLawOptions struct {
	// Seed optionally seeds the package random number generator.
	Seed *uint64
	// InExponent requests a directed graph and supplies its in-degree power-law
	// exponent. Nil requests an undirected graph. The pointed-to value is
	// borrowed only for the synchronous call.
	InExponent *float64
	// EdgeTypes controls whether loops and parallel edges are allowed.
	EdgeTypes EdgeType
	// FiniteSizeCorrection enables the Cho et al. correction.
	FiniteSizeCorrection bool
}

// ChungLuGame samples a simple graph from a Chung-Lu expected-degree model.
// expectedOutDegrees is vertex-ID-aligned. A nil or empty expectedInDegrees
// requests an undirected graph; otherwise it must have the same length and
// exactly the same finite sum and supplies directed in-degree expectations.
// All weights must be finite and non-negative.
//
// The original formulation rejects inputs that would yield a connection
// probability greater than one instead of silently clamping the probability
// and invalidating the expected-degree interpretation. The other formulations
// accept every finite non-negative weight sequence. This binds an experimental
// API in pinned C/igraph 1.0.1.
//
// Input slices and options are borrowed only for the synchronous call. The
// returned graph is independently Go-owned and must be closed by the caller.
// A non-nil options.Seed makes the sample reproducible under the package RNG
// contract.
//
//igraph:bind igraph_chung_lu_game
func ChungLuGame(expectedOutDegrees, expectedInDegrees []float64, options ChungLuOptions) (*Graph, error) {
	return chungLuGame(expectedOutDegrees, expectedInDegrees, options, nil)
}

// StaticFitnessGame samples a graph with exactly edgeCount edges, with endpoint
// probabilities proportional to vertex-ID-aligned fitness values. A nil or
// empty inFitness requests an undirected graph; otherwise it supplies directed
// in-fitness values and must match outFitness in length. Fitness values must be
// finite and non-negative. At least one eligible positive-fitness endpoint pair
// must exist when edgeCount is positive.
//
// Input slices and options are borrowed only for the synchronous call. The
// returned graph is independently Go-owned and must be closed by the caller.
// A non-nil options.Seed makes the sample reproducible under the package RNG
// contract.
//
//igraph:bind igraph_static_fitness_game
func StaticFitnessGame(edgeCount int, outFitness, inFitness []float64, options StaticFitnessOptions) (*Graph, error) {
	return staticFitnessGame(edgeCount, outFitness, inFitness, options, nil)
}

// StaticPowerLawGame samples a graph with vertexCount vertices and exactly
// edgeCount edges whose expected degrees follow power-law distributions.
// outExponent and a non-nil options.InExponent must be at least two; positive
// infinity is accepted and yields a uniform-fitness Erdős-Rényi-like model.
// Nil InExponent requests an undirected graph.
//
// Options are borrowed only for the synchronous call. The returned graph is
// independently Go-owned and must be closed by the caller. A non-nil
// options.Seed makes the sample reproducible under the package RNG contract.
//
//igraph:bind igraph_static_power_law_game
func StaticPowerLawGame(vertexCount, edgeCount int, outExponent float64, options StaticPowerLawOptions) (*Graph, error) {
	return staticPowerLawGame(vertexCount, edgeCount, outExponent, options, nil)
}

type expectedDegreeGraphCallResult struct {
	graph C.igraph_t
	code  int
}

type expectedDegreeRandomAdapters struct {
	newReal       func([]float64) (*realVector, error)
	closeReal     func(*realVector)
	chungLu       func(*realVector, *realVector, ChungLuOptions) expectedDegreeGraphCallResult
	staticFitness func(int, *realVector, *realVector, StaticFitnessOptions) expectedDegreeGraphCallResult
	staticPower   func(int, int, float64, StaticPowerLawOptions) expectedDegreeGraphCallResult
}

func defaultExpectedDegreeRandomAdapters() expectedDegreeRandomAdapters {
	return expectedDegreeRandomAdapters{
		newReal:   newRealVector,
		closeReal: (*realVector).close,
		chungLu: func(out, in *realVector, options ChungLuOptions) expectedDegreeGraphCallResult {
			variant, _ := options.Variant.cValue()
			var inPointer *C.igraph_vector_t
			if in != nil {
				inPointer = &in.value
			}
			var graph C.igraph_t
			code := C.go_igraph_chung_lu_game(&graph, &out.value, inPointer, booltoint(options.Loops), variant)
			return expectedDegreeGraphCallResult{graph: graph, code: int(code)}
		},
		staticFitness: func(edgeCount int, out, in *realVector, options StaticFitnessOptions) expectedDegreeGraphCallResult {
			edgeTypes, _ := options.EdgeTypes.cValue()
			var inPointer *C.igraph_vector_t
			if in != nil {
				inPointer = &in.value
			}
			var graph C.igraph_t
			code := C.go_igraph_static_fitness_game(&graph, C.igraph_int_t(edgeCount), &out.value, inPointer, edgeTypes)
			return expectedDegreeGraphCallResult{graph: graph, code: int(code)}
		},
		staticPower: func(vertexCount, edgeCount int, outExponent float64, options StaticPowerLawOptions) expectedDegreeGraphCallResult {
			edgeTypes, _ := options.EdgeTypes.cValue()
			inExponent := -1.0
			if options.InExponent != nil {
				inExponent = *options.InExponent
			}
			var graph C.igraph_t
			code := C.go_igraph_static_power_law_game(
				&graph,
				C.igraph_int_t(vertexCount),
				C.igraph_int_t(edgeCount),
				C.igraph_real_t(outExponent),
				C.igraph_real_t(inExponent),
				edgeTypes,
				booltoint(options.FiniteSizeCorrection),
			)
			return expectedDegreeGraphCallResult{graph: graph, code: int(code)}
		},
	}
}

func resolvedExpectedDegreeRandomAdapters(adapters *expectedDegreeRandomAdapters) expectedDegreeRandomAdapters {
	if adapters == nil {
		return defaultExpectedDegreeRandomAdapters()
	}
	return *adapters
}

func validateNonNegativeFiniteValues(name string, values []float64) (sum float64, positive int, err error) {
	for index, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
			return 0, 0, fmt.Errorf("igraph: %s value at index %d must be non-negative and finite: %v", name, index, value)
		}
		sum += value
		if math.IsInf(sum, 0) {
			return 0, 0, fmt.Errorf("igraph: %s sum must be finite", name)
		}
		if value > 0 {
			positive++
		}
	}
	return sum, positive, nil
}

func validateOriginalChungLuProbabilities(out, in []float64, loops bool, sum float64) error {
	if sum == 0 || len(out) == 0 {
		return nil
	}
	maxProduct := 0.0
	if len(in) == 0 {
		largest, second := 0.0, 0.0
		for _, value := range out {
			if value >= largest {
				second, largest = largest, value
			} else if value > second {
				second = value
			}
		}
		if loops {
			maxProduct = largest * largest
		} else {
			maxProduct = largest * second
		}
	} else {
		largest, second, largestIndex := 0.0, 0.0, -1
		for index, value := range in {
			if value >= largest {
				second, largest, largestIndex = largest, value, index
			} else if value > second {
				second = value
			}
		}
		for index, value := range out {
			candidate := largest
			if !loops && index == largestIndex {
				candidate = second
			}
			if product := value * candidate; product > maxProduct {
				maxProduct = product
			}
		}
	}
	if math.IsInf(maxProduct, 0) || maxProduct > sum {
		return fmt.Errorf("igraph: original Chung-Lu weights produce a connection probability greater than one")
	}
	return nil
}

func initializeOptionalRealVector(values []float64, adapters expectedDegreeRandomAdapters) (*realVector, error) {
	if len(values) == 0 {
		return nil, nil
	}
	return adapters.newReal(values)
}

func runExpectedDegreeRandomCall(operation string, seed *uint64, call func() expectedDegreeGraphCallResult) (*Graph, error) {
	var result expectedDegreeGraphCallResult
	err := withRNG(seed, func() error {
		result = call()
		if result.code != int(C.IGRAPH_SUCCESS) {
			return igraphError(operation, result.code)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return adoptInitializedGraph(&result.graph), nil
}

func chungLuGame(expectedOutDegrees, expectedInDegrees []float64, options ChungLuOptions, adapters *expectedDegreeRandomAdapters) (*Graph, error) {
	if _, err := options.Variant.cValue(); err != nil {
		return nil, err
	}
	outSum, _, err := validateNonNegativeFiniteValues("expected out-degree", expectedOutDegrees)
	if err != nil {
		return nil, err
	}
	directed := len(expectedInDegrees) != 0
	if directed && len(expectedInDegrees) != len(expectedOutDegrees) {
		return nil, fmt.Errorf("igraph: expected out-degree length (%d) and in-degree length (%d) must match", len(expectedOutDegrees), len(expectedInDegrees))
	}
	if directed {
		inSum, _, err := validateNonNegativeFiniteValues("expected in-degree", expectedInDegrees)
		if err != nil {
			return nil, err
		}
		if outSum != inSum {
			return nil, fmt.Errorf("igraph: expected out-degree sum %g and in-degree sum %g must match", outSum, inSum)
		}
	}
	if options.Variant == ChungLuOriginal {
		if err := validateOriginalChungLuProbabilities(expectedOutDegrees, expectedInDegrees, options.Loops, outSum); err != nil {
			return nil, err
		}
	}

	resolved := resolvedExpectedDegreeRandomAdapters(adapters)
	outVector, err := resolved.newReal(expectedOutDegrees)
	if err != nil {
		return nil, err
	}
	defer resolved.closeReal(outVector)
	inVector, err := initializeOptionalRealVector(expectedInDegrees, resolved)
	if err != nil {
		return nil, err
	}
	if inVector != nil {
		defer resolved.closeReal(inVector)
	}
	return runExpectedDegreeRandomCall("sample Chung-Lu graph", options.Seed, func() expectedDegreeGraphCallResult {
		return resolved.chungLu(outVector, inVector, options)
	})
}

func edgeTypeAllowsLoops(edgeTypes EdgeType) bool {
	return edgeTypes == EdgeTypeLoops || edgeTypes == EdgeTypeLoopsAndMulti
}

func edgeTypeAllowsMultiple(edgeTypes EdgeType) bool {
	return edgeTypes == EdgeTypeMulti || edgeTypes == EdgeTypeLoopsAndMulti
}

func saturatedProduct(left, right int) int {
	maximum := int(^uint(0) >> 1)
	if left != 0 && right > maximum/left {
		return maximum
	}
	return left * right
}

func staticFitnessPairCapacity(outFitness, inFitness []float64, loops bool) int {
	if len(inFitness) == 0 {
		positive := 0
		for _, value := range outFitness {
			if value > 0 {
				positive++
			}
		}
		product := saturatedProduct(positive, positive-1)
		capacity := product / 2
		if loops {
			maximum := int(^uint(0) >> 1)
			if capacity > maximum-positive {
				return maximum
			}
			capacity += positive
		}
		return capacity
	}
	outPositive, inPositive, commonPositive := 0, 0, 0
	for index := range outFitness {
		outSet := outFitness[index] > 0
		inSet := inFitness[index] > 0
		if outSet {
			outPositive++
		}
		if inSet {
			inPositive++
		}
		if outSet && inSet {
			commonPositive++
		}
	}
	capacity := saturatedProduct(outPositive, inPositive)
	if !loops {
		capacity -= commonPositive
	}
	return capacity
}

func staticFitnessGame(edgeCount int, outFitness, inFitness []float64, options StaticFitnessOptions, adapters *expectedDegreeRandomAdapters) (*Graph, error) {
	if err := validateConstructorSize("static-fitness edge count", edgeCount); err != nil {
		return nil, err
	}
	if _, err := options.EdgeTypes.cValue(); err != nil {
		return nil, err
	}
	if _, _, err := validateNonNegativeFiniteValues("out-fitness", outFitness); err != nil {
		return nil, err
	}
	directed := len(inFitness) != 0
	if directed && len(inFitness) != len(outFitness) {
		return nil, fmt.Errorf("igraph: out-fitness length (%d) and in-fitness length (%d) must match", len(outFitness), len(inFitness))
	}
	if directed {
		if _, _, err := validateNonNegativeFiniteValues("in-fitness", inFitness); err != nil {
			return nil, err
		}
	}
	capacity := staticFitnessPairCapacity(outFitness, inFitness, edgeTypeAllowsLoops(options.EdgeTypes))
	if edgeCount > 0 && capacity == 0 {
		return nil, fmt.Errorf("igraph: positive static-fitness edge count requires an eligible positive-fitness endpoint pair")
	}
	if !edgeTypeAllowsMultiple(options.EdgeTypes) && edgeCount > capacity {
		return nil, fmt.Errorf("igraph: static-fitness edge count %d exceeds maximum %d for the eligible endpoint pairs", edgeCount, capacity)
	}

	resolved := resolvedExpectedDegreeRandomAdapters(adapters)
	outVector, err := resolved.newReal(outFitness)
	if err != nil {
		return nil, err
	}
	defer resolved.closeReal(outVector)
	inVector, err := initializeOptionalRealVector(inFitness, resolved)
	if err != nil {
		return nil, err
	}
	if inVector != nil {
		defer resolved.closeReal(inVector)
	}
	return runExpectedDegreeRandomCall("sample static-fitness graph", options.Seed, func() expectedDegreeGraphCallResult {
		return resolved.staticFitness(edgeCount, outVector, inVector, options)
	})
}

func validatePowerLawExponent(name string, exponent float64) error {
	if math.IsNaN(exponent) || math.IsInf(exponent, -1) || exponent < 2 {
		return fmt.Errorf("igraph: %s exponent must be at least 2 or positive infinity: %v", name, exponent)
	}
	return nil
}

func staticPowerLawGame(vertexCount, edgeCount int, outExponent float64, options StaticPowerLawOptions, adapters *expectedDegreeRandomAdapters) (*Graph, error) {
	if err := validateConstructorSize("static power-law vertex count", vertexCount); err != nil {
		return nil, err
	}
	if err := validateConstructorSize("static power-law edge count", edgeCount); err != nil {
		return nil, err
	}
	if _, err := options.EdgeTypes.cValue(); err != nil {
		return nil, err
	}
	if err := validatePowerLawExponent("out-degree", outExponent); err != nil {
		return nil, err
	}
	if options.InExponent != nil {
		if err := validatePowerLawExponent("in-degree", *options.InExponent); err != nil {
			return nil, err
		}
	}
	if vertexCount == 0 && edgeCount > 0 {
		return nil, fmt.Errorf("igraph: positive static power-law edge count requires at least one vertex")
	}

	resolved := resolvedExpectedDegreeRandomAdapters(adapters)
	return runExpectedDegreeRandomCall("sample static power-law graph", options.Seed, func() expectedDegreeGraphCallResult {
		return resolved.staticPower(vertexCount, edgeCount, outExponent, options)
	})
}

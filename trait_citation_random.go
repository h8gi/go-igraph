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

// TraitGrowthOptions controls sampled vertex types and graph direction. Nil or
// empty TypeWeights selects a uniform distribution. Weights are borrowed only
// for the call and need not be normalized.
type TraitGrowthOptions struct {
	Seed        *uint64
	Directed    bool
	TypeWeights []float64
}

// CitationOptions controls citation graph generation. Vertex IDs are arrival
// indexes; directed citations point from the newer vertex to an older vertex.
type CitationOptions struct {
	Seed     *uint64
	Directed bool
}

func traitGrowthInputs(vertexCount, stepCount int, preference Matrix, options TraitGrowthOptions) (int, *realVector, *cMatrix, error) {
	if err := validateConstructorSize("vertex count", vertexCount); err != nil {
		return 0, nil, nil, err
	}
	if err := validateConstructorSize("connections per step", stepCount); err != nil {
		return 0, nil, nil, err
	}
	types, columns := preference.Dims()
	if types < 1 || columns != types {
		return 0, nil, nil, fmt.Errorf("igraph: trait preference matrix must be non-empty and square")
	}
	if err := validateProbabilityMatrix(preference, types, types, !options.Directed, "trait preference matrix"); err != nil {
		return 0, nil, nil, err
	}
	var distribution *realVector
	if len(options.TypeWeights) > 0 {
		if len(options.TypeWeights) != types {
			return 0, nil, nil, fmt.Errorf("igraph: type weight length %d does not match type count %d", len(options.TypeWeights), types)
		}
		sum := 0.0
		for i, v := range options.TypeWeights {
			if v < 0 || math.IsNaN(v) || math.IsInf(v, 0) {
				return 0, nil, nil, fmt.Errorf("igraph: type weight %d must be finite and non-negative", i)
			}
			sum += v
		}
		if !(sum > 0) {
			return 0, nil, nil, fmt.Errorf("igraph: type weights must have positive sum")
		}
		var err error
		distribution, err = newRealVector(options.TypeWeights)
		if err != nil {
			return 0, nil, nil, err
		}
	}
	matrix, err := newCMatrix(preference)
	if err != nil {
		if distribution != nil {
			distribution.close()
		}
		return 0, nil, nil, err
	}
	return types, distribution, matrix, nil
}

func traitGrowthResult(operation string, seed *uint64, distribution *realVector, matrix *cMatrix, call func(*C.igraph_t, *C.igraph_vector_t, *C.igraph_matrix_t, *C.igraph_vector_int_t) C.igraph_error_t) (PreferenceGraphResult, error) {
	defer matrix.close()
	if distribution != nil {
		defer distribution.close()
	}
	types, err := newIntVector(nil)
	if err != nil {
		return PreferenceGraphResult{}, err
	}
	defer types.close()
	var dist *C.igraph_vector_t
	if distribution != nil {
		dist = &distribution.value
	}
	graph, err := generateGraph(operation, seed, func(graph *C.igraph_t) C.igraph_error_t { return call(graph, dist, &matrix.value, &types.value) })
	if err != nil {
		return PreferenceGraphResult{}, err
	}
	values, err := types.slice()
	if err != nil {
		graph.Close()
		return PreferenceGraphResult{}, err
	}
	return PreferenceGraphResult{Graph: graph, Types: values}, nil
}

// CallawayTraitsGame grows one vertex per step, samples its type, then tries
// connectionTrials uniformly selected endpoint pairs. Returned types are a
// Go-owned vertex-ID-aligned copy and Graph must be closed.
//
//igraph:bind igraph_callaway_traits_game
func CallawayTraitsGame(vertexCount, connectionTrials int, preference Matrix, options TraitGrowthOptions) (PreferenceGraphResult, error) {
	types, dist, matrix, err := traitGrowthInputs(vertexCount, connectionTrials, preference, options)
	if err != nil {
		return PreferenceGraphResult{}, err
	}
	return traitGrowthResult("igraph_callaway_traits_game", options.Seed, dist, matrix, func(graph *C.igraph_t, d *C.igraph_vector_t, p *C.igraph_matrix_t, out *C.igraph_vector_int_t) C.igraph_error_t {
		return C.go_igraph_callaway_traits_game(graph, C.igraph_int_t(vertexCount), C.igraph_int_t(types), C.igraph_int_t(connectionTrials), d, p, booltoint(options.Directed), out)
	})
}

// EstablishmentGame grows one vertex per step. Each vertex after the first
// candidateCount arrivals samples that many distinct older candidates and
// accepts connections according to preference. Results are independently owned.
//
//igraph:bind igraph_establishment_game
func EstablishmentGame(vertexCount, candidateCount int, preference Matrix, options TraitGrowthOptions) (PreferenceGraphResult, error) {
	types, dist, matrix, err := traitGrowthInputs(vertexCount, candidateCount, preference, options)
	if err != nil {
		return PreferenceGraphResult{}, err
	}
	return traitGrowthResult("igraph_establishment_game", options.Seed, dist, matrix, func(graph *C.igraph_t, d *C.igraph_vector_t, p *C.igraph_matrix_t, out *C.igraph_vector_int_t) C.igraph_error_t {
		return C.go_igraph_establishment_game(graph, C.igraph_int_t(vertexCount), C.igraph_int_t(types), C.igraph_int_t(candidateCount), d, p, booltoint(options.Directed), out)
	})
}

func validateCitationSize(vertexCount, edges int) error {
	if err := validateConstructorSize("citation edge count per arrival", edges); err != nil {
		return err
	}
	if vertexCount > 1 && edges > int(^uint(0)>>1)/(vertexCount-1) {
		return fmt.Errorf("igraph: citation edge count overflows int")
	}
	return nil
}

// LastCitationGame weights older vertices by the arrival-time bin of their
// most recent citation. agePreferences contains one value per age bin followed
// by the strictly positive never-cited weight. Multiple edges are possible.
//
//igraph:bind igraph_lastcit_game
func LastCitationGame(vertexCount, edgesPerArrival int, agePreferences []float64, options CitationOptions) (*Graph, error) {
	if err := validateConstructorSize("vertex count", vertexCount); err != nil {
		return nil, err
	}
	if err := validateCitationSize(vertexCount, edgesPerArrival); err != nil {
		return nil, err
	}
	if len(agePreferences) < 2 {
		return nil, fmt.Errorf("igraph: citation preferences require at least one age bin and a never-cited weight")
	}
	for i, v := range agePreferences {
		if v < 0 || math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, fmt.Errorf("igraph: citation preference %d must be finite and non-negative", i)
		}
	}
	if agePreferences[len(agePreferences)-1] <= 0 {
		return nil, fmt.Errorf("igraph: never-cited preference must be positive")
	}
	pref, err := newRealVector(agePreferences)
	if err != nil {
		return nil, err
	}
	defer pref.close()
	return generateGraph("igraph_lastcit_game", options.Seed, func(graph *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_lastcit_game(graph, C.igraph_int_t(vertexCount), C.igraph_int_t(edgesPerArrival), C.igraph_int_t(len(agePreferences)-1), &pref.value, booltoint(options.Directed))
	})
}

func citationTypes(types []int, edges int) (*intVector, int, error) {
	if err := validateCitationSize(len(types), edges); err != nil {
		return nil, 0, err
	}
	typeCount := 0
	for i, v := range types {
		if v < 0 {
			return nil, 0, fmt.Errorf("igraph: vertex type at index %d is negative", i)
		}
		if v == int(^uint(0)>>1) {
			return nil, 0, fmt.Errorf("igraph: vertex type at index %d is too large", i)
		}
		if v+1 > typeCount {
			typeCount = v + 1
		}
	}
	vector, err := newIntVector(types)
	return vector, typeCount, err
}

// CitedTypeGame grows a citation graph whose target weights depend only on the
// cited vertex type. types is vertex-ID/arrival-order aligned and borrowed.
// Preference length must equal max(types)+1. Multiple edges are possible.
//
//igraph:bind igraph_cited_type_game
func CitedTypeGame(types []int, preferences []float64, edgesPerArrival int, options CitationOptions) (*Graph, error) {
	ct, typeCount, err := citationTypes(types, edgesPerArrival)
	if err != nil {
		return nil, err
	}
	defer ct.close()
	if len(preferences) != typeCount {
		return nil, fmt.Errorf("igraph: cited-type preference length %d does not match type count %d", len(preferences), typeCount)
	}
	for i, v := range preferences {
		if v < 0 || math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, fmt.Errorf("igraph: cited-type preference %d must be finite and non-negative", i)
		}
	}
	if len(types) > 1 && preferences[types[0]] <= 0 {
		return nil, fmt.Errorf("igraph: first vertex type must have positive cited weight to avoid self-citations")
	}
	cp, err := newRealVector(preferences)
	if err != nil {
		return nil, err
	}
	defer cp.close()
	return generateGraph("igraph_cited_type_game", options.Seed, func(graph *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_cited_type_game(graph, C.igraph_int_t(len(types)), &ct.value, &cp.value, C.igraph_int_t(edgesPerArrival), booltoint(options.Directed))
	})
}

// CitingCitedTypeGame grows a citation graph whose target weights use the
// citing type as matrix rows and cited type as columns. Types and preference
// are borrowed synchronously; multiple edges are possible.
//
//igraph:bind igraph_citing_cited_type_game
func CitingCitedTypeGame(types []int, preference Matrix, edgesPerArrival int, options CitationOptions) (*Graph, error) {
	ct, typeCount, err := citationTypes(types, edgesPerArrival)
	if err != nil {
		return nil, err
	}
	defer ct.close()
	rows, columns := preference.Dims()
	if rows != typeCount || columns != typeCount {
		return nil, fmt.Errorf("igraph: citing/cited preference dimensions are %dx%d, want %dx%d", rows, columns, typeCount, typeCount)
	}
	for i, row := range preference.Rows() {
		for j, v := range row {
			if v < 0 || math.IsNaN(v) || math.IsInf(v, 0) {
				return nil, fmt.Errorf("igraph: citing/cited preference[%d,%d] must be finite and non-negative", i, j)
			}
		}
	}
	cp, err := newCMatrix(preference)
	if err != nil {
		return nil, err
	}
	defer cp.close()
	return generateGraph("igraph_citing_cited_type_game", options.Seed, func(graph *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_citing_cited_type_game(graph, C.igraph_int_t(len(types)), &ct.value, &cp.value, C.igraph_int_t(edgesPerArrival), booltoint(options.Directed))
	})
}

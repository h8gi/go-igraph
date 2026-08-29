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

// GrowingRandomOptions controls uniform graph growth. Vertex IDs are arrival
// indexes. Citation makes every edge point from the newly arrived vertex to an
// older vertex; it is meaningful for directed graphs.
type GrowingRandomOptions struct {
	Seed     *uint64
	Directed bool
	Citation bool
}

// ForestFireOptions controls forest-fire growth. BackwardFactor multiplied by
// forwardProbability is the backward burning probability.
type ForestFireOptions struct {
	Seed           *uint64
	BackwardFactor float64
	Ambassadors    int
	Directed       bool
}

// AttachmentSchedule selects either a constant number of new edges per
// arrival or a vertex-ID-aligned sequence. A nil or empty OutSequence selects
// EdgesPerStep. Inputs are borrowed only for the synchronous call.
type AttachmentSchedule struct {
	EdgesPerStep int
	OutSequence  []int
}

// BarabasiAgingOptions controls preferential attachment with vertex aging.
type BarabasiAgingOptions struct {
	Seed               *uint64
	Schedule           AttachmentSchedule
	OutPreference      bool
	AttachmentExponent float64
	AgingExponent      float64
	AgingBins          int
	ZeroDegreeAppeal   float64
	ZeroAgeAppeal      float64
	DegreeCoefficient  float64
	AgeCoefficient     float64
	Directed           bool
}

// RecentDegreeOptions controls attachment based on edges gained within Window
// arrival steps.
type RecentDegreeOptions struct {
	Seed          *uint64
	Schedule      AttachmentSchedule
	Exponent      float64
	Window        int
	OutPreference bool
	ZeroAppeal    float64
	Directed      bool
}

// RecentDegreeAgingOptions combines recent-degree attachment and aging.
type RecentDegreeAgingOptions struct {
	Seed               *uint64
	Schedule           AttachmentSchedule
	OutPreference      bool
	AttachmentExponent float64
	AgingExponent      float64
	AgingBins          int
	Window             int
	ZeroAppeal         float64
	Directed           bool
}

// GrowingRandomGame starts with vertex zero and appends vertices in ID order,
// adding edgesPerStep random edges on every subsequent arrival. The returned
// graph is independently Go-owned and must be closed.
//
//igraph:bind igraph_growing_random_game
func GrowingRandomGame(vertexCount, edgesPerStep int, options GrowingRandomOptions) (*Graph, error) {
	if err := validateConstructorSize("vertex count", vertexCount); err != nil {
		return nil, err
	}
	if err := validateConstructorSize("edges per step", edgesPerStep); err != nil {
		return nil, err
	}
	if err := validateGrowthEdgeCount(vertexCount, edgesPerStep, nil); err != nil {
		return nil, err
	}
	return generateGraph("igraph_growing_random_game", options.Seed, func(graph *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_growing_random_game(graph, C.igraph_int_t(vertexCount), C.igraph_int_t(edgesPerStep), booltoint(options.Directed), booltoint(options.Citation))
	})
}

// ForestFireGame appends vertices in ID order and recursively burns neighbors
// from uniformly selected ambassadors. forwardProbability and its product with
// BackwardFactor must both be in [0,1). The returned graph must be closed.
//
//igraph:bind igraph_forest_fire_game
func ForestFireGame(vertexCount int, forwardProbability float64, options ForestFireOptions) (*Graph, error) {
	if err := validateConstructorSize("vertex count", vertexCount); err != nil {
		return nil, err
	}
	if math.IsNaN(forwardProbability) || math.IsInf(forwardProbability, 0) || forwardProbability < 0 || forwardProbability >= 1 {
		return nil, fmt.Errorf("igraph: forward probability must be finite and in [0, 1): %g", forwardProbability)
	}
	if math.IsNaN(options.BackwardFactor) || math.IsInf(options.BackwardFactor, 0) || options.BackwardFactor < 0 || forwardProbability*options.BackwardFactor >= 1 {
		return nil, fmt.Errorf("igraph: backward factor must be finite, non-negative, and produce a probability below one: %g", options.BackwardFactor)
	}
	if err := validateConstructorSize("ambassador count", options.Ambassadors); err != nil {
		return nil, err
	}
	return generateGraph("igraph_forest_fire_game", options.Seed, func(graph *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_forest_fire_game(graph, C.igraph_int_t(vertexCount), C.igraph_real_t(forwardProbability), C.igraph_real_t(options.BackwardFactor), C.igraph_int_t(options.Ambassadors), booltoint(options.Directed))
	})
}

func validateGrowthEdgeCount(vertexCount, edgesPerStep int, sequence []int) error {
	maximum := int(^uint(0) >> 1)
	if len(sequence) == 0 {
		if vertexCount > 1 && edgesPerStep > maximum/(vertexCount-1) {
			return fmt.Errorf("igraph: graph-growth edge count overflows int")
		}
		return nil
	}
	if len(sequence) != vertexCount {
		return fmt.Errorf("igraph: output sequence length %d does not match vertex count %d", len(sequence), vertexCount)
	}
	sum := 0
	for i, value := range sequence {
		if value < 0 {
			return fmt.Errorf("igraph: output sequence value at index %d is negative: %d", i, value)
		}
		if i > 0 {
			if value > maximum-sum {
				return fmt.Errorf("igraph: output sequence edge count overflows int")
			}
			sum += value
		}
	}
	return nil
}

func growthSchedule(vertexCount int, schedule AttachmentSchedule) (*intVector, *C.igraph_vector_int_t, error) {
	if err := validateConstructorSize("edges per step", schedule.EdgesPerStep); err != nil {
		return nil, nil, err
	}
	if err := validateGrowthEdgeCount(vertexCount, schedule.EdgesPerStep, schedule.OutSequence); err != nil {
		return nil, nil, err
	}
	if len(schedule.OutSequence) == 0 {
		return nil, nil, nil
	}
	vector, err := newIntVector(schedule.OutSequence)
	if err != nil {
		return nil, nil, err
	}
	return vector, &vector.value, nil
}

func finite(value float64, name string) error {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("igraph: %s must be finite: %g", name, value)
	}
	return nil
}

func nonNegativeFinite(value float64, name string) error {
	if err := finite(value, name); err != nil {
		return err
	}
	if value < 0 {
		return fmt.Errorf("igraph: %s must be non-negative: %g", name, value)
	}
	return nil
}

// BarabasiAgingGame appends vertices in ID order using multiplicative degree
// and age attractiveness. The returned graph must be closed.
//
//igraph:bind igraph_barabasi_aging_game
func BarabasiAgingGame(vertexCount int, options BarabasiAgingOptions) (*Graph, error) {
	if err := validateConstructorSize("vertex count", vertexCount); err != nil {
		return nil, err
	}
	if options.AgingBins <= 0 {
		return nil, fmt.Errorf("igraph: aging bins must be positive: %d", options.AgingBins)
	}
	if err := finite(options.AttachmentExponent, "attachment exponent"); err != nil {
		return nil, err
	}
	if err := finite(options.AgingExponent, "aging exponent"); err != nil {
		return nil, err
	}
	if err := nonNegativeFinite(options.ZeroDegreeAppeal, "zero-degree appeal"); err != nil {
		return nil, err
	}
	if err := nonNegativeFinite(options.ZeroAgeAppeal, "zero-age appeal"); err != nil {
		return nil, err
	}
	if err := nonNegativeFinite(options.DegreeCoefficient, "degree coefficient"); err != nil {
		return nil, err
	}
	if err := nonNegativeFinite(options.AgeCoefficient, "age coefficient"); err != nil {
		return nil, err
	}
	vector, pointer, err := growthSchedule(vertexCount, options.Schedule)
	if err != nil {
		return nil, err
	}
	if vector != nil {
		defer vector.close()
	}
	return generateGraph("igraph_barabasi_aging_game", options.Seed, func(graph *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_barabasi_aging_game(graph, C.igraph_int_t(vertexCount), C.igraph_int_t(options.Schedule.EdgesPerStep), pointer, booltoint(options.OutPreference), C.igraph_real_t(options.AttachmentExponent), C.igraph_real_t(options.AgingExponent), C.igraph_int_t(options.AgingBins), C.igraph_real_t(options.ZeroDegreeAppeal), C.igraph_real_t(options.ZeroAgeAppeal), C.igraph_real_t(options.DegreeCoefficient), C.igraph_real_t(options.AgeCoefficient), booltoint(options.Directed))
	})
}

// RecentDegreeGame appends vertices in ID order and weights old vertices by
// edges gained during Window recent arrivals. The returned graph must be closed.
//
//igraph:bind igraph_recent_degree_game
func RecentDegreeGame(vertexCount int, options RecentDegreeOptions) (*Graph, error) {
	if err := validateConstructorSize("vertex count", vertexCount); err != nil {
		return nil, err
	}
	if err := finite(options.Exponent, "attachment exponent"); err != nil {
		return nil, err
	}
	if err := validateConstructorSize("recent-degree window", options.Window); err != nil {
		return nil, err
	}
	if err := nonNegativeFinite(options.ZeroAppeal, "zero appeal"); err != nil {
		return nil, err
	}
	vector, pointer, err := growthSchedule(vertexCount, options.Schedule)
	if err != nil {
		return nil, err
	}
	if vector != nil {
		defer vector.close()
	}
	return generateGraph("igraph_recent_degree_game", options.Seed, func(graph *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_recent_degree_game(graph, C.igraph_int_t(vertexCount), C.igraph_real_t(options.Exponent), C.igraph_int_t(options.Window), C.igraph_int_t(options.Schedule.EdgesPerStep), pointer, booltoint(options.OutPreference), C.igraph_real_t(options.ZeroAppeal), booltoint(options.Directed))
	})
}

// RecentDegreeAgingGame appends vertices in ID order and combines recent-degree
// and age attractiveness. The returned graph must be closed.
//
//igraph:bind igraph_recent_degree_aging_game
func RecentDegreeAgingGame(vertexCount int, options RecentDegreeAgingOptions) (*Graph, error) {
	if err := validateConstructorSize("vertex count", vertexCount); err != nil {
		return nil, err
	}
	if options.AgingBins <= 0 {
		return nil, fmt.Errorf("igraph: aging bins must be positive: %d", options.AgingBins)
	}
	if err := validateConstructorSize("recent-degree window", options.Window); err != nil {
		return nil, err
	}
	if err := finite(options.AttachmentExponent, "attachment exponent"); err != nil {
		return nil, err
	}
	if err := finite(options.AgingExponent, "aging exponent"); err != nil {
		return nil, err
	}
	if err := nonNegativeFinite(options.ZeroAppeal, "zero appeal"); err != nil {
		return nil, err
	}
	vector, pointer, err := growthSchedule(vertexCount, options.Schedule)
	if err != nil {
		return nil, err
	}
	if vector != nil {
		defer vector.close()
	}
	return generateGraph("igraph_recent_degree_aging_game", options.Seed, func(graph *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_recent_degree_aging_game(graph, C.igraph_int_t(vertexCount), C.igraph_int_t(options.Schedule.EdgesPerStep), pointer, booltoint(options.OutPreference), C.igraph_real_t(options.AttachmentExponent), C.igraph_real_t(options.AgingExponent), C.igraph_int_t(options.AgingBins), C.igraph_int_t(options.Window), C.igraph_real_t(options.ZeroAppeal), booltoint(options.Directed))
	})
}

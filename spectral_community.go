package igraph

/*
#include <stdint.h>
#include <igraph.h>
#include "algorithm_cgo.h"
#include "community_cgo.h"
#include "spectral_community_cgo.h"
*/
import "C"

import (
	"fmt"
	"math"
)

// SpincommUpdateRule specifies the spin update rule for Spinglass community detection.
type SpincommUpdateRule int

const (
	// SpincommUpdateSimple updates spins simple/randomly.
	SpincommUpdateSimple SpincommUpdateRule = C.IGRAPH_SPINCOMM_UPDATE_SIMPLE
	// SpincommUpdateConfig updates spins configurationally.
	SpincommUpdateConfig SpincommUpdateRule = C.IGRAPH_SPINCOMM_UPDATE_CONFIG
)

func (u SpincommUpdateRule) cValue() (C.igraph_spincomm_update_t, error) {
	switch u {
	case SpincommUpdateSimple, SpincommUpdateConfig:
		return C.igraph_spincomm_update_t(u), nil
	default:
		return 0, fmt.Errorf("igraph: invalid spincomm update rule: %d", u)
	}
}

// SpinglassImplementation specifies the implementation variant for Spinglass community detection.
type SpinglassImplementation int

const (
	// SpinglassImplementationOriginal uses the original Potts model algorithm.
	SpinglassImplementationOriginal SpinglassImplementation = C.IGRAPH_SPINCOMM_IMP_ORIG
	// SpinglassImplementationNegative handles graphs with negative edge weights.
	SpinglassImplementationNegative SpinglassImplementation = C.IGRAPH_SPINCOMM_IMP_NEG
)

func (impl SpinglassImplementation) cValue() (C.igraph_spinglass_implementation_t, error) {
	switch impl {
	case SpinglassImplementationOriginal, SpinglassImplementationNegative:
		return C.igraph_spinglass_implementation_t(impl), nil
	default:
		return 0, fmt.Errorf("igraph: invalid spinglass implementation: %d", impl)
	}
}

// LeadingEigenvectorOptions controls the leading eigenvector community detection algorithm.
type LeadingEigenvectorOptions struct {
	// Weights optionally specifies edge weights.
	Weights []float64
	// Steps specifies the maximum number of steps (splits) to perform. If <= 0, defaults to -1 (unlimited).
	Steps int
	// Solver controls ARPACK solver convergence settings.
	Solver SpectralSolverOptions
	// Start indicates whether InitialMembership is provided as starting partition.
	Start bool
	// InitialMembership optionally provides initial community membership (must match vertex count).
	InitialMembership []int
	// Seed optionally seeds the package random number generator.
	Seed *uint64
}

// SpinglassOptions controls the Spinglass community detection algorithm.
type SpinglassOptions struct {
	// Weights optionally specifies edge weights.
	Weights []float64
	// Spins specifies the maximum number of communities (default: 25 if <= 0). Must be between 1 and 500.
	Spins int
	// ParallelUpdate enables parallel spin updates if true.
	ParallelUpdate bool
	// StartTemperature specifies the starting temperature (default: 1.0 if <= 0).
	StartTemperature float64
	// StopTemperature specifies the stopping temperature (default: 0.01 if <= 0).
	StopTemperature float64
	// CoolingFactor specifies the cooling factor (default: 0.99 if <= 0). Must be strictly between 0 and 1.
	CoolingFactor float64
	// UpdateRule specifies the spin update rule.
	UpdateRule SpincommUpdateRule
	// Gamma specifies the resolution/gamma parameter (default: 1.0 if <= 0).
	Gamma float64
	// Implementation specifies original vs negative implementation.
	Implementation SpinglassImplementation
	// Lambda specifies the lambda parameter for negative implementation (default: 1.0 if <= 0).
	Lambda float64
	// Seed optionally seeds the package random number generator.
	Seed *uint64
}

// SpinglassSingleOptions controls finding the community of a single vertex.
type SpinglassSingleOptions struct {
	// Weights optionally specifies edge weights.
	Weights []float64
	// Vertex is the target vertex index whose community is to be determined.
	Vertex int
	// Spins specifies the maximum number of communities (default: 25 if <= 0). Must be between 1 and 500.
	Spins int
	// UpdateRule specifies the spin update rule.
	UpdateRule SpincommUpdateRule
	// Gamma specifies the resolution/gamma parameter (default: 1.0 if <= 0).
	Gamma float64
	// Seed optionally seeds the package random number generator.
	Seed *uint64
}

// SpinglassSingleResult contains community membership and coupling metrics for a single vertex.
type SpinglassSingleResult struct {
	// Community contains the vertex IDs belonging to the same community as the target vertex.
	Community []int
	// Cohesion is the internal cohesion score of the community.
	Cohesion float64
	// Adhesion is the external adhesion score of the community.
	Adhesion float64
	// InnerLinks is the number of internal edges within the community.
	InnerLinks int
	// OuterLinks is the number of external edges connecting the community to the rest of the graph.
	OuterLinks int
}

// CommunityLeadingEigenvector finds community structure using the leading eigenvector algorithm.
//
// Inputs are borrowed; returned CommunityPartition is Go-owned.
//
//igraph:bind igraph_community_leading_eigenvector
func (g *Graph) CommunityLeadingEigenvector(options LeadingEigenvectorOptions) (CommunityPartition, error) {
	var partition CommunityPartition
	err := withRNG(options.Seed, func() error {
		p, err := g.executeFlat(func() (CommunityPartition, error) {
			vcount := int(C.igraph_vcount(&g.graph))
			if vcount == 0 {
				return partitionFromMembership(nil, math.NaN())
			}

			weightsVec, err := newOptionalEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)))
			if err != nil {
				return CommunityPartition{}, err
			}
			if weightsVec != nil {
				defer weightsVec.close()
			}

			maxIter, tol, err := validateSpectralSolver(options.Solver)
			if err != nil {
				return CommunityPartition{}, err
			}

			steps := options.Steps
			if steps <= 0 {
				steps = -1
			}

			var initMembership []int
			start := options.Start
			if len(options.InitialMembership) > 0 {
				if len(options.InitialMembership) != vcount {
					return CommunityPartition{}, fmt.Errorf("igraph: invalid initial membership length: %d (expected %d)", len(options.InitialMembership), vcount)
				}
				start = true
				initMembership = options.InitialMembership
			}

			memVec, err := newIntVector(initMembership)
			if err != nil {
				return CommunityPartition{}, err
			}
			defer memVec.close()

			var modularity C.igraph_real_t
			code := C.go_igraph_community_leading_eigenvector(
				&g.graph,
				edgeWeightPointer(weightsVec),
				nil,
				&memVec.value,
				C.igraph_integer_t(steps),
				C.int(maxIter),
				C.igraph_real_t(tol),
				&modularity,
				booltoint(start),
			)
			if code != 0 {
				return CommunityPartition{}, igraphError("igraph_community_leading_eigenvector", int(code))
			}

			memSlice, err := memVec.slice()
			if err != nil {
				return CommunityPartition{}, err
			}

			return partitionFromMembership(memSlice, float64(modularity))
		})
		if err != nil {
			return err
		}
		partition = p
		return nil
	})
	if err != nil {
		return CommunityPartition{}, err
	}
	return partition, nil
}

// CommunitySpinglass finds community structure using simulated annealing and a spinglass model.
//
// Inputs are borrowed; returned CommunityPartition is Go-owned.
//
//igraph:bind igraph_community_spinglass
func (g *Graph) CommunitySpinglass(options SpinglassOptions) (CommunityPartition, error) {
	return g.executeFlat(func() (CommunityPartition, error) {
		vcount := int(C.igraph_vcount(&g.graph))
		if vcount == 0 {
			return partitionFromMembership(nil, math.NaN())
		}

		weightsVec, err := newOptionalEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)))
		if err != nil {
			return CommunityPartition{}, err
		}
		if weightsVec != nil {
			defer weightsVec.close()
		}

		spins := options.Spins
		if spins <= 0 {
			spins = 25
		}
		if spins < 1 || spins > 500 {
			return CommunityPartition{}, fmt.Errorf("igraph: invalid spins: %d (must be between 1 and 500)", spins)
		}

		startTemp := options.StartTemperature
		if startTemp <= 0 {
			startTemp = 1.0
		}
		stopTemp := options.StopTemperature
		if stopTemp <= 0 {
			stopTemp = 0.01
		}
		if stopTemp >= startTemp {
			return CommunityPartition{}, fmt.Errorf("igraph: stop temperature (%g) must be less than start temperature (%g)", stopTemp, startTemp)
		}

		coolFact := options.CoolingFactor
		if coolFact <= 0 {
			coolFact = 0.99
		}
		if coolFact <= 0 || coolFact >= 1.0 {
			return CommunityPartition{}, fmt.Errorf("igraph: invalid cooling factor: %g (must be between 0 and 1)", coolFact)
		}

		gamma := options.Gamma
		if gamma <= 0 {
			gamma = 1.0
		}

		lambda := options.Lambda
		if lambda <= 0 {
			lambda = 1.0
		}

		updateRule, err := options.UpdateRule.cValue()
		if err != nil {
			return CommunityPartition{}, err
		}

		impl, err := options.Implementation.cValue()
		if err != nil {
			return CommunityPartition{}, err
		}

		memVec, err := newIntVector(nil)
		if err != nil {
			return CommunityPartition{}, err
		}
		defer memVec.close()

		csizeVec, err := newIntVector(nil)
		if err != nil {
			return CommunityPartition{}, err
		}
		defer csizeVec.close()

		var modularity C.igraph_real_t
		var temperature C.igraph_real_t

		errRNG := withRNG(options.Seed, func() error {
			code := C.go_igraph_community_spinglass(
				&g.graph,
				edgeWeightPointer(weightsVec),
				&modularity,
				&temperature,
				&memVec.value,
				&csizeVec.value,
				C.igraph_integer_t(spins),
				booltoint(options.ParallelUpdate),
				C.igraph_real_t(startTemp),
				C.igraph_real_t(stopTemp),
				C.igraph_real_t(coolFact),
				updateRule,
				C.igraph_real_t(gamma),
				impl,
				C.igraph_real_t(lambda),
			)
			if code != 0 {
				return igraphError("igraph_community_spinglass", int(code))
			}
			return nil
		})
		if errRNG != nil {
			return CommunityPartition{}, errRNG
		}

		memSlice, err := memVec.slice()
		if err != nil {
			return CommunityPartition{}, err
		}

		return partitionFromMembership(memSlice, float64(modularity))
	})
}

// CommunitySpinglassSingle identifies the community of a single vertex using the spinglass algorithm.
//
// Inputs are borrowed; returned SpinglassSingleResult is Go-owned.
//
//igraph:bind igraph_community_spinglass_single
func (g *Graph) CommunitySpinglassSingle(options SpinglassSingleOptions) (SpinglassSingleResult, error) {
	if g == nil {
		return SpinglassSingleResult{}, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if err := g.checkClosed(); err != nil {
		return SpinglassSingleResult{}, err
	}

	vcount := int(C.igraph_vcount(&g.graph))
	if options.Vertex < 0 || options.Vertex >= vcount {
		return SpinglassSingleResult{}, fmt.Errorf("igraph: invalid vertex index: %d (vcount: %d)", options.Vertex, vcount)
	}

	weightsVec, err := newOptionalEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return SpinglassSingleResult{}, err
	}
	if weightsVec != nil {
		defer weightsVec.close()
	}

	spins := options.Spins
	if spins <= 0 {
		spins = 25
	}
	if spins < 1 || spins > 500 {
		return SpinglassSingleResult{}, fmt.Errorf("igraph: invalid spins: %d (must be between 1 and 500)", spins)
	}

	gamma := options.Gamma
	if gamma <= 0 {
		gamma = 1.0
	}

	updateRule, err := options.UpdateRule.cValue()
	if err != nil {
		return SpinglassSingleResult{}, err
	}

	commVec, err := newIntVector(nil)
	if err != nil {
		return SpinglassSingleResult{}, err
	}
	defer commVec.close()

	var cohesion C.igraph_real_t
	var adhesion C.igraph_real_t
	var innerLinks C.igraph_real_t
	var outerLinks C.igraph_real_t

	errRNG := withRNG(options.Seed, func() error {
		code := C.go_igraph_community_spinglass_single(
			&g.graph,
			edgeWeightPointer(weightsVec),
			C.igraph_integer_t(options.Vertex),
			&commVec.value,
			&cohesion,
			&adhesion,
			&innerLinks,
			&outerLinks,
			C.igraph_integer_t(spins),
			updateRule,
			C.igraph_real_t(gamma),
		)
		if code != 0 {
			return igraphError("igraph_community_spinglass_single", int(code))
		}
		return nil
	})
	if errRNG != nil {
		return SpinglassSingleResult{}, errRNG
	}

	commSlice, err := commVec.slice()
	if err != nil {
		return SpinglassSingleResult{}, err
	}

	return SpinglassSingleResult{
		Community:  commSlice,
		Cohesion:   float64(cohesion),
		Adhesion:   float64(adhesion),
		InnerLinks: int(innerLinks),
		OuterLinks: int(outerLinks),
	}, nil
}

// CommunityOptimalModularity calculates the exact modularity-maximizing community partition for small graphs.
//
// Inputs are borrowed; returned CommunityPartition is Go-owned.
//
//igraph:bind igraph_community_optimal_modularity
func (g *Graph) CommunityOptimalModularity(weights []float64) (CommunityPartition, error) {
	return g.executeFlat(func() (CommunityPartition, error) {
		vcount := int(C.igraph_vcount(&g.graph))
		if vcount == 0 {
			return partitionFromMembership(nil, math.NaN())
		}

		weightsVec, err := newOptionalEdgeWeights(weights, int(C.igraph_ecount(&g.graph)))
		if err != nil {
			return CommunityPartition{}, err
		}
		if weightsVec != nil {
			defer weightsVec.close()
		}

		memVec, err := newIntVector(nil)
		if err != nil {
			return CommunityPartition{}, err
		}
		defer memVec.close()

		var modularity C.igraph_real_t
		code := C.go_igraph_community_optimal_modularity(
			&g.graph,
			edgeWeightPointer(weightsVec),
			1.0,
			&modularity,
			&memVec.value,
		)
		if code != 0 {
			return CommunityPartition{}, igraphError("igraph_community_optimal_modularity", int(code))
		}

		memSlice, err := memVec.slice()
		if err != nil {
			return CommunityPartition{}, err
		}

		return partitionFromMembership(memSlice, float64(modularity))
	})
}

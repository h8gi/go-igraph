package igraph

/*
#include <stdint.h>
#include <igraph.h>
#include "algorithm_cgo.h"
#include "community_cgo.h"
#include "flat_community_cgo.h"
*/
import "C"

import (
	"fmt"
	"math"
)

// MultilevelOptions contains options for the Multilevel (Louvain) community detection algorithm.
type MultilevelOptions struct {
	// Weights optionally specifies edge weights. If nil, edges are unweighted.
	Weights []float64
	// Resolution controls the scale of community detection (default: 1.0).
	Resolution float64
	// Seed optionally seeds the package random number generator.
	Seed *uint64
}

// LeidenOptions contains options for the Leiden community detection algorithm.
type LeidenOptions struct {
	// EdgeWeights optionally specifies edge weights.
	EdgeWeights []float64
	// VertexOutWeights optionally specifies node out-weights.
	VertexOutWeights []float64
	// VertexInWeights optionally specifies node in-weights.
	VertexInWeights []float64
	// Resolution controls the scale of community detection (default: 1.0).
	Resolution float64
	// Beta randomness parameter for refinement (default: 0.01).
	Beta float64
	// Start indicates whether InitialMembership should be used as initial partition.
	Start bool
	// InitialMembership optionally provides the initial partition.
	InitialMembership []int
	// NIterations specifies the number of iterations to perform (default: 2 if <= 0).
	NIterations int
	// Seed optionally seeds the package random number generator.
	Seed *uint64
}

// LabelPropagationOptions contains options for label propagation community detection.
type LabelPropagationOptions struct {
	// Mode specifies edge direction for directed graphs (default: Out).
	Mode NeiMode
	// Weights optionally specifies edge weights.
	Weights []float64
	// InitialMembership optionally provides initial labels for vertices.
	InitialMembership []int
	// Fixed optionally specifies which vertex labels are fixed during propagation.
	Fixed []bool
	// Seed optionally seeds the package random number generator.
	Seed *uint64
}

// InfomapOptions contains options for the Infomap community detection algorithm.
type InfomapOptions struct {
	// EdgeWeights optionally specifies edge weights.
	EdgeWeights []float64
	// VertexWeights optionally specifies vertex weights.
	VertexWeights []float64
	// NTrials specifies the number of attempts to find the best partition (default: 10 if <= 0).
	NTrials int
	// IsRegularized enables regularization.
	IsRegularized bool
	// RegularizationStrength specifies the regularization strength.
	RegularizationStrength float64
	// Seed optionally seeds the package random number generator.
	Seed *uint64
}

// FluidOptions contains options for the Fluid Communities detection algorithm.
type FluidOptions struct {
	// Seed optionally seeds the package random number generator.
	Seed *uint64
}

func newOptionalVertexWeights(values []float64, vertexCount int) (*realVector, error) {
	if len(values) == 0 {
		return nil, nil
	}
	if len(values) != vertexCount {
		return nil, fmt.Errorf("igraph: invalid vertex weight vector length: %d (expected %d)", len(values), vertexCount)
	}
	return newRealVector(values)
}

func partitionFromMembership(membership []int, modularity float64) (CommunityPartition, error) {
	if len(membership) == 0 {
		return CommunityPartition{
			Membership:     []int{},
			CommunityCount: 0,
			Sizes:          []int{},
			Modularity:     modularity,
		}, nil
	}
	reindexed, _, count, err := ReindexMembership(membership)
	if err != nil {
		return CommunityPartition{}, err
	}
	sizes := make([]int, count)
	for _, comm := range reindexed {
		sizes[comm]++
	}
	return CommunityPartition{
		Membership:     reindexed,
		CommunityCount: count,
		Sizes:          sizes,
		Modularity:     modularity,
	}, nil
}

func (g *Graph) executeFlat(fn func() (CommunityPartition, error)) (CommunityPartition, error) {
	if g == nil {
		return CommunityPartition{}, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if err := g.checkClosed(); err != nil {
		return CommunityPartition{}, err
	}
	return fn()
}

// CommunityMultilevel finds community structure using the Multilevel (Louvain) algorithm.
//
// Inputs are borrowed; returned CommunityPartition is Go-owned.
//
//igraph:bind igraph_community_multilevel
func (g *Graph) CommunityMultilevel(options MultilevelOptions) (CommunityPartition, error) {
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

		resolution := options.Resolution
		if resolution <= 0 {
			resolution = 1.0
		}

		var memVec C.igraph_vector_int_t
		if code := C.go_igraph_vector_int_init(&memVec, 0); code != 0 {
			return CommunityPartition{}, igraphError("initialize membership vector", int(code))
		}
		defer C.igraph_vector_int_destroy(&memVec)

		var modVec C.igraph_vector_t
		if code := C.go_igraph_vector_init(&modVec, 0); code != 0 {
			return CommunityPartition{}, igraphError("initialize modularity vector", int(code))
		}
		defer C.igraph_vector_destroy(&modVec)

		var finalModularity float64 = math.NaN()

		errRNG := withRNG(options.Seed, func() error {
			code := C.go_igraph_community_multilevel(
				&g.graph,
				edgeWeightPointer(weightsVec),
				C.igraph_real_t(resolution),
				&memVec,
				nil,
				&modVec,
			)
			if code != 0 {
				return igraphError("igraph_community_multilevel", int(code))
			}
			return nil
		})
		if errRNG != nil {
			return CommunityPartition{}, errRNG
		}

		memSlice, err := intVectorSlice(&memVec)
		if err != nil {
			return CommunityPartition{}, err
		}
		modSlice, err := (&realVector{value: modVec}).slice()
		if err == nil && len(modSlice) > 0 {
			finalModularity = modSlice[len(modSlice)-1]
		}

		return partitionFromMembership(memSlice, finalModularity)
	})
}

// CommunityLeiden finds community structure using the Leiden algorithm.
//
// Inputs are borrowed; returned CommunityPartition is Go-owned.
// If InitialMembership is non-empty, Start is automatically treated as true.
//
//igraph:bind igraph_community_leiden
func (g *Graph) CommunityLeiden(options LeidenOptions) (CommunityPartition, error) {
	return g.executeFlat(func() (CommunityPartition, error) {
		vcount := int(C.igraph_vcount(&g.graph))
		if vcount == 0 {
			return partitionFromMembership(nil, math.NaN())
		}

		edgeWeights, err := newOptionalEdgeWeights(options.EdgeWeights, int(C.igraph_ecount(&g.graph)))
		if err != nil {
			return CommunityPartition{}, err
		}
		if edgeWeights != nil {
			defer edgeWeights.close()
		}

		vertexOutWeights, err := newOptionalVertexWeights(options.VertexOutWeights, vcount)
		if err != nil {
			return CommunityPartition{}, err
		}
		if vertexOutWeights != nil {
			defer vertexOutWeights.close()
		}

		vertexInWeights, err := newOptionalVertexWeights(options.VertexInWeights, vcount)
		if err != nil {
			return CommunityPartition{}, err
		}
		if vertexInWeights != nil {
			defer vertexInWeights.close()
		}

		resolution := options.Resolution
		if resolution <= 0 {
			resolution = 1.0
		}
		beta := options.Beta
		if beta <= 0 {
			beta = 0.01
		}
		nIter := options.NIterations
		if nIter <= 0 {
			nIter = 2
		}
		cNIter, err := intToIgraphInt(nIter, "n_iterations")
		if err != nil {
			return CommunityPartition{}, err
		}

		useStart := options.Start
		if len(options.InitialMembership) > 0 {
			if len(options.InitialMembership) != vcount {
				return CommunityPartition{}, fmt.Errorf("igraph: invalid initial membership length: %d (expected %d)", len(options.InitialMembership), vcount)
			}
			useStart = true
		}

		var memVec *intVector
		if useStart && len(options.InitialMembership) > 0 {
			memVec, err = newIntVector(options.InitialMembership)
			if err != nil {
				return CommunityPartition{}, err
			}
		} else {
			memVec, err = newIntVector(make([]int, 0))
			if err != nil {
				return CommunityPartition{}, err
			}
		}
		defer memVec.close()

		var nbClusters C.igraph_int_t
		var quality C.igraph_real_t

		errRNG := withRNG(options.Seed, func() error {
			code := C.go_igraph_community_leiden(
				&g.graph,
				edgeWeightPointer(edgeWeights),
				edgeWeightPointer(vertexOutWeights),
				edgeWeightPointer(vertexInWeights),
				C.igraph_real_t(resolution),
				C.igraph_real_t(beta),
				booltoint(useStart),
				cNIter,
				&memVec.value,
				&nbClusters,
				&quality,
			)
			if code != 0 {
				return igraphError("igraph_community_leiden", int(code))
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

		return partitionFromMembership(memSlice, float64(quality))
	})
}

// CommunityLabelPropagation finds community structure using label propagation.
//
// Inputs are borrowed; returned CommunityPartition is Go-owned.
//
//igraph:bind igraph_community_label_propagation
func (g *Graph) CommunityLabelPropagation(options LabelPropagationOptions) (CommunityPartition, error) {
	return g.executeFlat(func() (CommunityPartition, error) {
		vcount := int(C.igraph_vcount(&g.graph))
		if vcount == 0 {
			return partitionFromMembership(nil, math.NaN())
		}

		cMode, err := options.Mode.cValue()
		if err != nil {
			return CommunityPartition{}, err
		}

		weightsVec, err := newOptionalEdgeWeights(options.Weights, int(C.igraph_ecount(&g.graph)))
		if err != nil {
			return CommunityPartition{}, err
		}
		if weightsVec != nil {
			defer weightsVec.close()
		}

		var initVec *intVector
		if len(options.InitialMembership) > 0 {
			if len(options.InitialMembership) != vcount {
				return CommunityPartition{}, fmt.Errorf("igraph: invalid initial membership length: %d (expected %d)", len(options.InitialMembership), vcount)
			}
			initVec, err = newIntVector(options.InitialMembership)
			if err != nil {
				return CommunityPartition{}, err
			}
			defer initVec.close()
		}

		var fixedVec *boolVector
		if len(options.Fixed) > 0 {
			if len(options.Fixed) != vcount {
				return CommunityPartition{}, fmt.Errorf("igraph: invalid fixed mask length: %d (expected %d)", len(options.Fixed), vcount)
			}
			fixedVec, err = newBoolVector(options.Fixed)
			if err != nil {
				return CommunityPartition{}, err
			}
			defer fixedVec.close()
		}

		var memVec C.igraph_vector_int_t
		if code := C.go_igraph_vector_int_init(&memVec, 0); code != 0 {
			return CommunityPartition{}, igraphError("initialize membership vector", int(code))
		}
		defer C.igraph_vector_int_destroy(&memVec)

		var initPtr *C.igraph_vector_int_t
		if initVec != nil {
			initPtr = &initVec.value
		}
		var fixedPtr *C.igraph_vector_bool_t
		if fixedVec != nil {
			fixedPtr = &fixedVec.value
		}

		errRNG := withRNG(options.Seed, func() error {
			code := C.go_igraph_community_label_propagation(
				&g.graph,
				&memVec,
				cMode,
				edgeWeightPointer(weightsVec),
				initPtr,
				fixedPtr,
				C.IGRAPH_LPA_DOMINANCE,
			)
			if code != 0 {
				return igraphError("igraph_community_label_propagation", int(code))
			}
			return nil
		})
		if errRNG != nil {
			return CommunityPartition{}, errRNG
		}

		memSlice, err := intVectorSlice(&memVec)
		if err != nil {
			return CommunityPartition{}, err
		}

		return partitionFromMembership(memSlice, math.NaN())
	})
}

// CommunityInfomap finds community structure by minimizing the map equation description length.
//
// Inputs are borrowed; returned CommunityPartition is Go-owned.
// Modularity field in the returned CommunityPartition contains the calculated codelength score.
//
//igraph:bind igraph_community_infomap
func (g *Graph) CommunityInfomap(options InfomapOptions) (CommunityPartition, error) {
	return g.executeFlat(func() (CommunityPartition, error) {
		vcount := int(C.igraph_vcount(&g.graph))
		if vcount == 0 {
			return partitionFromMembership(nil, math.NaN())
		}

		edgeWeights, err := newOptionalEdgeWeights(options.EdgeWeights, int(C.igraph_ecount(&g.graph)))
		if err != nil {
			return CommunityPartition{}, err
		}
		if edgeWeights != nil {
			defer edgeWeights.close()
		}

		vertexWeights, err := newOptionalVertexWeights(options.VertexWeights, vcount)
		if err != nil {
			return CommunityPartition{}, err
		}
		if vertexWeights != nil {
			defer vertexWeights.close()
		}

		nTrials := options.NTrials
		if nTrials <= 0 {
			nTrials = 10
		}
		cNTrials, err := intToIgraphInt(nTrials, "nb_trials")
		if err != nil {
			return CommunityPartition{}, err
		}

		var memVec C.igraph_vector_int_t
		if code := C.go_igraph_vector_int_init(&memVec, 0); code != 0 {
			return CommunityPartition{}, igraphError("initialize membership vector", int(code))
		}
		defer C.igraph_vector_int_destroy(&memVec)

		var codelength C.igraph_real_t

		errRNG := withRNG(options.Seed, func() error {
			code := C.go_igraph_community_infomap(
				&g.graph,
				edgeWeightPointer(edgeWeights),
				edgeWeightPointer(vertexWeights),
				cNTrials,
				booltoint(options.IsRegularized),
				C.igraph_real_t(options.RegularizationStrength),
				&memVec,
				&codelength,
			)
			if code != 0 {
				return igraphError("igraph_community_infomap", int(code))
			}
			return nil
		})
		if errRNG != nil {
			return CommunityPartition{}, errRNG
		}

		memSlice, err := intVectorSlice(&memVec)
		if err != nil {
			return CommunityPartition{}, err
		}

		return partitionFromMembership(memSlice, float64(codelength))
	})
}

// CommunityFluid finds community structure based on the interaction of fluids in topology.
//
// Inputs are borrowed; returned CommunityPartition is Go-owned.
//
//igraph:bind igraph_community_fluid_communities
func (g *Graph) CommunityFluid(noOfCommunities int, options FluidOptions) (CommunityPartition, error) {
	return g.executeFlat(func() (CommunityPartition, error) {
		vcount := int(C.igraph_vcount(&g.graph))
		if vcount == 0 {
			return partitionFromMembership(nil, math.NaN())
		}

		if noOfCommunities <= 0 {
			return CommunityPartition{}, fmt.Errorf("igraph: invalid number of communities: %d (must be > 0)", noOfCommunities)
		}
		if noOfCommunities > vcount {
			return CommunityPartition{}, fmt.Errorf("igraph: number of communities exceeds vertex count: %d > %d", noOfCommunities, vcount)
		}
		cNoOfComm, err := intToIgraphInt(noOfCommunities, "no_of_communities")
		if err != nil {
			return CommunityPartition{}, err
		}

		var memVec C.igraph_vector_int_t
		if code := C.go_igraph_vector_int_init(&memVec, 0); code != 0 {
			return CommunityPartition{}, igraphError("initialize membership vector", int(code))
		}
		defer C.igraph_vector_int_destroy(&memVec)

		errRNG := withRNG(options.Seed, func() error {
			code := C.go_igraph_community_fluid_communities(
				&g.graph,
				cNoOfComm,
				&memVec,
			)
			if code != 0 {
				return igraphError("igraph_community_fluid_communities", int(code))
			}
			return nil
		})
		if errRNG != nil {
			return CommunityPartition{}, errRNG
		}

		memSlice, err := intVectorSlice(&memVec)
		if err != nil {
			return CommunityPartition{}, err
		}

		return partitionFromMembership(memSlice, math.NaN())
	})
}

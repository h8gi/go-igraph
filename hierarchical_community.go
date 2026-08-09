package igraph

/*
#include <stdint.h>
#include <igraph.h>
#include "algorithm_cgo.h"
#include "community_cgo.h"
#include "hierarchical_community_cgo.h"
*/
import "C"

func (g *Graph) executeHierarchical(weights []float64, fn func(weightsVec *realVector, mergesMat *C.igraph_matrix_int_t, modVec *C.igraph_vector_t) (C.igraph_error_t, string)) (HierarchicalCommunity, error) {
	if g == nil {
		return HierarchicalCommunity{}, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if err := g.checkClosed(); err != nil {
		return HierarchicalCommunity{}, err
	}

	weightsVec, err := newOptionalEdgeWeights(weights, int(C.igraph_ecount(&g.graph)))
	if err != nil {
		return HierarchicalCommunity{}, err
	}
	if weightsVec != nil {
		defer weightsVec.close()
	}

	var mergesMat C.igraph_matrix_int_t
	if code := C.go_igraph_matrix_int_init(&mergesMat, 0, 0); code != 0 {
		return HierarchicalCommunity{}, igraphError("initialize merges matrix", int(code))
	}
	defer C.go_igraph_matrix_int_destroy(&mergesMat)

	var modVec C.igraph_vector_t
	if code := C.go_igraph_vector_init(&modVec, 0); code != 0 {
		return HierarchicalCommunity{}, igraphError("initialize modularity vector", int(code))
	}
	defer C.igraph_vector_destroy(&modVec)

	code, op := fn(weightsVec, &mergesMat, &modVec)
	if code != 0 {
		return HierarchicalCommunity{}, igraphError(op, int(code))
	}

	modSlice, err := (&realVector{value: modVec}).slice()
	if err != nil {
		return HierarchicalCommunity{}, err
	}

	return HierarchicalCommunity{
		Merges:     cMatrixIntToMerges(&mergesMat),
		Modularity: modSlice,
		NodeCount:  int(C.igraph_vcount(&g.graph)),
	}, nil
}

// CommunityFastGreedy finds community structure using the fast greedy modularity optimization algorithm.
//
// Inputs are borrowed; returned HierarchicalCommunity is Go-owned.
// Nil weights select an unweighted calculation.
//
// //igraph:bind igraph_community_fastgreedy
func (g *Graph) CommunityFastGreedy(weights []float64) (HierarchicalCommunity, error) {
	return g.executeHierarchical(weights, func(weightsVec *realVector, mergesMat *C.igraph_matrix_int_t, modVec *C.igraph_vector_t) (C.igraph_error_t, string) {
		return C.go_igraph_community_fastgreedy(&g.graph, edgeWeightPointer(weightsVec), mergesMat, modVec, nil), "igraph_community_fastgreedy"
	})
}

// CommunityWalktrap finds community structure based on random walks.
//
// Inputs are borrowed; returned HierarchicalCommunity is Go-owned.
// Nil weights select an unweighted calculation.
//
// //igraph:bind igraph_community_walktrap
func (g *Graph) CommunityWalktrap(weights []float64, steps int) (HierarchicalCommunity, error) {
	cSteps, err := intToIgraphInt(steps, "steps")
	if err != nil {
		return HierarchicalCommunity{}, err
	}
	return g.executeHierarchical(weights, func(weightsVec *realVector, mergesMat *C.igraph_matrix_int_t, modVec *C.igraph_vector_t) (C.igraph_error_t, string) {
		return C.go_igraph_community_walktrap(&g.graph, edgeWeightPointer(weightsVec), cSteps, mergesMat, modVec, nil), "igraph_community_walktrap"
	})
}

// CommunityEdgeBetweenness finds community structure using Girvan-Newman edge betweenness community detection.
//
// Inputs are borrowed; returned HierarchicalCommunity is Go-owned.
// Nil weights select an unweighted calculation.
//
// //igraph:bind igraph_community_edge_betweenness
func (g *Graph) CommunityEdgeBetweenness(weights []float64, directed bool) (HierarchicalCommunity, error) {
	return g.executeHierarchical(weights, func(weightsVec *realVector, mergesMat *C.igraph_matrix_int_t, modVec *C.igraph_vector_t) (C.igraph_error_t, string) {
		return C.go_igraph_community_edge_betweenness(&g.graph, nil, nil, mergesMat, nil, modVec, nil, booltoint(directed), edgeWeightPointer(weightsVec), nil), "igraph_community_edge_betweenness"
	})
}

// CommunityEBGetMerges calculates merge operations and modularities for edge betweenness community detection given a sequence of removed edges.
//
// Inputs are borrowed; returned HierarchicalCommunity is Go-owned.
// Nil weights select an unweighted calculation.
//
// //igraph:bind igraph_community_eb_get_merges
func (g *Graph) CommunityEBGetMerges(edges []int, weights []float64, directed bool) (HierarchicalCommunity, error) {
	edgesVec, err := newIntVector(edges)
	if err != nil {
		return HierarchicalCommunity{}, err
	}
	defer edgesVec.close()

	return g.executeHierarchical(weights, func(weightsVec *realVector, mergesMat *C.igraph_matrix_int_t, modVec *C.igraph_vector_t) (C.igraph_error_t, string) {
		return C.go_igraph_community_eb_get_merges(&g.graph, booltoint(directed), &edgesVec.value, edgeWeightPointer(weightsVec), mergesMat, nil, modVec, nil), "igraph_community_eb_get_merges"
	})
}

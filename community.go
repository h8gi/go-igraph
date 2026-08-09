package igraph

/*
#include <stdint.h>
#include <igraph.h>
#include "algorithm_cgo.h"
#include "community_cgo.h"
*/
import "C"
import (
	"fmt"
	"math"
	"sync"
)

var rngMutex sync.Mutex

// withRNG acquires the package RNG lock, sets the global C igraph seed if seed is non-nil,
// runs fn, and unlocks.
//
// //igraph:internal igraph_rng_seed
func withRNG(seed *uint64, fn func() error) error {
	rngMutex.Lock()
	defer rngMutex.Unlock()

	if seed != nil {
		code := C.go_igraph_rng_seed(C.uint64_t(*seed))
		if code != 0 {
			return igraphError("igraph_rng_seed", int(code))
		}
	}
	return fn()
}

// CommunityPartition represents a partition of graph vertices into community clusters.
type CommunityPartition struct {
	// Membership maps each vertex index to its assigned community ID (0..CommunityCount-1).
	Membership []int
	// CommunityCount is the total number of communities.
	CommunityCount int
	// Sizes contains the number of vertices in each community.
	Sizes []int
	// Modularity is the modularity score of the partition, or NaN if not calculated/applicable.
	Modularity float64
}

// HierarchicalCommunity represents a hierarchical community structure (dendrogram).
type HierarchicalCommunity struct {
	// Merges is an (N-1) x 2 matrix of cluster merge operations.
	Merges [][2]int
	// Modularity contains the modularity score at each step, if calculated.
	Modularity []float64
	// NodeCount is the total number of vertices in the original graph.
	NodeCount int
}

// CommunityDendrogram is an alias for HierarchicalCommunity.
type CommunityDendrogram = HierarchicalCommunity

// MembershipAt returns the CommunityPartition at a specific step in the dendrogram.
func (h HierarchicalCommunity) MembershipAt(steps int) (CommunityPartition, error) {
	if steps < 0 || steps > len(h.Merges) {
		return CommunityPartition{}, fmt.Errorf("igraph: MembershipAt: steps out of bounds: %d (max %d)", steps, len(h.Merges))
	}
	part, err := CommunityToMembership(h.Merges, h.NodeCount, steps)
	if err != nil {
		return CommunityPartition{}, err
	}
	if len(h.Modularity) > 0 && steps < len(h.Modularity) {
		part.Modularity = h.Modularity[steps]
	} else {
		part.Modularity = math.NaN()
	}
	return part, nil
}

// OptimalMembership returns the CommunityPartition corresponding to the maximum modularity score.
func (h HierarchicalCommunity) OptimalMembership() (CommunityPartition, error) {
	if len(h.Merges) == 0 && h.NodeCount == 0 {
		return CommunityPartition{
			Membership:     []int{},
			CommunityCount: 0,
			Sizes:          []int{},
			Modularity:     math.NaN(),
		}, nil
	}

	bestStep := len(h.Merges)
	maxMod := -math.MaxFloat64
	foundValidMod := false

	for step, mod := range h.Modularity {
		if !math.IsNaN(mod) && mod > maxMod {
			maxMod = mod
			bestStep = step
			foundValidMod = true
		}
	}

	if !foundValidMod {
		bestStep = len(h.Merges)
	}

	return h.MembershipAt(bestStep)
}

// ReindexMembership reindexes vertex community IDs so that they are contiguous 0-indexed integers (0..C-1).
//
// Input membership is borrowed and copied; returned slices are Go-owned.
//
// //igraph:bind igraph_reindex_membership
func ReindexMembership(membership []int) (reindexed []int, newToOld []int, count int, err error) {
	if len(membership) == 0 {
		return []int{}, []int{}, 0, nil
	}

	memVec, err := newIntVector(membership)
	if err != nil {
		return nil, nil, 0, err
	}
	defer memVec.close()

	var newToOldVec C.igraph_vector_int_t
	codeInit := C.go_igraph_vector_int_init(&newToOldVec, 0)
	if codeInit != 0 {
		return nil, nil, 0, igraphError("igraph_vector_int_init", int(codeInit))
	}
	defer C.igraph_vector_int_destroy(&newToOldVec)

	var nbClusters C.igraph_int_t
	code := C.go_igraph_reindex_membership(&memVec.value, &newToOldVec, &nbClusters)
	if code != 0 {
		return nil, nil, 0, igraphError("igraph_reindex_membership", int(code))
	}

	reindexedSlice, err := memVec.slice()
	if err != nil {
		return nil, nil, 0, err
	}
	newToOldSlice, err := intVectorSlice(&newToOldVec)
	if err != nil {
		return nil, nil, 0, err
	}
	return reindexedSlice, newToOldSlice, int(nbClusters), nil
}

// CommunityToMembership converts a merge matrix into a vertex membership slice at the specified step count.
//
// Inputs are borrowed; returned CommunityPartition is Go-owned.
//
// //igraph:bind igraph_community_to_membership
func CommunityToMembership(merges [][2]int, nodeCount int, steps int) (CommunityPartition, error) {
	if nodeCount <= 0 {
		return CommunityPartition{
			Membership:     []int{},
			CommunityCount: 0,
			Sizes:          []int{},
			Modularity:     math.NaN(),
		}, nil
	}

	if steps < 0 || steps > len(merges) {
		return CommunityPartition{}, fmt.Errorf("igraph: CommunityToMembership: steps out of bounds: %d (max %d)", steps, len(merges))
	}

	cMatrix, cleanupMatrix, err := mergesToCMatrixInt(merges)
	if err != nil {
		return CommunityPartition{}, err
	}
	defer cleanupMatrix()

	var memVec, csizeVec C.igraph_vector_int_t
	if code := C.go_igraph_vector_int_init(&memVec, 0); code != 0 {
		return CommunityPartition{}, igraphError("igraph_vector_int_init", int(code))
	}
	defer C.igraph_vector_int_destroy(&memVec)

	if code := C.go_igraph_vector_int_init(&csizeVec, 0); code != 0 {
		return CommunityPartition{}, igraphError("igraph_vector_int_init", int(code))
	}
	defer C.igraph_vector_int_destroy(&csizeVec)

	code := C.go_igraph_community_to_membership(
		&cMatrix,
		C.igraph_int_t(nodeCount),
		C.igraph_int_t(steps),
		&memVec,
		&csizeVec,
	)
	if code != 0 {
		return CommunityPartition{}, igraphError("igraph_community_to_membership", int(code))
	}

	memSlice, err := intVectorSlice(&memVec)
	if err != nil {
		return CommunityPartition{}, err
	}
	csizeSlice, err := intVectorSlice(&csizeVec)
	if err != nil {
		return CommunityPartition{}, err
	}

	return CommunityPartition{
		Membership:     memSlice,
		CommunityCount: len(csizeSlice),
		Sizes:          csizeSlice,
		Modularity:     math.NaN(),
	}, nil
}

// LeadingEigenvectorCommunityToMembership converts a leading eigenvector merge matrix into a vertex membership slice at the specified step count.
//
// Inputs are borrowed; returned CommunityPartition is Go-owned.
//
// //igraph:bind igraph_le_community_to_membership
func LeadingEigenvectorCommunityToMembership(merges [][2]int, steps int) (CommunityPartition, error) {
	if len(merges) == 0 && steps == 0 {
		return CommunityPartition{
			Membership:     []int{},
			CommunityCount: 0,
			Sizes:          []int{},
			Modularity:     math.NaN(),
		}, nil
	}

	if steps < 0 || steps > len(merges) {
		return CommunityPartition{}, fmt.Errorf("igraph: LeadingEigenvectorCommunityToMembership: steps out of bounds: %d (max %d)", steps, len(merges))
	}

	cMatrix, cleanupMatrix, err := mergesToCMatrixIntCols(merges, 3)
	if err != nil {
		return CommunityPartition{}, err
	}
	defer cleanupMatrix()

	var memVec, csizeVec C.igraph_vector_int_t
	if code := C.go_igraph_vector_int_init(&memVec, 0); code != 0 {
		return CommunityPartition{}, igraphError("igraph_vector_int_init", int(code))
	}
	defer C.igraph_vector_int_destroy(&memVec)

	if code := C.go_igraph_vector_int_init(&csizeVec, 0); code != 0 {
		return CommunityPartition{}, igraphError("igraph_vector_int_init", int(code))
	}
	defer C.igraph_vector_int_destroy(&csizeVec)

	code := C.go_igraph_le_community_to_membership(
		&cMatrix,
		C.igraph_int_t(steps),
		&memVec,
		&csizeVec,
	)
	if code != 0 {
		return CommunityPartition{}, igraphError("igraph_le_community_to_membership", int(code))
	}

	memSlice, err := intVectorSlice(&memVec)
	if err != nil {
		return CommunityPartition{}, err
	}
	csizeSlice, err := intVectorSlice(&csizeVec)
	if err != nil {
		return CommunityPartition{}, err
	}

	return CommunityPartition{
		Membership:     memSlice,
		CommunityCount: len(csizeSlice),
		Sizes:          csizeSlice,
		Modularity:     math.NaN(),
	}, nil
}

// mergesToCMatrixInt allocates an igraph_matrix_int_t from a Go merges slice [][2]int.
func mergesToCMatrixInt(merges [][2]int) (C.igraph_matrix_int_t, func(), error) {
	return mergesToCMatrixIntCols(merges, 2)
}

func mergesToCMatrixIntCols(merges [][2]int, numCols int) (C.igraph_matrix_int_t, func(), error) {
	var m C.igraph_matrix_int_t
	rows := len(merges)
	code := C.go_igraph_matrix_int_init(&m, C.igraph_int_t(rows), C.igraph_int_t(numCols))
	if code != 0 {
		return m, func() {}, igraphError("igraph_matrix_int_init", int(code))
	}
	cleanup := func() {
		C.go_igraph_matrix_int_destroy(&m)
	}

	for r := 0; r < rows; r++ {
		for c := 0; c < numCols; c++ {
			val := 0
			if c < 2 {
				val = merges[r][c]
			}
			C.go_igraph_matrix_int_set(&m, C.igraph_int_t(r), C.igraph_int_t(c), C.igraph_int_t(val))
		}
	}
	return m, cleanup, nil
}

// cMatrixIntToMerges converts an igraph_matrix_int_t to a Go merges slice [][2]int.
func cMatrixIntToMerges(m *C.igraph_matrix_int_t) [][2]int {
	if m == nil {
		return [][2]int{}
	}
	rows := int(C.go_igraph_matrix_int_nrow(m))
	if rows == 0 {
		return [][2]int{}
	}
	res := make([][2]int, rows)
	for r := 0; r < rows; r++ {
		res[r][0] = int(C.go_igraph_matrix_int_get(m, C.igraph_int_t(r), 0))
		res[r][1] = int(C.go_igraph_matrix_int_get(m, C.igraph_int_t(r), 1))
	}
	return res
}

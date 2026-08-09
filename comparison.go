package igraph

/*
#cgo pkg-config: igraph
#include <igraph.h>
#include "comparison_cgo.h"
*/
import "C"

import (
	"fmt"
)

// CommunityComparisonMethod specifies the method used to compare two community partitions.
type CommunityComparisonMethod int

const (
	// CommunityCompareVI calculates the Variation of Information metric (Meila 2003).
	CommunityCompareVI CommunityComparisonMethod = 0
	// CommunityCompareNMI calculates the Normalized Mutual Information metric (Danon et al. 2005).
	CommunityCompareNMI CommunityComparisonMethod = 1
	// CommunityCompareARI calculates the Split-Join distance or Adjusted Rand Index metric.
	CommunityCompareARI CommunityComparisonMethod = 2
	// CommunityCompareRand calculates the Rand index (Rand 1971).
	CommunityCompareRand CommunityComparisonMethod = 3
	// CommunityCompareAdjustedRand calculates the Adjusted Rand Index (Hubert and Arabie 1985).
	CommunityCompareAdjustedRand CommunityComparisonMethod = 4
)

func (m CommunityComparisonMethod) cValue() (C.igraph_community_comparison_t, error) {
	switch m {
	case CommunityCompareVI:
		return C.IGRAPH_COMMCMP_VI, nil
	case CommunityCompareNMI:
		return C.IGRAPH_COMMCMP_NMI, nil
	case CommunityCompareARI:
		return C.IGRAPH_COMMCMP_SPLIT_JOIN, nil
	case CommunityCompareRand:
		return C.IGRAPH_COMMCMP_RAND, nil
	case CommunityCompareAdjustedRand:
		return C.IGRAPH_COMMCMP_ADJUSTED_RAND, nil
	default:
		return 0, fmt.Errorf("igraph: CompareCommunities: invalid method: %d", m)
	}
}

// SplitJoinDistanceResult contains the asymmetric projection distances between two community structures.
type SplitJoinDistanceResult struct {
	// Distance1To2 is the projection distance of the first community partition from the second.
	Distance1To2 int
	// Distance2To1 is the projection distance of the second community partition from the first.
	Distance2To1 int
}

func validateCommunityInputs(comm1, comm2 []int, opName string) error {
	if len(comm1) == 0 {
		return fmt.Errorf("igraph: %s: comm1 must not be empty", opName)
	}
	if len(comm2) == 0 {
		return fmt.Errorf("igraph: %s: comm2 must not be empty", opName)
	}
	if len(comm1) != len(comm2) {
		return fmt.Errorf("igraph: %s: slice lengths do not match: len(comm1)=%d, len(comm2)=%d", opName, len(comm1), len(comm2))
	}
	for i, v := range comm1 {
		if v < 0 {
			return fmt.Errorf("igraph: %s: negative community ID in comm1 at index %d: %d", opName, i, v)
		}
	}
	for i, v := range comm2 {
		if v < 0 {
			return fmt.Errorf("igraph: %s: negative community ID in comm2 at index %d: %d", opName, i, v)
		}
	}
	return nil
}

// CompareCommunities compares two community partitions of the same graph.
//
// Inputs comm1 and comm2 are borrowed and copied; returned float64 score is Go-owned.
//
// //igraph:bind igraph_compare_communities
func CompareCommunities(comm1, comm2 []int, method CommunityComparisonMethod) (float64, error) {
	if err := validateCommunityInputs(comm1, comm2, "CompareCommunities"); err != nil {
		return 0, err
	}
	cMethod, err := method.cValue()
	if err != nil {
		return 0, err
	}

	vec1, err := newIntVector(comm1)
	if err != nil {
		return 0, err
	}
	defer vec1.close()

	vec2, err := newIntVector(comm2)
	if err != nil {
		return 0, err
	}
	defer vec2.close()

	var result C.igraph_real_t
	code := C.go_igraph_compare_communities(&vec1.value, &vec2.value, &result, cMethod)
	if code != 0 {
		return 0, igraphError("igraph_compare_communities", int(code))
	}

	return float64(result), nil
}

// SplitJoinDistance calculates the split-join distance between two community partitions.
//
// Inputs comm1 and comm2 are borrowed and copied; returned SplitJoinDistanceResult is Go-owned.
//
// //igraph:bind igraph_split_join_distance
func SplitJoinDistance(comm1, comm2 []int) (SplitJoinDistanceResult, error) {
	if err := validateCommunityInputs(comm1, comm2, "SplitJoinDistance"); err != nil {
		return SplitJoinDistanceResult{}, err
	}

	vec1, err := newIntVector(comm1)
	if err != nil {
		return SplitJoinDistanceResult{}, err
	}
	defer vec1.close()

	vec2, err := newIntVector(comm2)
	if err != nil {
		return SplitJoinDistanceResult{}, err
	}
	defer vec2.close()

	var d12, d21 C.igraph_int_t
	code := C.go_igraph_split_join_distance(&vec1.value, &vec2.value, &d12, &d21)
	if code != 0 {
		return SplitJoinDistanceResult{}, igraphError("igraph_split_join_distance", int(code))
	}

	dist12, err := igraphIntToInt(d12, "distance 1 to 2")
	if err != nil {
		return SplitJoinDistanceResult{}, err
	}
	dist21, err := igraphIntToInt(d21, "distance 2 to 1")
	if err != nil {
		return SplitJoinDistanceResult{}, err
	}

	return SplitJoinDistanceResult{
		Distance1To2: dist12,
		Distance2To1: dist21,
	}, nil
}

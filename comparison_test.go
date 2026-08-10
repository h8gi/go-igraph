package igraph_test

import (
	"math"
	"testing"

	"github.com/h8gi/go-igraph"
)

func TestCompareCommunities_Normal(t *testing.T) {
	comm1 := []int{0, 0, 1, 1}
	comm2 := []int{0, 0, 1, 1}

	// Identical partitions
	vi, err := igraph.CompareCommunities(comm1, comm2, igraph.CommunityCompareVI)
	if err != nil {
		t.Fatalf("unexpected error for CompareCommunities VI: %v", err)
	}
	if math.Abs(vi) > 1e-9 {
		t.Errorf("expected VI ~ 0 for identical partitions, got %f", vi)
	}

	nmi, err := igraph.CompareCommunities(comm1, comm2, igraph.CommunityCompareNMI)
	if err != nil {
		t.Fatalf("unexpected error for CompareCommunities NMI: %v", err)
	}
	if math.Abs(nmi-1.0) > 1e-9 {
		t.Errorf("expected NMI ~ 1.0 for identical partitions, got %f", nmi)
	}

	ari, err := igraph.CompareCommunities(comm1, comm2, igraph.CommunityCompareARI)
	if err != nil {
		t.Fatalf("unexpected error for CompareCommunities ARI: %v", err)
	}
	if math.Abs(ari) > 1e-9 {
		t.Errorf("expected ARI (split-join) ~ 0 for identical partitions, got %f", ari)
	}

	randIdx, err := igraph.CompareCommunities(comm1, comm2, igraph.CommunityCompareRand)
	if err != nil {
		t.Fatalf("unexpected error for CompareCommunities Rand: %v", err)
	}
	if math.Abs(randIdx-1.0) > 1e-9 {
		t.Errorf("expected Rand ~ 1.0 for identical partitions, got %f", randIdx)
	}

	adjRand, err := igraph.CompareCommunities(comm1, comm2, igraph.CommunityCompareAdjustedRand)
	if err != nil {
		t.Fatalf("unexpected error for CompareCommunities AdjustedRand: %v", err)
	}
	if math.Abs(adjRand-1.0) > 1e-9 {
		t.Errorf("expected AdjustedRand ~ 1.0 for identical partitions, got %f", adjRand)
	}

	// Relabeled identical partitions
	comm3 := []int{1, 1, 0, 0}
	nmiRelabeled, err := igraph.CompareCommunities(comm1, comm3, igraph.CommunityCompareNMI)
	if err != nil {
		t.Fatalf("unexpected error for CompareCommunities NMI relabeled: %v", err)
	}
	if math.Abs(nmiRelabeled-1.0) > 1e-9 {
		t.Errorf("expected NMI ~ 1.0 for relabeled partitions, got %f", nmiRelabeled)
	}

	// Different partitions
	comm4 := []int{0, 1, 0, 1}
	randDiff, err := igraph.CompareCommunities(comm1, comm4, igraph.CommunityCompareRand)
	if err != nil {
		t.Fatalf("unexpected error for CompareCommunities Rand diff: %v", err)
	}
	if randDiff >= 1.0 || randDiff < 0.0 {
		t.Errorf("expected Rand between 0 and 1 for different partitions, got %f", randDiff)
	}
}

func TestCompareCommunities_InvalidInputs(t *testing.T) {
	valid := []int{0, 0, 1, 1}

	t.Run("empty comm1", func(t *testing.T) {
		_, err := igraph.CompareCommunities([]int{}, valid, igraph.CommunityCompareVI)
		if err == nil {
			t.Error("expected error for empty comm1, got nil")
		}
	})

	t.Run("empty comm2", func(t *testing.T) {
		_, err := igraph.CompareCommunities(valid, nil, igraph.CommunityCompareVI)
		if err == nil {
			t.Error("expected error for nil comm2, got nil")
		}
	})

	t.Run("mismatched lengths", func(t *testing.T) {
		_, err := igraph.CompareCommunities(valid, []int{0, 1}, igraph.CommunityCompareVI)
		if err == nil {
			t.Error("expected error for mismatched lengths, got nil")
		}
	})

	t.Run("negative community ID in comm1", func(t *testing.T) {
		_, err := igraph.CompareCommunities([]int{-1, 0, 1, 1}, valid, igraph.CommunityCompareVI)
		if err == nil {
			t.Error("expected error for negative community ID in comm1, got nil")
		}
	})

	t.Run("negative community ID in comm2", func(t *testing.T) {
		_, err := igraph.CompareCommunities(valid, []int{0, 0, -2, 1}, igraph.CommunityCompareVI)
		if err == nil {
			t.Error("expected error for negative community ID in comm2, got nil")
		}
	})

	t.Run("invalid method lower bound", func(t *testing.T) {
		_, err := igraph.CompareCommunities(valid, valid, igraph.CommunityComparisonMethod(-1))
		if err == nil {
			t.Error("expected error for method -1, got nil")
		}
	})

	t.Run("invalid method upper bound", func(t *testing.T) {
		_, err := igraph.CompareCommunities(valid, valid, igraph.CommunityComparisonMethod(5))
		if err == nil {
			t.Error("expected error for method 5, got nil")
		}
	})
}

func TestSplitJoinDistance_Normal(t *testing.T) {
	t.Run("identical partitions", func(t *testing.T) {
		comm1 := []int{0, 0, 1, 1}
		comm2 := []int{0, 0, 1, 1}
		res, err := igraph.SplitJoinDistance(comm1, comm2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Distance1To2 != 0 || res.Distance2To1 != 0 {
			t.Errorf("expected distance 0, 0 for identical partitions, got %d, %d", res.Distance1To2, res.Distance2To1)
		}
	})

	t.Run("relabeled partitions", func(t *testing.T) {
		comm1 := []int{0, 0, 1, 1}
		comm2 := []int{1, 1, 0, 0}
		res, err := igraph.SplitJoinDistance(comm1, comm2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Distance1To2 != 0 || res.Distance2To1 != 0 {
			t.Errorf("expected distance 0, 0 for relabeled partitions, got %d, %d", res.Distance1To2, res.Distance2To1)
		}
	})

	t.Run("different partitions", func(t *testing.T) {
		comm1 := []int{0, 0, 0, 0}
		comm2 := []int{0, 1, 2, 3}
		res, err := igraph.SplitJoinDistance(comm1, comm2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Distance1To2 < 0 || res.Distance2To1 < 0 {
			t.Errorf("expected non-negative distances, got %d, %d", res.Distance1To2, res.Distance2To1)
		}
	})
}

func TestSplitJoinDistance_InvalidInputs(t *testing.T) {
	valid := []int{0, 0, 1, 1}

	t.Run("empty comm1", func(t *testing.T) {
		_, err := igraph.SplitJoinDistance([]int{}, valid)
		if err == nil {
			t.Error("expected error for empty comm1, got nil")
		}
	})

	t.Run("empty comm2", func(t *testing.T) {
		_, err := igraph.SplitJoinDistance(valid, nil)
		if err == nil {
			t.Error("expected error for nil comm2, got nil")
		}
	})

	t.Run("mismatched lengths", func(t *testing.T) {
		_, err := igraph.SplitJoinDistance(valid, []int{0, 1})
		if err == nil {
			t.Error("expected error for mismatched lengths, got nil")
		}
	})

	t.Run("negative community ID in comm1", func(t *testing.T) {
		_, err := igraph.SplitJoinDistance([]int{-1, 0, 1, 1}, valid)
		if err == nil {
			t.Error("expected error for negative community ID in comm1, got nil")
		}
	})

	t.Run("negative community ID in comm2", func(t *testing.T) {
		_, err := igraph.SplitJoinDistance(valid, []int{0, 0, -2, 1})
		if err == nil {
			t.Error("expected error for negative community ID in comm2, got nil")
		}
	})

	t.Run("invalid method", func(t *testing.T) {
		comm1 := []int{0, 1}
		comm2 := []int{0, 1}
		if _, err := igraph.CompareCommunities(comm1, comm2, igraph.CommunityComparisonMethod(99)); err == nil {
			t.Error("expected error for invalid CommunityComparisonMethod")
		}
	})
}

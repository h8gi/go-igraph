package igraph

import (
	"testing"
)

func TestComparisonInternalInvalidMethod(t *testing.T) {
	if _, err := CommunityComparisonMethod(99).cValue(); err == nil {
		t.Error("expected error for invalid CommunityComparisonMethod")
	}
	if _, err := CompareCommunities(nil, []int{0}, CommunityCompareVI); err == nil {
		t.Error("expected error for empty comm1 in CompareCommunities")
	}
	if _, err := CompareCommunities([]int{0}, nil, CommunityCompareVI); err == nil {
		t.Error("expected error for empty comm2 in CompareCommunities")
	}
}

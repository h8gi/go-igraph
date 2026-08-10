package igraph

import (
	"math"
	"testing"
)

func TestCommunityToMembershipNodeCountZeroOrNegative(t *testing.T) {
	part0, err := CommunityToMembership(nil, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error for 0 node count: %v", err)
	}
	if len(part0.Membership) != 0 || part0.CommunityCount != 0 || !math.IsNaN(part0.Modularity) {
		t.Errorf("unexpected partition for 0 node count: %+v", part0)
	}

	partNeg, err := CommunityToMembership(nil, -5, 0)
	if err != nil {
		t.Fatalf("unexpected error for negative node count: %v", err)
	}
	if len(partNeg.Membership) != 0 || partNeg.CommunityCount != 0 || !math.IsNaN(partNeg.Modularity) {
		t.Errorf("unexpected partition for negative node count: %+v", partNeg)
	}
}

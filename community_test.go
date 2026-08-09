package igraph_test

import (
	"math"
	"reflect"
	"sync"
	"testing"

	"github.com/h8gi/go-igraph"
)

func TestReindexMembership(t *testing.T) {
	t.Run("empty membership", func(t *testing.T) {
		reindexed, newToOld, count, err := igraph.ReindexMembership([]int{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(reindexed) != 0 || len(newToOld) != 0 || count != 0 {
			t.Errorf("expected empty results, got reindexed=%v, newToOld=%v, count=%d", reindexed, newToOld, count)
		}
	})

	t.Run("already indexed 0..C-1", func(t *testing.T) {
		mem := []int{0, 1, 0, 1, 2}
		reindexed, newToOld, count, err := igraph.ReindexMembership(mem)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 3 {
			t.Errorf("expected 3 clusters, got %d", count)
		}
		if len(reindexed) != len(mem) {
			t.Fatalf("expected len %d, got %d", len(mem), len(reindexed))
		}
		if !reflect.DeepEqual(reindexed, []int{0, 1, 0, 1, 2}) {
			t.Errorf("unexpected reindexed: %v", reindexed)
		}
		if !reflect.DeepEqual(newToOld, []int{0, 1, 2}) {
			t.Errorf("unexpected newToOld: %v", newToOld)
		}
	})

	t.Run("non-contiguous IDs", func(t *testing.T) {
		mem := []int{10, 42, 10, 5}
		reindexed, newToOld, count, err := igraph.ReindexMembership(mem)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 3 {
			t.Errorf("expected 3 clusters, got %d", count)
		}
		// Verify every element is in 0..count-1
		for _, id := range reindexed {
			if id < 0 || id >= count {
				t.Errorf("id %d out of bounds 0..%d", id, count-1)
			}
		}
		if len(newToOld) != count {
			t.Errorf("expected newToOld len %d, got %d", count, len(newToOld))
		}
	})
}

func TestCommunityToMembership(t *testing.T) {
	// A small 4-node tree merge matrix:
	// step 0: merge node 0 and 1 -> cluster 4
	// step 1: merge node 2 and 3 -> cluster 5
	// step 2: merge cluster 4 and 5 -> cluster 6
	merges := [][2]int{
		{0, 1},
		{2, 3},
		{4, 5},
	}
	nodeCount := 4

	t.Run("step 0", func(t *testing.T) {
		part, err := igraph.CommunityToMembership(merges, nodeCount, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if part.CommunityCount != 4 {
			t.Errorf("expected 4 communities at step 0, got %d", part.CommunityCount)
		}
		if len(part.Membership) != 4 {
			t.Errorf("expected membership len 4, got %d", len(part.Membership))
		}
	})

	t.Run("step 1", func(t *testing.T) {
		part, err := igraph.CommunityToMembership(merges, nodeCount, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if part.CommunityCount != 3 {
			t.Errorf("expected 3 communities at step 1, got %d", part.CommunityCount)
		}
		if part.Membership[0] != part.Membership[1] {
			t.Errorf("nodes 0 and 1 should be in same community")
		}
	})

	t.Run("step 3 (all merged)", func(t *testing.T) {
		part, err := igraph.CommunityToMembership(merges, nodeCount, 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if part.CommunityCount != 1 {
			t.Errorf("expected 1 community at final step, got %d", part.CommunityCount)
		}
	})

	t.Run("out of bounds steps", func(t *testing.T) {
		_, err := igraph.CommunityToMembership(merges, nodeCount, -1)
		if err == nil {
			t.Errorf("expected error for negative steps")
		}
		_, err = igraph.CommunityToMembership(merges, nodeCount, 4)
		if err == nil {
			t.Errorf("expected error for steps > len(merges)")
		}
	})

	t.Run("empty graph (nodeCount 0)", func(t *testing.T) {
		part, err := igraph.CommunityToMembership([][2]int{}, 0, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if part.CommunityCount != 0 || len(part.Membership) != 0 {
			t.Errorf("expected empty partition, got %+v", part)
		}
	})
}

func TestLeadingEigenvectorCommunityToMembership(t *testing.T) {
	t.Run("empty merges", func(t *testing.T) {
		part, err := igraph.LeadingEigenvectorCommunityToMembership([][2]int{}, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if part.CommunityCount != 0 {
			t.Errorf("expected empty partition")
		}
	})

	t.Run("invalid merge matrix returns error", func(t *testing.T) {
		invalidMerges := [][2]int{
			{0, 1},
		}
		_, err := igraph.LeadingEigenvectorCommunityToMembership(invalidMerges, 1)
		if err == nil {
			t.Errorf("expected error for invalid leading eigenvector merge matrix")
		}
	})

	t.Run("out of bounds steps", func(t *testing.T) {
		merges := [][2]int{{0, 1}}
		_, err := igraph.LeadingEigenvectorCommunityToMembership(merges, 10)
		if err == nil {
			t.Errorf("expected error for steps > len(merges)")
		}
	})
}

func TestHierarchicalCommunity(t *testing.T) {
	merges := [][2]int{
		{0, 1},
		{2, 3},
		{4, 5},
	}
	mods := []float64{0.1, 0.45, 0.3}
	hc := igraph.HierarchicalCommunity{
		Merges:     merges,
		Modularity: mods,
		NodeCount:  4,
	}

	t.Run("MembershipAt valid step", func(t *testing.T) {
		part, err := hc.MembershipAt(1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if part.CommunityCount != 3 {
			t.Errorf("expected 3 communities, got %d", part.CommunityCount)
		}
		if part.Modularity != 0.45 {
			t.Errorf("expected modularity 0.45, got %f", part.Modularity)
		}
	})

	t.Run("OptimalMembership", func(t *testing.T) {
		part, err := hc.OptimalMembership()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Max modularity is at step 1 (0.45)
		if part.Modularity != 0.45 {
			t.Errorf("expected optimal modularity 0.45, got %f", part.Modularity)
		}
		if part.CommunityCount != 3 {
			t.Errorf("expected 3 communities, got %d", part.CommunityCount)
		}
	})

	t.Run("OptimalMembership with no modularity", func(t *testing.T) {
		hcNoMod := igraph.HierarchicalCommunity{
			Merges:     merges,
			Modularity: nil,
			NodeCount:  4,
		}
		part, err := hcNoMod.OptimalMembership()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if part.CommunityCount != 1 {
			t.Errorf("expected 1 community at max step when no modularity, got %d", part.CommunityCount)
		}
		if !math.IsNaN(part.Modularity) {
			t.Errorf("expected NaN modularity")
		}
	})

	t.Run("Empty HierarchicalCommunity", func(t *testing.T) {
		emptyHC := igraph.HierarchicalCommunity{}
		part, err := emptyHC.OptimalMembership()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if part.CommunityCount != 0 || len(part.Membership) != 0 {
			t.Errorf("expected empty partition, got %+v", part)
		}
	})
}

func TestThreadSafeRNGContract(t *testing.T) {
	// Concurrent calls setting seeds must not panic or race
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(seed uint64) {
			defer wg.Done()
			mem := []int{0, 1, 2, 1, 0}
			_, _, _, err := igraph.ReindexMembership(mem)
			if err != nil {
				t.Errorf("error in concurrent run: %v", err)
			}
		}(uint64(i * 100))
	}
	wg.Wait()
}

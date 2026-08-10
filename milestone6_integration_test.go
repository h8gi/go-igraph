package igraph

import (
	"math"
	"testing"
)

func TestMilestone6CommunityIntegrationPipeline(t *testing.T) {
	// Build a benchmark graph: two cliques of 4 vertices connected by a bridge edge (3-4)
	edges := []Edge{
		{From: 0, To: 1}, {From: 0, To: 2}, {From: 0, To: 3}, {From: 1, To: 2}, {From: 1, To: 3}, {From: 2, To: 3},
		{From: 4, To: 5}, {From: 4, To: 6}, {From: 4, To: 7}, {From: 5, To: 6}, {From: 5, To: 7}, {From: 6, To: 7},
		{From: 3, To: 4},
	}
	g, err := NewGraphFromEdges(8, edges, false)
	if err != nil {
		t.Fatalf("failed to create benchmark graph: %v", err)
	}

	// 1. Modularity calculation and Modularity Matrix
	weights := make([]float64, len(edges))
	for i := range weights {
		weights[i] = 1.0
	}
	trueMembership := []int{0, 0, 0, 0, 1, 1, 1, 1}

	modScore, err := g.Modularity(trueMembership, weights, 1.0)
	if err != nil {
		t.Fatalf("Modularity failed: %v", err)
	}
	if math.IsNaN(modScore) || modScore <= 0 {
		t.Errorf("expected positive modularity for true partition, got %f", modScore)
	}

	modMat, err := g.ModularityMatrix(weights, 1.0)
	if err != nil {
		t.Fatalf("ModularityMatrix failed: %v", err)
	}
	rows, cols := modMat.Dims()
	if rows != 8 || cols != 8 {
		t.Errorf("expected 8x8 modularity matrix, got %dx%d", rows, cols)
	}

	// 2. Flat Community Detection (Multilevel, Leiden, Infomap)
	seed := uint64(42)
	multilevelPart, err := g.CommunityMultilevel(MultilevelOptions{Seed: &seed})
	if err != nil {
		t.Fatalf("CommunityMultilevel failed: %v", err)
	}
	if multilevelPart.CommunityCount < 2 {
		t.Errorf("expected >= 2 communities, got %d", multilevelPart.CommunityCount)
	}

	leidenPart, err := g.CommunityLeiden(LeidenOptions{Seed: &seed})
	if err != nil {
		t.Fatalf("CommunityLeiden failed: %v", err)
	}
	if leidenPart.CommunityCount < 2 {
		t.Errorf("expected >= 2 communities, got %d", leidenPart.CommunityCount)
	}

	infomapPart, err := g.CommunityInfomap(InfomapOptions{NTrials: 5})
	if err != nil {
		t.Fatalf("CommunityInfomap failed: %v", err)
	}
	if infomapPart.CommunityCount < 2 {
		t.Errorf("expected >= 2 communities, got %d", infomapPart.CommunityCount)
	}

	// 3. Hierarchical Community Detection (Walktrap & FastGreedy)
	walktrapHier, err := g.CommunityWalktrap(nil, 4)
	if err != nil {
		t.Fatalf("CommunityWalktrap failed: %v", err)
	}
	if len(walktrapHier.Merges) == 0 {
		t.Error("expected non-empty merges in Walktrap hierarchical community")
	}

	walktrapPartAt, err := walktrapHier.MembershipAt(len(walktrapHier.Merges) - 2)
	if err != nil {
		t.Fatalf("MembershipAt failed: %v", err)
	}
	if walktrapPartAt.CommunityCount < 2 {
		t.Errorf("expected >= 2 communities from dendrogram cut, got %d", walktrapPartAt.CommunityCount)
	}

	walktrapOptimalPart, err := walktrapHier.OptimalMembership()
	if err != nil {
		t.Fatalf("OptimalMembership failed: %v", err)
	}
	if walktrapOptimalPart.CommunityCount < 2 {
		t.Errorf("expected >= 2 communities from optimal membership, got %d", walktrapOptimalPart.CommunityCount)
	}

	// 4. Spectral & Exact Community Detection (Leading Eigenvector, Spinglass, Optimal Modularity)
	lePart, err := g.CommunityLeadingEigenvector(LeadingEigenvectorOptions{})
	if err != nil {
		t.Fatalf("CommunityLeadingEigenvector failed: %v", err)
	}
	if lePart.CommunityCount < 2 {
		t.Errorf("expected >= 2 communities from Leading Eigenvector, got %d", lePart.CommunityCount)
	}

	spinglassPart, err := g.CommunitySpinglass(SpinglassOptions{Seed: &seed})
	if err != nil {
		t.Fatalf("CommunitySpinglass failed: %v", err)
	}
	if spinglassPart.CommunityCount < 2 {
		t.Errorf("expected >= 2 communities from Spinglass, got %d", spinglassPart.CommunityCount)
	}

	singleRes, err := g.CommunitySpinglassSingle(SpinglassSingleOptions{Vertex: 0, Seed: &seed})
	if err != nil {
		t.Fatalf("CommunitySpinglassSingle failed: %v", err)
	}
	if len(singleRes.Community) == 0 {
		t.Error("expected non-empty single vertex community result")
	}

	optModPart, err := g.CommunityOptimalModularity(nil)
	if err != nil {
		t.Fatalf("CommunityOptimalModularity failed: %v", err)
	}
	if optModPart.CommunityCount < 2 {
		t.Errorf("expected >= 2 communities from Optimal Modularity, got %d", optModPart.CommunityCount)
	}

	// 5. Partition Comparison Metrics (CompareCommunities)
	viDist, err := CompareCommunities(multilevelPart.Membership, leidenPart.Membership, CommunityCompareVI)
	if err != nil {
		t.Fatalf("CompareCommunities VI failed: %v", err)
	}
	if viDist < 0 || math.IsNaN(viDist) {
		t.Errorf("invalid VI distance: %f", viDist)
	}

	nmiSim, err := CompareCommunities(multilevelPart.Membership, lePart.Membership, CommunityCompareNMI)
	if err != nil {
		t.Fatalf("CompareCommunities NMI failed: %v", err)
	}
	if nmiSim < 0 || nmiSim > 1.0 || math.IsNaN(nmiSim) {
		t.Errorf("invalid NMI similarity: %f", nmiSim)
	}

	splitJoinDist, err := SplitJoinDistance(multilevelPart.Membership, optModPart.Membership)
	if err != nil {
		t.Fatalf("SplitJoinDistance failed: %v", err)
	}
	if splitJoinDist.Distance1To2 < 0 || splitJoinDist.Distance2To1 < 0 {
		t.Errorf("invalid SplitJoinDistance: %+v", splitJoinDist)
	}

	// 6. Verification of Go ownership after Graph Close
	if err := g.Close(); err != nil {
		t.Fatalf("Graph.Close failed: %v", err)
	}

	// All returned structs must remain fully valid and accessible after graph closure
	if len(multilevelPart.Membership) != 8 || multilevelPart.CommunityCount == 0 {
		t.Error("multilevelPart modified or corrupted after graph closure")
	}
	if len(walktrapHier.Merges) == 0 || walktrapHier.NodeCount != 8 {
		t.Error("walktrapHier modified or corrupted after graph closure")
	}
	if len(singleRes.Community) == 0 {
		t.Error("singleRes modified or corrupted after graph closure")
	}
	if len(optModPart.Membership) != 8 {
		t.Error("optModPart modified or corrupted after graph closure")
	}
}

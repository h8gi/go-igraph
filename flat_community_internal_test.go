package igraph

import (
	"testing"
)

func TestFlatCommunityInternalHelpers(t *testing.T) {
	t.Run("partitionFromMembership empty", func(t *testing.T) {
		part, err := partitionFromMembership(nil, 0.5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(part.Membership) != 0 || part.CommunityCount != 0 || part.Modularity != 0.5 {
			t.Errorf("unexpected empty partition: %+v", part)
		}
	})

	t.Run("partitionFromMembership non-empty reindexing", func(t *testing.T) {
		mem := []int{10, 10, 5, 5, 20}
		part, err := partitionFromMembership(mem, 0.42)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if part.CommunityCount != 3 {
			t.Errorf("expected 3 communities, got %d", part.CommunityCount)
		}
		if len(part.Sizes) != 3 {
			t.Errorf("expected 3 sizes, got %v", part.Sizes)
		}
		if part.Modularity != 0.42 {
			t.Errorf("expected modularity 0.42, got %f", part.Modularity)
		}
	})

	t.Run("newOptionalVertexWeights empty and length mismatch", func(t *testing.T) {
		vec, err := newOptionalVertexWeights(nil, 5)
		if err != nil || vec != nil {
			t.Errorf("expected nil vector for empty slice, got %v, %v", vec, err)
		}

		vecEmpty, err := newOptionalVertexWeights([]float64{}, 5)
		if err != nil || vecEmpty != nil {
			t.Errorf("expected nil vector for empty slice, got %v, %v", vecEmpty, err)
		}

		vecValid, err := newOptionalVertexWeights([]float64{1, 2, 3}, 3)
		if err != nil || vecValid == nil {
			t.Errorf("expected valid vector, got %v, %v", vecValid, err)
		} else {
			vecValid.close()
		}

		_, errMismatch := newOptionalVertexWeights([]float64{1, 2}, 3)
		if errMismatch == nil {
			t.Errorf("expected error for vertex weights length mismatch")
		}
	})

	t.Run("cMatrixIntToMerges nil and zero rows", func(t *testing.T) {
		resNil := cMatrixIntToMerges(nil)
		if len(resNil) != 0 {
			t.Errorf("expected empty merges for nil matrix, got %v", resNil)
		}
	})
}

func TestFlatCommunityInternalFull(t *testing.T) {
	edges := []Edge{
		{From: 0, To: 1}, {From: 0, To: 2}, {From: 0, To: 3}, {From: 1, To: 2},
		{From: 4, To: 5}, {From: 4, To: 6}, {From: 4, To: 7}, {From: 5, To: 6},
		{From: 3, To: 4},
	}
	g, err := NewGraphFromEdges(8, edges, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	seed := uint64(999)
	eWeights := []float64{1, 2, 1, 2, 1, 2, 1, 2, 1}

	// Multilevel
	partM, err := g.CommunityMultilevel(MultilevelOptions{
		Weights:    eWeights,
		Resolution: 0.8,
		Seed:       &seed,
	})
	if err != nil || len(partM.Membership) != 8 {
		t.Fatalf("CommunityMultilevel failed: %v, %v", partM, err)
	}

	// Leiden
	partL, err := g.CommunityLeiden(LeidenOptions{
		EdgeWeights: eWeights,
		Resolution:  0.8,
		Beta:        0.02,
		NIterations: 3,
		Seed:        &seed,
	})
	if err != nil || len(partL.Membership) != 8 {
		t.Fatalf("CommunityLeiden failed: %v, %v", partL, err)
	}

	// Label Propagation
	initial := []int{0, 0, 0, 0, 1, 1, 1, 1}
	fixed := []bool{true, false, true, false, true, false, true, false}
	partLPA, err := g.CommunityLabelPropagation(LabelPropagationOptions{
		Mode:              DirectionAll,
		Weights:           eWeights,
		InitialMembership: initial,
		Fixed:             fixed,
		Seed:              &seed,
	})
	if err != nil || len(partLPA.Membership) != 8 {
		t.Fatalf("CommunityLabelPropagation failed: %v, %v", partLPA, err)
	}

	// Infomap
	vWeights := []float64{1, 1, 1, 1, 1, 1, 1, 1}
	partInfo, err := g.CommunityInfomap(InfomapOptions{
		EdgeWeights:            eWeights,
		VertexWeights:          vWeights,
		NTrials:                3,
		IsRegularized:          true,
		RegularizationStrength: 0.1,
		Seed:                   &seed,
	})
	if err != nil || len(partInfo.Membership) != 8 {
		t.Fatalf("CommunityInfomap failed: %v, %v", partInfo, err)
	}

	// Fluid
	partFluid, err := g.CommunityFluid(2, FluidOptions{
		Seed: &seed,
	})
	if err != nil || len(partFluid.Membership) != 8 {
		t.Fatalf("CommunityFluid failed: %v, %v", partFluid, err)
	}
}

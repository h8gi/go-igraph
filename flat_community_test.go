package igraph

import (
	"math"
	"reflect"
	"testing"
)

func createZacharyKarateClub(t *testing.T) *Graph {
	t.Helper()
	// Zachary Karate Club graph (simplified benchmark: 2 cliques connected by bridge)
	edges := []Edge{
		{From: 0, To: 1}, {From: 0, To: 2}, {From: 0, To: 3}, {From: 1, To: 2}, {From: 1, To: 3}, {From: 2, To: 3},
		{From: 4, To: 5}, {From: 4, To: 6}, {From: 4, To: 7}, {From: 5, To: 6}, {From: 5, To: 7}, {From: 6, To: 7},
		{From: 3, To: 4},
	}
	g, err := NewGraphFromEdges(8, edges, false)
	if err != nil {
		t.Fatalf("failed to create graph: %v", err)
	}
	return g
}

func TestCommunityMultilevel(t *testing.T) {
	g := createZacharyKarateClub(t)
	defer g.Close()

	seed := uint64(42)
	part, err := g.CommunityMultilevel(MultilevelOptions{
		Seed: &seed,
	})
	if err != nil {
		t.Fatalf("CommunityMultilevel failed: %v", err)
	}

	if part.CommunityCount < 2 {
		t.Errorf("expected at least 2 communities, got %d", part.CommunityCount)
	}
	if len(part.Membership) != 8 {
		t.Errorf("expected membership length 8, got %d", len(part.Membership))
	}
	if math.IsNaN(part.Modularity) || part.Modularity <= 0 {
		t.Errorf("expected positive modularity, got %f", part.Modularity)
	}

	// Verify reproducibility with same seed
	part2, err := g.CommunityMultilevel(MultilevelOptions{
		Seed: &seed,
	})
	if err != nil {
		t.Fatalf("CommunityMultilevel failed on second run: %v", err)
	}
	if !reflect.DeepEqual(part.Membership, part2.Membership) {
		t.Errorf("expected reproducible membership with same seed, got %v vs %v", part.Membership, part2.Membership)
	}

	// Test with edge weights and default resolution
	eWeights := []float64{1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0}
	partW, err := g.CommunityMultilevel(MultilevelOptions{
		Weights:    eWeights,
		Resolution: -1.0,
		Seed:       &seed,
	})
	if err != nil {
		t.Fatalf("CommunityMultilevel with weights failed: %v", err)
	}
	if len(partW.Membership) != 8 {
		t.Errorf("expected membership length 8, got %d", len(partW.Membership))
	}

	// Test with custom positive resolution
	partCustom, err := g.CommunityMultilevel(MultilevelOptions{
		Resolution: 0.5,
		Seed:       &seed,
	})
	if err != nil {
		t.Fatalf("CommunityMultilevel with custom resolution failed: %v", err)
	}
	if len(partCustom.Membership) != 8 {
		t.Errorf("expected membership length 8, got %d", len(partCustom.Membership))
	}
}

func TestCommunityLeiden(t *testing.T) {
	g := createZacharyKarateClub(t)
	defer g.Close()

	seed := uint64(12345)
	part, err := g.CommunityLeiden(LeidenOptions{
		Resolution:  1.0,
		Beta:        0.01,
		NIterations: 5,
		Seed:        &seed,
	})
	if err != nil {
		t.Fatalf("CommunityLeiden failed: %v", err)
	}

	if part.CommunityCount < 2 {
		t.Errorf("expected at least 2 communities, got %d", part.CommunityCount)
	}
	if len(part.Membership) != 8 {
		t.Errorf("expected membership length 8, got %d", len(part.Membership))
	}

	// Test with edge weights
	eWeights := []float64{1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0}
	partWeighted, err := g.CommunityLeiden(LeidenOptions{
		EdgeWeights: eWeights,
		Resolution:  -1.0,
		Beta:        -1.0,
		NIterations: -1,
		Seed:        &seed,
	})
	if err != nil {
		t.Fatalf("CommunityLeiden with edge weights failed: %v", err)
	}
	if len(partWeighted.Membership) != 8 {
		t.Errorf("expected membership length 8, got %d", len(partWeighted.Membership))
	}

	// Test with custom positive beta, resolution and iterations
	partCustom, err := g.CommunityLeiden(LeidenOptions{
		Resolution:  0.8,
		Beta:        0.05,
		NIterations: 3,
		Seed:        &seed,
	})
	if err != nil {
		t.Fatalf("CommunityLeiden with custom beta failed: %v", err)
	}
	if len(partCustom.Membership) != 8 {
		t.Errorf("expected membership length 8, got %d", len(partCustom.Membership))
	}

	// Test with InitialMembership & Start: true
	initial := []int{0, 0, 0, 0, 1, 1, 1, 1}
	partStart, err := g.CommunityLeiden(LeidenOptions{
		Start:             true,
		InitialMembership: initial,
		Seed:              &seed,
	})
	if err != nil {
		t.Fatalf("CommunityLeiden with initial membership failed: %v", err)
	}
	if len(partStart.Membership) != 8 {
		t.Errorf("expected membership length 8, got %d", len(partStart.Membership))
	}

	// Test directed graph with vertex in/out weights
	dg, err := NewGraphFromEdges(2, []Edge{{From: 0, To: 1}}, true)
	if err == nil {
		defer dg.Close()
		partDG, errLeiden := dg.CommunityLeiden(LeidenOptions{
			VertexOutWeights: []float64{1.0, 0.0},
			VertexInWeights:  []float64{0.0, 1.0},
			Seed:             &seed,
		})
		if errLeiden != nil {
			t.Errorf("CommunityLeiden on directed graph with vertex weights failed: %v", errLeiden)
		}
		if len(partDG.Membership) != 2 {
			t.Errorf("expected membership length 2, got %d", len(partDG.Membership))
		}
	}
}

func TestCommunityLabelPropagation(t *testing.T) {
	g := createZacharyKarateClub(t)
	defer g.Close()

	seed := uint64(42)
	part, err := g.CommunityLabelPropagation(LabelPropagationOptions{
		Seed: &seed,
	})
	if err != nil {
		t.Fatalf("CommunityLabelPropagation failed: %v", err)
	}

	if len(part.Membership) != 8 {
		t.Errorf("expected membership length 8, got %d", len(part.Membership))
	}

	// Test with initial membership and fixed masks
	initial := []int{0, 0, 0, 0, 1, 1, 1, 1}
	fixed := []bool{true, true, true, true, true, true, true, true}
	partFixed, err := g.CommunityLabelPropagation(LabelPropagationOptions{
		InitialMembership: initial,
		Fixed:             fixed,
		Seed:              &seed,
	})
	if err != nil {
		t.Fatalf("CommunityLabelPropagation with fixed masks failed: %v", err)
	}
	if !reflect.DeepEqual(partFixed.Membership, initial) {
		t.Errorf("expected fixed membership to remain %v, got %v", initial, partFixed.Membership)
	}

	// Test with edge weights and fixed mask without initial
	eWeights := []float64{1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0}
	fixedHalf := []bool{true, true, true, true, false, false, false, false}
	partOpt, err := g.CommunityLabelPropagation(LabelPropagationOptions{
		Mode:    DirectionIn,
		Weights: eWeights,
		Fixed:   fixedHalf,
		Seed:    &seed,
	})
	if err != nil {
		t.Fatalf("CommunityLabelPropagation with options failed: %v", err)
	}
	if len(partOpt.Membership) != 8 {
		t.Errorf("expected membership length 8, got %d", len(partOpt.Membership))
	}
}

func TestCommunityInfomap(t *testing.T) {
	g := createZacharyKarateClub(t)
	defer g.Close()

	seed := uint64(42)
	part, err := g.CommunityInfomap(InfomapOptions{
		NTrials: 5,
		Seed:    &seed,
	})
	if err != nil {
		t.Fatalf("CommunityInfomap failed: %v", err)
	}

	if len(part.Membership) != 8 {
		t.Errorf("expected membership length 8, got %d", len(part.Membership))
	}
	if math.IsNaN(part.Modularity) {
		t.Errorf("expected codelength in Modularity field, got NaN")
	}
}

func TestCommunityFluid(t *testing.T) {
	g := createZacharyKarateClub(t)
	defer g.Close()

	seed := uint64(42)
	part, err := g.CommunityFluid(2, FluidOptions{
		Seed: &seed,
	})
	if err != nil {
		t.Fatalf("CommunityFluid failed: %v", err)
	}

	if part.CommunityCount != 2 {
		t.Errorf("expected 2 communities, got %d", part.CommunityCount)
	}
	if len(part.Membership) != 8 {
		t.Errorf("expected membership length 8, got %d", len(part.Membership))
	}
}

func TestFlatCommunityEdgeCasesAndValidation(t *testing.T) {
	// Empty graph checks
	emptyG, err := NewGraph()
	if err != nil {
		t.Fatalf("failed to create empty graph: %v", err)
	}
	defer emptyG.Close()

	if p, err := emptyG.CommunityMultilevel(MultilevelOptions{}); err != nil || len(p.Membership) != 0 {
		t.Errorf("unexpected Multilevel on empty graph: %v, %v", p, err)
	}
	if p, err := emptyG.CommunityLeiden(LeidenOptions{}); err != nil || len(p.Membership) != 0 {
		t.Errorf("unexpected Leiden on empty graph: %v, %v", p, err)
	}
	if p, err := emptyG.CommunityLabelPropagation(LabelPropagationOptions{}); err != nil || len(p.Membership) != 0 {
		t.Errorf("unexpected LabelPropagation on empty graph: %v, %v", p, err)
	}
	if p, err := emptyG.CommunityInfomap(InfomapOptions{}); err != nil || len(p.Membership) != 0 {
		t.Errorf("unexpected Infomap on empty graph: %v, %v", p, err)
	}
	if p, err := emptyG.CommunityFluid(2, FluidOptions{}); err != nil || len(p.Membership) != 0 {
		t.Errorf("unexpected Fluid on empty graph: %v, %v", p, err)
	}

	// Nil graph checks
	var nilG *Graph
	if _, err := nilG.CommunityMultilevel(MultilevelOptions{}); err != ErrClosed {
		t.Errorf("expected ErrClosed for nil CommunityMultilevel, got %v", err)
	}
	if _, err := nilG.CommunityLeiden(LeidenOptions{}); err != ErrClosed {
		t.Errorf("expected ErrClosed for nil CommunityLeiden, got %v", err)
	}
	if _, err := nilG.CommunityLabelPropagation(LabelPropagationOptions{}); err != ErrClosed {
		t.Errorf("expected ErrClosed for nil CommunityLabelPropagation, got %v", err)
	}
	if _, err := nilG.CommunityInfomap(InfomapOptions{}); err != ErrClosed {
		t.Errorf("expected ErrClosed for nil CommunityInfomap, got %v", err)
	}
	if _, err := nilG.CommunityFluid(2, FluidOptions{}); err != ErrClosed {
		t.Errorf("expected ErrClosed for nil CommunityFluid, got %v", err)
	}

	// Closed graph checks
	g := createZacharyKarateClub(t)
	g.Close()

	if _, err := g.CommunityMultilevel(MultilevelOptions{}); err != ErrClosed {
		t.Errorf("expected ErrClosed for CommunityMultilevel, got %v", err)
	}
	if _, err := g.CommunityLeiden(LeidenOptions{}); err != ErrClosed {
		t.Errorf("expected ErrClosed for CommunityLeiden, got %v", err)
	}
	if _, err := g.CommunityLabelPropagation(LabelPropagationOptions{}); err != ErrClosed {
		t.Errorf("expected ErrClosed for CommunityLabelPropagation, got %v", err)
	}
	if _, err := g.CommunityInfomap(InfomapOptions{}); err != ErrClosed {
		t.Errorf("expected ErrClosed for CommunityInfomap, got %v", err)
	}
	if _, err := g.CommunityFluid(2, FluidOptions{}); err != ErrClosed {
		t.Errorf("expected ErrClosed for CommunityFluid, got %v", err)
	}

	// Invalid parameters
	gActive := createZacharyKarateClub(t)
	defer gActive.Close()

	seedVal := uint64(42)
	eWeights := []float64{1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0}
	vWeights := []float64{1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0}
	if _, err := gActive.CommunityInfomap(InfomapOptions{EdgeWeights: eWeights, VertexWeights: vWeights, IsRegularized: true, RegularizationStrength: 0.1}); err != nil {
		t.Errorf("CommunityInfomap with weights and regularization failed: %v", err)
	}
	if _, err := gActive.CommunityInfomap(InfomapOptions{EdgeWeights: eWeights, Seed: &seedVal}); err != nil {
		t.Errorf("CommunityInfomap with EdgeWeights only failed: %v", err)
	}
	if _, err := gActive.CommunityInfomap(InfomapOptions{VertexWeights: vWeights, Seed: &seedVal}); err != nil {
		t.Errorf("CommunityInfomap with VertexWeights only failed: %v", err)
	}
	if _, err := gActive.CommunityInfomap(InfomapOptions{NTrials: 2, IsRegularized: false}); err != nil {
		t.Errorf("CommunityInfomap with IsRegularized false failed: %v", err)
	}
	if _, err := gActive.CommunityLeiden(LeidenOptions{Start: false, NIterations: 1}); err != nil {
		t.Errorf("CommunityLeiden with Start false failed: %v", err)
	}
	if _, err := gActive.CommunityLeiden(LeidenOptions{Start: true, NIterations: 1}); err != nil {
		t.Errorf("CommunityLeiden with Start true and empty InitialMembership failed: %v", err)
	}

	// Invalid edge weights length
	if _, err := gActive.CommunityMultilevel(MultilevelOptions{Weights: []float64{1.0}}); err == nil {
		t.Errorf("expected error for invalid edge weights length in Multilevel")
	}
	if _, err := gActive.CommunityLeiden(LeidenOptions{EdgeWeights: []float64{1.0}}); err == nil {
		t.Errorf("expected error for invalid edge weights length in Leiden")
	}

	// Invalid vertex weights length
	if _, err := gActive.CommunityLeiden(LeidenOptions{VertexOutWeights: []float64{1.0}}); err == nil {
		t.Errorf("expected error for invalid vertex out weights length")
	}
	if _, err := gActive.CommunityLeiden(LeidenOptions{VertexInWeights: []float64{1.0}}); err == nil {
		t.Errorf("expected error for invalid vertex in weights length")
	}
	if _, err := gActive.CommunityLeiden(LeidenOptions{InitialMembership: []int{0}}); err == nil {
		t.Errorf("expected error for invalid initial membership length")
	}

	// Invalid label propagation options
	if _, err := gActive.CommunityLabelPropagation(LabelPropagationOptions{Mode: NeiMode(99)}); err == nil {
		t.Errorf("expected error for invalid mode in LabelPropagation")
	}
	if _, err := gActive.CommunityLabelPropagation(LabelPropagationOptions{Weights: []float64{1.0}}); err == nil {
		t.Errorf("expected error for invalid edge weights length in LabelPropagation")
	}
	if _, err := gActive.CommunityLabelPropagation(LabelPropagationOptions{InitialMembership: []int{0}}); err == nil {
		t.Errorf("expected error for invalid initial membership length")
	}
	if _, err := gActive.CommunityLabelPropagation(LabelPropagationOptions{Fixed: []bool{true}}); err == nil {
		t.Errorf("expected error for invalid fixed mask length")
	}

	// Invalid Infomap options
	if _, err := gActive.CommunityInfomap(InfomapOptions{EdgeWeights: []float64{1.0}}); err == nil {
		t.Errorf("expected error for invalid edge weights length in Infomap")
	}
	if _, err := gActive.CommunityInfomap(InfomapOptions{VertexWeights: []float64{1.0}}); err == nil {
		t.Errorf("expected error for invalid vertex weights length in Infomap")
	}

	// Invalid Fluid community count
	if _, err := gActive.CommunityFluid(0, FluidOptions{}); err == nil {
		t.Errorf("expected error for fluid community count <= 0")
	}
	if _, err := gActive.CommunityFluid(100, FluidOptions{}); err == nil {
		t.Errorf("expected error for fluid community count > vcount")
	}
}

func TestFlatCommunityOptionsCombinations(t *testing.T) {
	g := createZacharyKarateClub(t)
	defer g.Close()

	seed := uint64(777)
	eWeights := []float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	vWeights := []float64{1, 1, 1, 1, 1, 1, 1, 1}
	initial := []int{0, 0, 0, 0, 1, 1, 1, 1}
	fixed := []bool{true, true, true, true, false, false, false, false}

	// Multilevel unweighted vs weighted
	if _, err := g.CommunityMultilevel(MultilevelOptions{Resolution: 1.0}); err != nil {
		t.Fatal(err)
	}

	// Leiden with all weights
	dg, err := NewGraphFromEdges(2, []Edge{{From: 0, To: 1}}, true)
	if err == nil {
		defer dg.Close()
		if _, err := dg.CommunityLeiden(LeidenOptions{
			EdgeWeights:      []float64{1.0},
			VertexOutWeights: []float64{1.0, 0.0},
			VertexInWeights:  []float64{0.0, 1.0},
			Resolution:       1.0,
			Beta:             0.01,
			NIterations:      2,
			Seed:             &seed,
		}); err != nil {
			t.Fatal(err)
		}
	}

	// LabelPropagation all options
	for _, mode := range []DirectionMode{DirectionOut, DirectionIn, DirectionAll} {
		if _, err := g.CommunityLabelPropagation(LabelPropagationOptions{
			Mode:              mode,
			Weights:           eWeights,
			InitialMembership: initial,
			Fixed:             fixed,
			Seed:              &seed,
		}); err != nil {
			t.Fatal(err)
		}
	}

	// Infomap all options
	if _, err := g.CommunityInfomap(InfomapOptions{
		EdgeWeights:            eWeights,
		VertexWeights:          vWeights,
		NTrials:                5,
		IsRegularized:          true,
		RegularizationStrength: 0.2,
		Seed:                   &seed,
	}); err != nil {
		t.Fatal(err)
	}
}

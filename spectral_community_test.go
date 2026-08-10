package igraph_test

import (
	"math"
	"testing"

	"github.com/h8gi/go-igraph"
)

func createBenchmarkGraph(t *testing.T) *igraph.Graph {
	t.Helper()
	edges := []igraph.Edge{
		{From: 0, To: 1}, {From: 0, To: 2}, {From: 0, To: 3}, {From: 1, To: 2}, {From: 1, To: 3}, {From: 2, To: 3},
		{From: 4, To: 5}, {From: 4, To: 6}, {From: 4, To: 7}, {From: 5, To: 6}, {From: 5, To: 7}, {From: 6, To: 7},
		{From: 3, To: 4},
	}
	g, err := igraph.NewGraphFromEdges(8, edges, false)
	if err != nil {
		t.Fatalf("failed to create benchmark graph: %v", err)
	}
	return g
}

func TestCommunityLeadingEigenvector(t *testing.T) {
	seed := uint64(42)

	t.Run("benchmark graph", func(t *testing.T) {
		g := createBenchmarkGraph(t)
		defer g.Close()

		part, err := g.CommunityLeadingEigenvector(igraph.LeadingEigenvectorOptions{
			Seed: &seed,
		})
		if err != nil {
			t.Fatalf("CommunityLeadingEigenvector failed: %v", err)
		}
		if part.CommunityCount <= 1 {
			t.Errorf("expected > 1 communities, got %d", part.CommunityCount)
		}
		if len(part.Membership) != 8 {
			t.Errorf("expected membership length 8, got %d", len(part.Membership))
		}
		if math.IsNaN(part.Modularity) || part.Modularity <= 0 {
			t.Errorf("expected positive modularity, got %f", part.Modularity)
		}
	})

	t.Run("weighted graph with steps and initial membership", func(t *testing.T) {
		g := createBenchmarkGraph(t)
		defer g.Close()

		weights := []float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0.5}
		initialMem := []int{0, 0, 0, 0, 1, 1, 1, 1}

		part, err := g.CommunityLeadingEigenvector(igraph.LeadingEigenvectorOptions{
			Weights:           weights,
			Steps:             2,
			Start:             true,
			InitialMembership: initialMem,
			Seed:              &seed,
		})
		if err != nil {
			t.Fatalf("CommunityLeadingEigenvector with options failed: %v", err)
		}
		if len(part.Membership) != 8 {
			t.Errorf("expected membership length 8, got %d", len(part.Membership))
		}
	})

	t.Run("empty graph", func(t *testing.T) {
		g, err := igraph.NewGraph()
		if err != nil {
			t.Fatalf("NewGraph failed: %v", err)
		}
		defer g.Close()

		part, err := g.CommunityLeadingEigenvector(igraph.LeadingEigenvectorOptions{
			Seed: &seed,
		})
		if err != nil {
			t.Fatalf("CommunityLeadingEigenvector failed: %v", err)
		}
		if len(part.Membership) != 0 || part.CommunityCount != 0 {
			t.Errorf("expected empty partition, got membership len=%d count=%d", len(part.Membership), part.CommunityCount)
		}
	})

	t.Run("disconnected graph", func(t *testing.T) {
		g1 := createBenchmarkGraph(t)
		defer g1.Close()

		g2, err := igraph.NewRing(5, false, false)
		if err != nil {
			t.Fatalf("Ring failed: %v", err)
		}
		defer g2.Close()

		res, err := g1.DisjointUnion(g2)
		if err != nil {
			t.Fatalf("DisjointUnion failed: %v", err)
		}
		gDis := res.Graph
		defer gDis.Close()

		part, err := gDis.CommunityLeadingEigenvector(igraph.LeadingEigenvectorOptions{
			Seed: &seed,
		})
		if err != nil {
			t.Fatalf("CommunityLeadingEigenvector on disconnected graph failed: %v", err)
		}
		if part.CommunityCount < 2 {
			t.Errorf("expected at least 2 communities for disconnected graph, got %d", part.CommunityCount)
		}
	})

	t.Run("invalid options", func(t *testing.T) {
		g, err := igraph.NewFull(4, false, false)
		if err != nil {
			t.Fatalf("Full failed: %v", err)
		}
		defer g.Close()

		_, err = g.CommunityLeadingEigenvector(igraph.LeadingEigenvectorOptions{
			Weights: []float64{1.0}, // invalid weights length
		})
		if err == nil {
			t.Error("expected error for mismatched weights length")
		}

		_, err = g.CommunityLeadingEigenvector(igraph.LeadingEigenvectorOptions{
			InitialMembership: []int{0, 1}, // mismatched initial membership
		})
		if err == nil {
			t.Error("expected error for mismatched initial membership length")
		}

		_, err = g.CommunityLeadingEigenvector(igraph.LeadingEigenvectorOptions{
			Solver: igraph.SpectralSolverOptions{MaxIterations: -5},
		})
		if err == nil {
			t.Error("expected error for invalid solver iterations")
		}
	})

	t.Run("use after close", func(t *testing.T) {
		g, err := igraph.NewFull(4, false, false)
		if err != nil {
			t.Fatalf("Full failed: %v", err)
		}
		g.Close()

		_, err = g.CommunityLeadingEigenvector(igraph.LeadingEigenvectorOptions{})
		if err == nil {
			t.Error("expected error on closed graph")
		}
	})
}

func TestCommunitySpinglass(t *testing.T) {
	t.Run("benchmark graph", func(t *testing.T) {
		g := createBenchmarkGraph(t)
		defer g.Close()

		seed := uint64(42)
		part, err := g.CommunitySpinglass(igraph.SpinglassOptions{
			Seed: &seed,
		})
		if err != nil {
			t.Fatalf("CommunitySpinglass failed: %v", err)
		}
		if part.CommunityCount <= 1 {
			t.Errorf("expected > 1 communities, got %d", part.CommunityCount)
		}
		if len(part.Membership) != 8 {
			t.Errorf("expected membership length 8, got %d", len(part.Membership))
		}

		// Reproducibility check
		part2, err := g.CommunitySpinglass(igraph.SpinglassOptions{
			Seed: &seed,
		})
		if err != nil {
			t.Fatalf("CommunitySpinglass reproducibility call failed: %v", err)
		}
		if part.CommunityCount != part2.CommunityCount {
			t.Errorf("seed reproducibility mismatch in CommunityCount: %d vs %d", part.CommunityCount, part2.CommunityCount)
		}
	})

	t.Run("weighted graph original implementation and custom options", func(t *testing.T) {
		g := createBenchmarkGraph(t)
		defer g.Close()

		weights := []float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0.5}
		seed := uint64(99)
		part, err := g.CommunitySpinglass(igraph.SpinglassOptions{
			Weights:          weights,
			Spins:            10,
			ParallelUpdate:   true,
			StartTemperature: 2.0,
			StopTemperature:  0.05,
			CoolingFactor:    0.95,
			UpdateRule:       igraph.SpincommUpdateSimple,
			Gamma:            0.8,
			Implementation:   igraph.SpinglassImplementationOriginal,
			Lambda:           0.5,
			Seed:             &seed,
		})
		if err != nil {
			t.Fatalf("CommunitySpinglass with custom options failed: %v", err)
		}
		if len(part.Membership) != 8 {
			t.Errorf("expected membership length 8, got %d", len(part.Membership))
		}
	})

	t.Run("empty graph", func(t *testing.T) {
		g, err := igraph.NewGraph()
		if err != nil {
			t.Fatalf("NewGraph failed: %v", err)
		}
		defer g.Close()

		part, err := g.CommunitySpinglass(igraph.SpinglassOptions{})
		if err != nil {
			t.Fatalf("CommunitySpinglass failed: %v", err)
		}
		if len(part.Membership) != 0 {
			t.Errorf("expected empty membership, got len=%d", len(part.Membership))
		}
	})

	t.Run("invalid parameters", func(t *testing.T) {
		g, err := igraph.NewFull(4, false, false)
		if err != nil {
			t.Fatalf("Full failed: %v", err)
		}
		defer g.Close()

		// invalid spins
		_, err = g.CommunitySpinglass(igraph.SpinglassOptions{Spins: 1000})
		if err == nil {
			t.Error("expected error for spins > 500")
		}

		// stop temp >= start temp
		_, err = g.CommunitySpinglass(igraph.SpinglassOptions{
			StartTemperature: 1.0,
			StopTemperature:  2.0,
		})
		if err == nil {
			t.Error("expected error for stopTemp >= startTemp")
		}

		// invalid cooling factor
		_, err = g.CommunitySpinglass(igraph.SpinglassOptions{
			CoolingFactor: 1.5,
		})
		if err == nil {
			t.Error("expected error for cooling factor >= 1.0")
		}

		// invalid update rule
		_, err = g.CommunitySpinglass(igraph.SpinglassOptions{
			UpdateRule: igraph.SpincommUpdateRule(99),
		})
		if err == nil {
			t.Error("expected error for invalid update rule")
		}

		// invalid implementation
		_, err = g.CommunitySpinglass(igraph.SpinglassOptions{
			Implementation: igraph.SpinglassImplementation(99),
		})
		if err == nil {
			t.Error("expected error for invalid implementation")
		}
	})

	t.Run("use after close", func(t *testing.T) {
		g, err := igraph.NewFull(4, false, false)
		if err != nil {
			t.Fatalf("Full failed: %v", err)
		}
		g.Close()

		_, err = g.CommunitySpinglass(igraph.SpinglassOptions{})
		if err == nil {
			t.Error("expected error on closed graph")
		}
	})
}

func TestCommunitySpinglassSingle(t *testing.T) {
	t.Run("benchmark single vertex default options", func(t *testing.T) {
		g := createBenchmarkGraph(t)
		defer g.Close()

		res, err := g.CommunitySpinglassSingle(igraph.SpinglassSingleOptions{
			Vertex: 0,
		})
		if err != nil {
			t.Fatalf("CommunitySpinglassSingle failed: %v", err)
		}
		if len(res.Community) == 0 {
			t.Error("expected non-empty community for vertex 0")
		}
	})

	t.Run("benchmark single vertex with weights and options", func(t *testing.T) {
		g := createBenchmarkGraph(t)
		defer g.Close()

		seed := uint64(123)
		weights := []float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0.5}
		res, err := g.CommunitySpinglassSingle(igraph.SpinglassSingleOptions{
			Weights:    weights,
			Vertex:     0,
			Spins:      15,
			UpdateRule: igraph.SpincommUpdateConfig,
			Gamma:      1.2,
			Seed:       &seed,
		})
		if err != nil {
			t.Fatalf("CommunitySpinglassSingle failed: %v", err)
		}
		if len(res.Community) == 0 {
			t.Error("expected non-empty community for vertex 0")
		}
		foundTarget := false
		for _, v := range res.Community {
			if v == 0 {
				foundTarget = true
				break
			}
		}
		if !foundTarget {
			t.Error("expected target vertex 0 to be present in community result")
		}
	})

	t.Run("out of bounds vertex and invalid params", func(t *testing.T) {
		g, err := igraph.NewFull(4, false, false)
		if err != nil {
			t.Fatalf("Full failed: %v", err)
		}
		defer g.Close()

		_, err = g.CommunitySpinglassSingle(igraph.SpinglassSingleOptions{Vertex: 10})
		if err == nil {
			t.Error("expected error for out of bounds vertex")
		}
		_, err = g.CommunitySpinglassSingle(igraph.SpinglassSingleOptions{Vertex: -1})
		if err == nil {
			t.Error("expected error for negative vertex")
		}
		_, err = g.CommunitySpinglassSingle(igraph.SpinglassSingleOptions{Vertex: 0, Spins: 600})
		if err == nil {
			t.Error("expected error for spins > 500")
		}
		_, err = g.CommunitySpinglassSingle(igraph.SpinglassSingleOptions{Vertex: 0, UpdateRule: igraph.SpincommUpdateRule(99)})
		if err == nil {
			t.Error("expected error for invalid update rule")
		}
		_, err = g.CommunitySpinglassSingle(igraph.SpinglassSingleOptions{
			Vertex:  0,
			Weights: []float64{1.0}, // invalid weights length
		})
		if err == nil {
			t.Error("expected error for mismatched weights length")
		}
	})

	t.Run("use after close", func(t *testing.T) {
		g, err := igraph.NewFull(4, false, false)
		if err != nil {
			t.Fatalf("Full failed: %v", err)
		}
		g.Close()

		_, err = g.CommunitySpinglassSingle(igraph.SpinglassSingleOptions{Vertex: 0})
		if err == nil {
			t.Error("expected error on closed graph")
		}
	})
}

func TestCommunityOptimalModularity(t *testing.T) {
	t.Run("small graph optimal modularity unweighted and weighted", func(t *testing.T) {
		g, err := igraph.NewFull(4, false, false)
		if err != nil {
			t.Fatalf("Full failed: %v", err)
		}
		defer g.Close()

		part, err := g.CommunityOptimalModularity(nil)
		if err != nil {
			t.Fatalf("CommunityOptimalModularity failed: %v", err)
		}
		if len(part.Membership) != 4 {
			t.Errorf("expected membership length 4, got %d", len(part.Membership))
		}

		weights := []float64{1, 2, 1, 2, 1, 2}
		partW, err := g.CommunityOptimalModularity(weights)
		if err != nil {
			t.Fatalf("CommunityOptimalModularity with weights failed: %v", err)
		}
		if len(partW.Membership) != 4 {
			t.Errorf("expected membership length 4, got %d", len(partW.Membership))
		}
	})

	t.Run("empty graph", func(t *testing.T) {
		g, err := igraph.NewGraph()
		if err != nil {
			t.Fatalf("NewGraph failed: %v", err)
		}
		defer g.Close()

		part, err := g.CommunityOptimalModularity(nil)
		if err != nil {
			t.Fatalf("CommunityOptimalModularity failed: %v", err)
		}
		if len(part.Membership) != 0 {
			t.Errorf("expected empty membership, got len=%d", len(part.Membership))
		}
	})

	t.Run("mismatched weights", func(t *testing.T) {
		g, err := igraph.NewFull(4, false, false)
		if err != nil {
			t.Fatalf("Full failed: %v", err)
		}
		defer g.Close()

		_, err = g.CommunityOptimalModularity([]float64{1.0})
		if err == nil {
			t.Error("expected error for mismatched weights length")
		}
	})

	t.Run("use after close", func(t *testing.T) {
		g, err := igraph.NewFull(4, false, false)
		if err != nil {
			t.Fatalf("Full failed: %v", err)
		}
		g.Close()

		_, err = g.CommunityOptimalModularity(nil)
		if err == nil {
			t.Error("expected error on closed graph")
		}
	})
}

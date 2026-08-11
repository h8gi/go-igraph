package igraph

import (
	"errors"
	"math"
	"testing"
)

func TestCycleBasesKnownAnswersAndInvariants(t *testing.T) {
	tests := []struct {
		name     string
		vertices int
		edges    []Edge
		directed bool
		rank     int
	}{
		{name: "empty", rank: 0},
		{name: "tree", vertices: 4, edges: []Edge{{0, 1}, {1, 2}, {1, 3}}, rank: 0},
		{name: "single cycle", vertices: 3, edges: []Edge{{0, 1}, {1, 2}, {2, 0}}, rank: 1},
		{name: "overlapping cycles", vertices: 4, edges: []Edge{{0, 1}, {1, 2}, {2, 0}, {1, 3}, {3, 2}}, rank: 2},
		{name: "disconnected cycles", vertices: 6, edges: []Edge{{0, 1}, {1, 2}, {2, 0}, {3, 4}, {4, 5}, {5, 3}}, rank: 2},
		{name: "loop and parallel", vertices: 2, edges: []Edge{{0, 0}, {0, 1}, {0, 1}, {0, 1}}, rank: 3},
		{name: "directed ignored", vertices: 3, edges: []Edge{{0, 1}, {2, 1}, {2, 0}}, directed: true, rank: 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := newCycleTestGraph(t, tc.vertices, tc.edges, tc.directed)
			defer g.Close()
			components, err := g.ConnectedComponents(ConnectednessWeak)
			if err != nil {
				t.Fatal(err)
			}
			cycleRank := len(tc.edges) - tc.vertices + components.Count
			if cycleRank != tc.rank {
				t.Fatalf("fixture cycle rank = %d, want %d", cycleRank, tc.rank)
			}
			fundamental, err := g.FundamentalCycleBasis(FundamentalCycleBasisOptions{})
			if err != nil {
				t.Fatal(err)
			}
			minimum, err := g.MinimumCycleBasis(MinimumCycleBasisOptions{})
			if err != nil {
				t.Fatal(err)
			}
			for name, basis := range map[string][][]int{"fundamental": fundamental, "minimum": minimum} {
				if basis == nil || len(basis) != cycleRank {
					t.Errorf("%s basis = %#v, want rank %d", name, basis, cycleRank)
					continue
				}
				assertCycleBasisEdgeIDs(t, basis, len(tc.edges))
				assertGF2Independent(t, basis, len(tc.edges))
			}
		})
	}
}

func TestFundamentalCycleBasisRootAndCutoff(t *testing.T) {
	g := newCycleTestGraph(t, 6, []Edge{
		{0, 1}, {1, 2}, {2, 0},
		{3, 4}, {4, 5}, {5, 3},
	}, false)
	defer g.Close()
	all, err := g.FundamentalCycleBasis(FundamentalCycleBasisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("all-components basis length = %d, want 2", len(all))
	}
	root := 0
	rooted, err := g.FundamentalCycleBasis(FundamentalCycleBasisOptions{Root: &root})
	if err != nil {
		t.Fatal(err)
	}
	if len(rooted) != 1 {
		t.Errorf("rooted basis length = %d, want 1", len(rooted))
	}

	cutoff0 := 0.0
	incomplete, err := g.FundamentalCycleBasis(FundamentalCycleBasisOptions{
		BFSCutoff: &cutoff0, AllowIncomplete: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if incomplete == nil || len(incomplete) != 0 {
		t.Errorf("cutoff 0 basis = %#v, want non-nil empty", incomplete)
	}
	cutoff1 := 1.0
	boundary, err := g.FundamentalCycleBasis(FundamentalCycleBasisOptions{
		BFSCutoff: &cutoff1, AllowIncomplete: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(boundary) != 2 {
		t.Errorf("cutoff 1 basis length = %d, want 2", len(boundary))
	}
}

func TestMinimumCycleBasisCutoffModes(t *testing.T) {
	g := newCycleTestGraph(t, 3, []Edge{{0, 1}, {1, 2}, {2, 0}}, false)
	defer g.Close()
	cutoff0 := 0.0
	incomplete, err := g.MinimumCycleBasis(MinimumCycleBasisOptions{
		BFSCutoff: &cutoff0, AllowIncomplete: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if incomplete == nil || len(incomplete) != 0 {
		t.Errorf("incomplete cutoff basis = %#v, want empty", incomplete)
	}
	completed, err := g.MinimumCycleBasis(MinimumCycleBasisOptions{BFSCutoff: &cutoff0})
	if err != nil {
		t.Fatal(err)
	}
	if len(completed) != 1 {
		t.Errorf("completed cutoff basis = %#v, want one cycle", completed)
	}
	cutoff1 := 1.0
	boundary, err := g.MinimumCycleBasis(MinimumCycleBasisOptions{
		BFSCutoff: &cutoff1, AllowIncomplete: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(boundary) != 1 {
		t.Errorf("cutoff boundary basis = %#v, want one cycle", boundary)
	}
}

func TestMinimumCycleBasisNaturalOrder(t *testing.T) {
	edges := []Edge{{0, 1}, {1, 2}, {2, 3}, {3, 0}, {0, 2}}
	g := newCycleTestGraph(t, 4, edges, false)
	defer g.Close()
	basis, err := g.MinimumCycleBasis(MinimumCycleBasisOptions{NaturalOrder: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(basis) != 2 {
		t.Fatalf("natural-order basis length = %d, want 2", len(basis))
	}
	for _, cycle := range basis {
		assertNaturalEdgeCycle(t, cycle, edges)
	}
}

func TestCycleBasisValidation(t *testing.T) {
	g := newCycleTestGraph(t, 1, nil, false)
	defer g.Close()
	negativeRoot, tooLargeRoot := -1, 1
	negative, nan, infinity := -1.0, math.NaN(), math.Inf(1)
	for _, options := range []FundamentalCycleBasisOptions{
		{Root: &negativeRoot},
		{Root: &tooLargeRoot},
		{BFSCutoff: &negative, AllowIncomplete: true},
		{BFSCutoff: &nan, AllowIncomplete: true},
		{BFSCutoff: &infinity, AllowIncomplete: true},
		{BFSCutoff: intToFloatPointer(1)},
	} {
		if _, err := g.FundamentalCycleBasis(options); err == nil {
			t.Errorf("FundamentalCycleBasis(%+v) succeeded", options)
		}
	}
	for _, options := range []MinimumCycleBasisOptions{
		{AllowIncomplete: true},
		{BFSCutoff: &negative},
		{BFSCutoff: &nan},
		{BFSCutoff: &infinity},
	} {
		if _, err := g.MinimumCycleBasis(options); err == nil {
			t.Errorf("MinimumCycleBasis(%+v) succeeded", options)
		}
	}
}

func TestCycleBasisResultsSurviveClose(t *testing.T) {
	g := newCycleTestGraph(t, 3, []Edge{{0, 1}, {1, 2}, {2, 0}}, false)
	fundamental, err := g.FundamentalCycleBasis(FundamentalCycleBasisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	minimum, err := g.MinimumCycleBasis(MinimumCycleBasisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	if len(fundamental) != 1 || len(minimum) != 1 {
		t.Fatalf("basis changed after Close: %v, %v", fundamental, minimum)
	}
	fundamental[0][0] = 99
	minimum[0][0] = 98
}

func TestCycleBasisUseAfterCloseAndNil(t *testing.T) {
	graphs := []*Graph{nil, newCycleTestGraph(t, 0, nil, false)}
	_ = graphs[1].Close()
	for _, g := range graphs {
		if _, err := g.FundamentalCycleBasis(FundamentalCycleBasisOptions{}); !errors.Is(err, ErrClosed) {
			t.Errorf("FundamentalCycleBasis error = %v", err)
		}
		if _, err := g.MinimumCycleBasis(MinimumCycleBasisOptions{}); !errors.Is(err, ErrClosed) {
			t.Errorf("MinimumCycleBasis error = %v", err)
		}
	}
}

func assertCycleBasisEdgeIDs(t *testing.T, basis [][]int, edgeCount int) {
	t.Helper()
	for i, cycle := range basis {
		if cycle == nil || len(cycle) == 0 {
			t.Errorf("basis cycle %d is nil or empty: %#v", i, cycle)
		}
		seen := make(map[int]struct{}, len(cycle))
		for _, edge := range cycle {
			if edge < 0 || edge >= edgeCount {
				t.Errorf("basis cycle %d edge %d out of range [0, %d)", i, edge, edgeCount)
			}
			if _, duplicate := seen[edge]; duplicate {
				t.Errorf("basis cycle %d repeats edge %d", i, edge)
			}
			seen[edge] = struct{}{}
		}
	}
}

func assertGF2Independent(t *testing.T, basis [][]int, edgeCount int) {
	t.Helper()
	if edgeCount > 64 {
		t.Fatal("test helper supports at most 64 edges")
	}
	pivots := make(map[int]uint64)
	for cycleIndex, cycle := range basis {
		var vector uint64
		for _, edge := range cycle {
			vector ^= uint64(1) << edge
		}
		for bit := edgeCount - 1; bit >= 0; bit-- {
			if vector&(uint64(1)<<bit) == 0 {
				continue
			}
			if pivot, exists := pivots[bit]; exists {
				vector ^= pivot
				continue
			}
			pivots[bit] = vector
			break
		}
		if vector == 0 {
			t.Errorf("basis cycle %d is GF(2)-dependent: %v", cycleIndex, basis)
		}
	}
}

func assertNaturalEdgeCycle(t *testing.T, cycle []int, edges []Edge) {
	t.Helper()
	if len(cycle) == 0 {
		t.Error("natural edge cycle is empty")
		return
	}
	first := edges[cycle[0]]
	for _, start := range []struct{ from, to int }{{first.From, first.To}, {first.To, first.From}} {
		current := start.to
		valid := true
		for _, edgeID := range cycle[1:] {
			edge := edges[edgeID]
			switch current {
			case edge.From:
				current = edge.To
			case edge.To:
				current = edge.From
			default:
				valid = false
			}
			if !valid {
				break
			}
		}
		if valid && current == start.from {
			return
		}
	}
	t.Errorf("edges are not in natural cycle order: %v", cycle)
}

func intToFloatPointer(value int) *float64 {
	converted := float64(value)
	return &converted
}
